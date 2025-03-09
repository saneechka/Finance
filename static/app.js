async function makeRequest(endpoint, method, data) {
    try {
        const operationName = getOperationName(endpoint);
        const button = document.querySelector(`button[onclick="${operationName}()"]`);
        if (button) button.classList.add('processing');

        if (!navigator.onLine) {
            throw new Error('Нет подключения к интернету');
        }

        console.log(`[Request] ${method} ${endpoint}`, data);
        
        addToRequestLog(method, endpoint, data);

        const headers = {
            'Content-Type': 'application/json'
        };
        
        // Add auth token if available and not login/register request
        const token = sessionStorage.getItem('authToken');
        const isAuthEndpoint = endpoint.startsWith('/auth/login') || endpoint.startsWith('/auth/register');
        
        if (token && !isAuthEndpoint) {
            headers['Authorization'] = `Bearer ${token}`;
            
            // Check if token refresh is needed
            const tokenExpires = sessionStorage.getItem('tokenExpires');
            if (tokenExpires && !endpoint.startsWith('/auth/refresh')) {
                const expiresDate = new Date(tokenExpires);
                const now = new Date();
                
                // If token expires in less than 30 minutes, refresh it
                if ((expiresDate - now) < (30 * 60 * 1000)) {
                    const refreshed = await refreshToken();
                    if (refreshed) {
                        const newToken = sessionStorage.getItem('authToken');
                        if (newToken) {
                            headers['Authorization'] = `Bearer ${newToken}`;
                        }
                    } else {
                        // If refresh failed and this is not an auth endpoint, show message
                        if (!isAuthEndpoint) {
                            showFeedback('error', 'Сессия завершена. Пожалуйста, войдите снова.');
                            updateAuthUI(false);
                            // Abort the request
                            return { status: 401, data: { error: 'Authentication required' } };
                        }
                    }
                }
            }
        }
        
        const response = await fetch(endpoint, {
            method: method,
            headers: headers,
            body: JSON.stringify(data)
        });

        // Handle unauthorized errors (e.g., expired token)
        if (response.status === 401 && !isAuthEndpoint) {
            // Try to refresh the token
            const refreshed = await refreshToken();
            if (refreshed) {
                // Retry the original request with the new token
                return await makeRequest(endpoint, method, data);
            } else {
                // Token refresh failed, log user out
                logoutUser();
                showFeedback('error', 'Сессия завершена. Пожалуйста, войдите снова.');
                return { status: 401, data: { error: 'Authentication required' } };
            }
        }

        let responseData;
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            responseData = await response.json();
        } else {
            const text = await response.text();
            try {
                responseData = JSON.parse(text);
            } catch (e) {
                responseData = { message: text || 'No response data' };
            }
        }

        console.log(`[Response] ${response.status}`, responseData);
        
        updateResponseUI(endpoint, response.status, responseData);
        
        if (response.ok) {
            showFeedback('success', getSuccessMessage(endpoint));
        } else {
            const errorMessage = responseData.error || responseData.message || `Error: ${response.status}`;
            showFeedback('error', errorMessage);
        }

        return { status: response.status, data: responseData };
    } catch (error) {
        console.error(`[Error] ${endpoint}:`, error);
        showFeedback('error', error.message || 'Network error occurred');
        return { status: 500, data: { error: error.message } };
    } finally {
        const operationName = getOperationName(endpoint);
        const button = document.querySelector(`button[onclick="${operationName}()"]`);
        if (button) button.classList.remove('processing');
    }
}

async function createDeposit() {
    const data = {
        bank_name: document.getElementById('create_bank_name').value,
        amount: parseFloat(document.getElementById('create_amount').value),
        interest: parseFloat(document.getElementById('create_interest').value)
    };
    
    // Basic validation
    if (!data.bank_name || !data.amount || data.interest === undefined) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/create', 'POST', data);
} 

async function transferBetweenAccounts() {
    const data = {
        bank_name: document.getElementById('transfer_bank_name').value,
        from_account: parseInt(document.getElementById('from_account').value),
        to_account: parseInt(document.getElementById('to_account').value),
        amount: parseFloat(document.getElementById('transfer_amount').value),
        deposit_id: parseInt(document.getElementById('transfer_deposit_id').value)
    };
    
    // Basic validation
    if (!data.bank_name || !data.from_account || 
        !data.to_account || !data.amount || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/transfer', 'POST', data);
}

// Freeze deposit
async function freezeDeposit() {
    const data = {
        bank_name: document.getElementById('freeze_bank_name').value,
        deposit_id: parseInt(document.getElementById('freeze_deposit_id').value),
        freeze_duration: parseInt(document.getElementById('freeze_duration').value)
    };
    
    // Basic validation
    if (!data.bank_name || !data.deposit_id || !data.freeze_duration) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/freeze', 'POST', data);
}

// Block deposit
async function blockDeposit() {
    const data = {
        bank_name: document.getElementById('block_bank_name').value,
        deposit_id: parseInt(document.getElementById('block_deposit_id').value)
    };
    
    // Basic validation
    if (!data.bank_name || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/block', 'POST', data);
}

// Unblock deposit
async function unblockDeposit() {
    const data = {
        bank_name: document.getElementById('unblock_bank_name').value,
        deposit_id: parseInt(document.getElementById('unblock_deposit_id').value)
    };
    
    // Basic validation
    if (!data.bank_name || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/unblock', 'POST', data);
}

// Delete deposit
async function deleteDeposit() {
    const data = {
        bank_name: document.getElementById('delete_bank_name').value,
        deposit_id: parseInt(document.getElementById('delete_deposit_id').value)
    };
    
    // Basic validation
    if (!data.bank_name || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/delete', 'DELETE', data);
}


function getOperationName(endpoint) {
    const operations = {
        '/deposit/create': 'createDeposit',
        '/deposit/transfer': 'transferBetweenAccounts',
        '/deposit/freeze': 'freezeDeposit',
        '/deposit/block': 'blockDeposit',
        '/deposit/unblock': 'unblockDeposit',
        '/deposit/delete': 'deleteDeposit',
        '/auth/register': 'registerUser',
        '/auth/login': 'loginUser',
        '/auth/refresh': 'refreshToken'
    };
    
    return operations[endpoint] || 'unknownOperation';
}


function getSuccessMessage(endpoint) {
    const messages = {
        '/deposit/create': 'Депозит успешно создан',
        '/deposit/transfer': 'Перевод успешно выполнен',
        '/deposit/freeze': 'Депозит успешно заморожен',
        '/deposit/block': 'Депозит успешно заблокирован',
        '/deposit/unblock': 'Депозит успешно разблокирован',
        '/deposit/delete': 'Депозит успешно удалён',
        '/auth/register': 'Регистрация успешно завершена',
        '/auth/login': 'Вход выполнен успешно',
        '/auth/refresh': 'Токен успешно обновлен'
    };
    
    return messages[endpoint] || 'Операция выполнена успешно';
}


function updateResponseUI(endpoint, status, responseData) {
    const operationName = getOperationName(endpoint);
    const responseElement = document.getElementById(`${operationName.replace('Deposit', '')}_response`);
    
    if (!responseElement) return;
    
  
    let content = '';
    
    if (status >= 200 && status < 300) {
        content = `<div class="operation-status status-success">
                    <div class="status-icon"><i class="fas fa-check-circle"></i></div>
                    <div class="status-content">
                        <h4>Успешно</h4>
                        <p>${responseData.message || 'Операция выполнена успешно'}</p>
                    </div>
                </div>`;
    } else {
        content = `<div class="operation-status status-error">
                    <div class="status-icon"><i class="fas fa-exclamation-circle"></i></div>
                    <div class="status-content">
                        <h4>Ошибка</h4>
                        <p>${responseData.error || responseData.message || `Ошибка ${status}`}</p>
                    </div>
                </div>`;
    }
    
    if (typeof responseData === 'object' && Object.keys(responseData).length > 0) {
        content += `<div class="operation-details">
                    <pre>${JSON.stringify(responseData, null, 2)}</pre>
                </div>`;
    }
    
    responseElement.innerHTML = content;
    responseElement.classList.add('visible');
    
    
    const clearButton = document.createElement('button');
    clearButton.classList.add('clear-response');
    clearButton.innerHTML = 'Очистить';
    clearButton.onclick = () => {
        responseElement.innerHTML = '';
        responseElement.classList.remove('visible');
    };
    
    responseElement.appendChild(clearButton);
}


function addToRequestLog(method, endpoint, data) {
    const logEntries = document.getElementById('request-log-entries');
    if (!logEntries) return;
    
    const timestamp = new Date().toLocaleTimeString();
    const entry = document.createElement('div');
    entry.className = 'log-entry';
    
    entry.innerHTML = `
        <div class="log-header">
            <span class="log-method">${method}</span>
            <span class="log-endpoint">${endpoint}</span>
            <span class="log-time">${timestamp}</span>
        </div>
        <pre class="log-data">${JSON.stringify(data, null, 2)}</pre>
    `;
    
    logEntries.appendChild(entry);
    
 
 
    logEntries.scrollTop = logEntries.scrollHeight
}

function showFeedback(type, message) {
    const existingFeedback = document.querySelector('.feedback-message');
    if (existingFeedback) {
        document.body.removeChild(existingFeedback);
    }
    
    const feedback = document.createElement('div');
    feedback.className = `feedback-message ${type}`;
    
    let icon = '';
    switch(type) {
        case 'success':
            icon = '<i class="fas fa-check-circle"></i>';
            break;
        case 'error':
            icon = '<i class="fas fa-exclamation-circle"></i>';
            break;
        case 'processing':
            icon = '<i class="fas fa-circle-notch fa-spin"></i>';
            break;
        case 'info':
            icon = '<i class="fas fa-info-circle"></i>';
            break;
        case 'warning':
            icon = '<i class="fas fa-exclamation-triangle"></i>';
            break;
    }
    
    feedback.innerHTML = `
        <div class="feedback-icon">${icon}</div>
        <div class="feedback-text">${message}</div>
        <button class="feedback-close">&times;</button>
    `;
    
    // Apply modern styling
    Object.assign(feedback.style, {
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        padding: '16px',
        borderRadius: '8px',
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
        display: 'flex',
        alignItems: 'center',
        maxWidth: '350px',
        zIndex: '9999',
        opacity: '0',
        transform: 'translateY(20px)',
        transition: 'opacity 0.3s ease, transform 0.3s ease'
    });
    
    // Style based on message type
    switch(type) {
        case 'success':
            Object.assign(feedback.style, {
                background: '#f0fdfa',
                borderLeft: '4px solid #10b981'
            });
            break;
        case 'error':
            Object.assign(feedback.style, {
                background: '#fef2f2',
                borderLeft: '4px solid #ef4444'
            });
            break;
        case 'info':
            Object.assign(feedback.style, {
                background: '#eff6ff',
                borderLeft: '4px solid #3b82f6'
            });
            break;
        case 'warning':
            Object.assign(feedback.style, {
                background: '#fffbeb',
                borderLeft: '4px solid #f59e0b'
            });
            break;
    }
    
    // Style the icon
    const iconDiv = feedback.querySelector('.feedback-icon');
    Object.assign(iconDiv.style, {
        marginRight: '12px',
        fontSize: '24px'
    });
    
    // Style close button
    const closeButton = feedback.querySelector('.feedback-close');
    Object.assign(closeButton.style, {
        background: 'none',
        border: 'none',
        fontSize: '20px',
        marginLeft: '8px',
        cursor: 'pointer',
        opacity: '0.5',
        transition: 'opacity 0.2s'
    });
    
    closeButton.addEventListener('mouseover', () => {
        closeButton.style.opacity = '1';
    });
    
    closeButton.addEventListener('mouseout', () => {
        closeButton.style.opacity = '0.5';
    });
    
    document.body.appendChild(feedback);
    
    // Animate in
    setTimeout(() => {
        feedback.style.opacity = '1';
        feedback.style.transform = 'translateY(0)';
    }, 10);
    
    // Handle close button click
    closeButton.addEventListener('click', () => {
        feedback.style.opacity = '0';
        feedback.style.transform = 'translateY(20px)';
        setTimeout(() => {
            if (feedback.parentNode) {
                document.body.removeChild(feedback);
            }
        }, 300);
    });
    
    // Auto dismiss success messages
    if (type === 'success') {
        setTimeout(() => {
            if (feedback.parentNode) {
                feedback.style.opacity = '0';
                feedback.style.transform = 'translateY(20px)';
                setTimeout(() => {
                    if (feedback.parentNode) {
                        document.body.removeChild(feedback);
                    }
                }, 300);
            }
        }, 5000);
    }
}


document.addEventListener('DOMContentLoaded', async () => {
    const style = document.createElement('style');
    style.textContent = `
        .log-entry {
            margin-bottom: 10px;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            overflow: hidden;
        }
        
        .log-header {
            display: flex;
            padding: 8px 12px;
            background: #f1f5f9;
            font-size: 0.9rem;
        }
        
        .log-method {
            font-weight: bold;
            margin-right: 10px;
        }
        
        .log-endpoint {
            color: var(--primary-color);
            flex-grow: 1;
        }
        
        .log-time {
            color: var(--text-secondary);
            font-size: 0.8rem;
        }
        
        .log-data {
            margin: 0;
            padding: 10px;
            background: #fafafa;
            font-size: 0.9rem;
            overflow-x: auto;
        }
    `;
    document.head.appendChild(style);
    
    // Check if user is logged in
    const token = sessionStorage.getItem('authToken');
    if (token) {
        try {
            // Verify token validity
            const tokenExpires = sessionStorage.getItem('tokenExpires');
            const expiresDate = tokenExpires ? new Date(tokenExpires) : null;
            const now = new Date();
            
            if (expiresDate && expiresDate > now) {
                // Token still valid
                updateAuthUI(true);
            } else if (expiresDate) {
                // Token expired, try to refresh
                const refreshed = await refreshToken();
                updateAuthUI(refreshed);
                if (!refreshed) {
                    showFeedback('error', 'Сессия истекла. Пожалуйста, войдите снова.');
                }
            } else {
                // No expiration info, try to validate token
                const response = await fetch('/auth/refresh', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({})
                });
                
                if (response.ok) {
                    const data = await response.json();
                    sessionStorage.setItem('authToken', data.token);
                    sessionStorage.setItem('tokenExpires', data.expires);
                    updateAuthUI(true);
                } else {
                    // Token invalid
                    logoutUser();
                }
            }
        } catch (error) {
            console.error('Error validating token:', error);
            logoutUser();
        }
    } else {
        updateAuthUI(false);
    }
});

// Authentication functions
async function registerUser() {
    const data = {
        username: document.getElementById('register_username').value,
        password: document.getElementById('register_password').value,
        email: document.getElementById('register_email').value,
        role: document.getElementById('register_role') ? document.getElementById('register_role').value : 'client'
    };
    
    // Basic validation
    if (!data.username || !data.password) {
        showFeedback('error', 'Пожалуйста, заполните обязательные поля');
        return;
    }
    
    try {
        const response = await fetch('/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        
        const result = await response.json();
        
        if (response.ok) {
            // Different behavior based on role
            if (data.role === 'admin' || data.role === 'operator') {
                // Admin and operator accounts are auto-approved
                showFeedback('success', `Регистрация ${data.role === 'admin' ? 'администратора' : 'оператора'} успешна. Вы можете войти в систему.`);
                
                // Navigate user to login form
                if (document.querySelector('.tab-button[data-tab="login-tab"]')) {
                    document.querySelector('.tab-button[data-tab="login-tab"]').click();
                    document.getElementById('login_username').value = data.username;
                } else {
                    setTimeout(() => {
                        window.location.href = '/auth';
                    }, 2000);
                }
            } else {
                // Client accounts need approval - store with username to prevent conflicts
                localStorage.setItem('pendingUser_' + data.username, data.username);
                localStorage.setItem('pendingApproval_' + data.username, 'true');
                
                showFeedback('success', 'Регистрация успешна! Ваша учетная запись ожидает подтверждения администратором.');
                
                // If we're on auth page, show pending approval tab
                if (document.getElementById('pending-approval')) {
                    showPendingApprovalTab(data.username);
                } else {
                    setTimeout(() => {
                        window.location.href = '/auth';
                    }, 2000);
                }
            }
        } else {
            showFeedback('error', result.error || 'Ошибка регистрации');
        }
        
        return { status: response.status, data: result };
    } catch (error) {
        showFeedback('error', 'Ошибка сети при регистрации');
        return { status: 500, data: { error: error.message } };
    }
}

async function loginUser() {
    const data = {
        username: document.getElementById('login_username').value,
        password: document.getElementById('login_password').value
    };
    
    // Basic validation
    if (!data.username || !data.password) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    try {
        const response = await fetch('/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        
        const responseData = await response.json();
        const result = { status: response.status, data: responseData };
        
        if (response.ok && responseData.token) {
            // Save token in a way that allows multiple users in different tabs
            // Use session storage for tab-specific auth
            sessionStorage.setItem('authToken', responseData.token);
            sessionStorage.setItem('userID', responseData.user_id);
            sessionStorage.setItem('username', responseData.username);
            sessionStorage.setItem('userRole', responseData.role);
            sessionStorage.setItem('tokenExpires', responseData.expires);
            
            // Clear pending approval status if it exists
            localStorage.removeItem('pendingApproval_' + data.username);
            localStorage.removeItem('pendingUser_' + data.username);
            
            // Update UI for logged in state
            updateAuthUI(true);
            
            showFeedback('success', 'Вы успешно вошли в систему');
            
            // Redirect to appropriate page based on role
            setTimeout(() => {
                if (responseData.role === 'admin') {
                    window.location.href = '/admin';
                } else if (responseData.role === 'operator') {
                    window.location.href = '/operator';
                } else {
                    window.location.href = '/';
                }
            }, 1000);
        } else {
            if (response.status === 403 && responseData.error && responseData.error.includes('pending approval')) {
                localStorage.setItem('pendingUser_' + data.username, data.username);
                localStorage.setItem('pendingApproval_' + data.username, 'true');
                
                if (document.getElementById('pending-approval')) {
                    showPendingApprovalTab(data.username);
                }
            }
            
            const errorMessage = responseData.error || 'Ошибка авторизации';
            showFeedback('error', errorMessage);
        }
        
        return result;
    } catch (error) {
        console.error('Login error:', error);
        showFeedback('error', 'Ошибка при входе в систему');
        return { status: 500, data: { error: error.message } };
    }
}

function logoutUser() {
    // Clear session-specific storage instead of localStorage
    sessionStorage.removeItem('authToken');
    sessionStorage.removeItem('userID');
    sessionStorage.removeItem('username');
    sessionStorage.removeItem('userRole');
    sessionStorage.removeItem('tokenExpires');
    
    // Update UI for logged out state
    updateAuthUI(false);
    
    showFeedback('success', 'Вы вышли из системы');
    
    // Redirect to login page
    setTimeout(() => {
        window.location.href = '/auth';
    }, 1000);
}

async function refreshToken() {
    const token = sessionStorage.getItem('authToken');
    if (!token) return false;
    
    try {
        const response = await fetch('/auth/refresh', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Token refresh failed');
        }
        
        const data = await response.json();
        
        if (data.token) {
            sessionStorage.setItem('authToken', data.token);
            sessionStorage.setItem('tokenExpires', data.expires);
            return true;
        } else {
            throw new Error('Invalid refresh token response');
        }
    } catch (error) {
        console.error('Error refreshing token:', error);
        // Clear invalid token data
        logoutUser();
        return false;
    }
}

function updateAuthUI(isLoggedIn) {
    const authForms = document.getElementById('auth-forms');
    const appContent = document.getElementById('app-content');
    const userInfo = document.getElementById('user-info');
    
    if (isLoggedIn) {
        if (authForms) authForms.style.display = 'none';
        if (appContent) appContent.style.display = 'block';
        if (userInfo) {
            const userID = sessionStorage.getItem('userID');
            const username = sessionStorage.getItem('username') || userID;
            const role = sessionStorage.getItem('userRole') || 'client';
            
            userInfo.innerHTML = `
                <div class="user-info-details">
                    <span class="username">${username}</span>
                    <span class="user-role ${role}">${role}</span>
                </div>
                <button onclick="logoutUser()">Logout</button>
            `;
            userInfo.style.display = 'flex';
            
            // Show admin panel if user is an admin
            if (role === 'admin') {
                const adminElements = document.querySelectorAll('.admin-only');
                if (adminElements) {
                    adminElements.forEach(el => el.style.display = 'block');
                }
            }
        }
    } else {
        if (authForms) authForms.style.display = 'block';
        if (appContent) appContent.style.display = 'none';
        if (userInfo) userInfo.style.display = 'none';
    }
}

// Modified functions to handle username-specific pending status
function showPendingApprovalTab(username) {
    const pendingUsername = username || document.getElementById('login_username').value;
    document.getElementById('pending-username').textContent = pendingUsername;
    
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(content => content.style.display = 'none');
    
    document.getElementById('pending-approval').style.display = 'block';
    
    const tabButtons = document.querySelectorAll('.tab-button');
    tabButtons.forEach(btn => btn.classList.remove('active'));
}

async function checkApprovalStatus() {
    const usernameElement = document.getElementById('pending-username');
    const username = usernameElement ? usernameElement.textContent : '';
    
    if (!username) {
        showFeedback('error', 'No pending registration found');
        return;
    }
    
    // ...existing code...
}

// Loan Management Functions
async function loadLoans() {
    const container = document.getElementById('loans-list');
    if (!container) return;

    try {
        const response = await fetch('/loan/list', {
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken')
            }
        });

        if (!response.ok) throw new Error('Failed to load loans');

        const result = await response.json();
        // Fix: Access the loans array from the response data
        const loans = result.loans || [];

        if (loans.length === 0) {
            container.innerHTML = `
                <div class="no-data">
                    <i class="fas fa-file-invoice-dollar"></i>
                    <p>У вас пока нет кредитов</p>
                    <button class="primary-button" onclick="showCreateLoanModal()">
                        <i class="fas fa-plus"></i> Оформить кредит или рассрочку
                    </button>
                </div>
            `;
            return;
        }

        let html = '<div class="cards-grid">';
        loans.forEach(loan => {
            const progress = loan.status === 'Active' ? 
                (loan.paid_amount / loan.total_payable) * 100 : 0;

            html += `
                <div class="loan-card">
                    <div class="loan-header">
                        <h3 class="loan-type">
                            ${loan.type === 'standard' ? 'Кредит' : 'Рассрочка'}
                        </h3>
                        <span class="loan-status ${loan.status.toLowerCase()}">
                            ${translateStatus(loan.status)}
                        </span>
                    </div>
                    <div class="loan-content">
                        <div class="loan-amount">₸${loan.amount.toFixed(2)}</div>
                        <div class="loan-details">
                            <div class="loan-detail-item">
                                <span class="detail-label">Срок</span>
                                <span class="detail-value">${loan.term_months} месяцев</span>
                            </div>
                            <div class="loan-detail-item">
                                <span class="detail-label">Процентная ставка</span>
                                <span class="detail-value">${loan.interest_rate}%</span>
                            </div>
                            <div class="loan-detail-item">
                                <span class="detail-label">Ежемесячный платеж</span>
                                <span class="detail-value">₸${loan.monthly_payment.toFixed(2)}</span>
                            </div>
                            <div class="loan-detail-item">
                                <span class="detail-label">Общая сумма к оплате</span>
                                <span class="detail-value">₸${loan.total_payable.toFixed(2)}</span>
                            </div>
                        </div>
                        ${loan.status === 'Active' ? `
                            <div class="loan-progress">
                                <div class="progress-header">
                                    <span class="progress-label">Прогресс погашения</span>
                                    <span class="progress-value">${progress.toFixed(1)}%</span>
                                </div>
                                <div class="progress-bar">
                                    <div class="progress-fill" style="width: ${progress}%"></div>
                                </div>
                            </div>
                        ` : ''}
                    </div>
                    ${loan.status === 'Active' ? `
                        <div class="loan-actions">
                            <button class="action-btn primary" onclick="showPaymentModal(${loan.id})">
                                Внести платеж
                            </button>
                            <button class="action-btn secondary" onclick="showLoanDetails(${loan.id})">
                                Детали кредита
                            </button>
                        </div>
                    ` : ''}
                </div>
            `;
        });
        html += '</div>';
        container.innerHTML = html;
    } catch (error) {
        container.innerHTML = `
            <div class="error">
                <i class="fas fa-exclamation-circle"></i>
                <p>Ошибка загрузки кредитов: ${error.message}</p>
            </div>
        `;
    }
}

// Add helper function to translate loan statuses
function translateStatus(status) {
    const translations = {
        'pending': 'На рассмотрении',
        'approved': 'Одобрен',
        'active': 'Активен',
        'completed': 'Погашен',
        'rejected': 'Отклонен',
        'default': 'Просрочен'
    };
    return translations[status.toLowerCase()] || status;
}

async function createLoan(event) {
    event.preventDefault();
    const form = document.getElementById('create-loan-form');
    const data = {
        type: form.querySelector('#loan_type').value,
        amount: parseFloat(form.querySelector('#loan_amount').value),
        term_months: parseInt(form.querySelector('#loan_term').value)
    };

    const customRate = form.querySelector('#custom_rate').value;
    if (customRate) {
        data.interest_rate = parseFloat(customRate);
    }

    try {
        const response = await fetch('/loan/request', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            document.getElementById('create-loan-modal').style.display = 'none';
            form.reset();
            showFeedback('success', 'Loan request submitted successfully');
            loadLoans();
        } else {
            showFeedback('error', result.error || 'Failed to submit loan request');
        }
    } catch (error) {
        showFeedback('error', error.message);
    }
}

async function makeLoanPayment(loanId, amount) {
    try {
        const response = await fetch('/loan/payment', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                loan_id: loanId,
                amount: amount
            })
        });

        const result = await response.json();

        if (response.ok) {
            showFeedback('success', 'Payment processed successfully');
            loadLoans();
            return true;
        } else {
            showFeedback('error', result.error || 'Failed to process payment');
            return false;
        }
    } catch (error) {
        showFeedback('error', error.message);
        return false;
    }
}

function showPaymentModal(loanId) {
    // Create modal if it doesn't exist
    let modal = document.getElementById('payment-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'payment-modal';
        modal.className = 'modal';
        modal.innerHTML = `
            <div class="modal-content">
                <span class="close">&times;</span>
                <h3>Make Payment</h3>
                <form id="payment-form">
                    <input type="hidden" id="payment-loan-id">
                    <div class="form-group">
                        <label for="payment-amount">Payment Amount</label>
                        <input type="number" id="payment-amount" min="0" step="0.01" required>
                    </div>
                    <button type="submit" class="primary-button">Submit Payment</button>
                </form>
            </div>
        `;
        document.body.appendChild(modal);

        // Set up event listeners
        modal.querySelector('.close').onclick = () => modal.style.display = 'none';
        modal.querySelector('#payment-form').onsubmit = async (e) => {
            e.preventDefault();
            const amount = parseFloat(document.getElementById('payment-amount').value);
            const loanId = parseInt(document.getElementById('payment-loan-id').value);
            if (await makeLoanPayment(loanId, amount)) {
                modal.style.display = 'none';
            }
        };
    }

    // Show modal with loan ID
    document.getElementById('payment-loan-id').value = loanId;
    document.getElementById('payment-amount').value = '';
    modal.style.display = 'block';
}

function updateLoanTermDisplay() {
    const termSelect = document.getElementById('loan_term');
    const rates = {
        '3': 5.0,
        '6': 7.5,
        '12': 10.0,
        '24': 15.0,
        '36': 20.0
    };
    
    const term = termSelect.value;
    const options = termSelect.options;
    for (let i = 0; i < options.length; i++) {
        if (options[i].value === term) {
            options[i].text = `${term} Months (${rates[term]}% Interest)`;
        }
    }
}

// Initialize loan functionality
document.addEventListener('DOMContentLoaded', function() {
    // ...existing initialization code...

    // Set up loan term change handler
    const loanTerm = document.getElementById('loan_term');
    if (loanTerm) {
        loanTerm.addEventListener('change', updateLoanTermDisplay);
    }

    // Set up loan form submission
    const loanForm = document.getElementById('create-loan-form');
    if (loanForm) {
        loanForm.addEventListener('submit', createLoan);
    }

    // Load loans if we're on the main page
    if (document.getElementById('loans-list')) {
        loadLoans();
    }
});

async function loadDeposits() {
    const container = document.getElementById('deposits-list');
    if (!container) return;
    
    try {
        container.innerHTML = `
            <div class="loading">
                <i class="fas fa-spinner fa-spin"></i>
                <p>Загрузка депозитов...</p>
            </div>
        `;

        const response = await fetch('/deposit/list', {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(await response.text() || 'Failed to load deposits');
        }
        
        const result = await response.json();
        
        if (!result.data || !result.data.deposits) {
            throw new Error('Invalid response format');
        }

        const deposits = result.data.deposits;

        if (deposits.length === 0) {
            container.innerHTML = `
                <div class="no-data">
                    <i class="fas fa-piggy-bank"></i>
                    <p>У вас пока нет депозитов</p>
                    <button class="primary-button" onclick="showCreateDepositModal()">
                        <i class="fas fa-plus"></i> Открыть новый депозит
                    </button>
                </div>
            `;
            return;
        }

        renderDeposits(deposits);
    } catch (error) {
        console.error('Error loading deposits:', error);
        container.innerHTML = `
            <div class="error">
                <i class="fas fa-exclamation-circle"></i>
                <p>Ошибка загрузки депозитов: ${error.message}</p>
                <button class="secondary-button" onclick="loadDeposits()">
                    <i class="fas fa-sync"></i> Попробовать снова
                </button>
            </div>
        `;
    }
}

function renderDeposits(deposits) {
    const container = document.getElementById('deposits-list');
    if (!Array.isArray(deposits)) {
        console.error('Deposits data is not an array:', deposits);
        container.innerHTML = '<div class="error">Некорректный формат данных</div>';
        return;
    }

    let html = '<div class="cards-grid">';
    deposits.forEach(deposit => {
        html += `
            <div class="finance-card">
                <div class="card-header">
                    <h3>${deposit.bank_name || 'Банк'}</h3>
                    <span class="status ${deposit.is_blocked ? 'blocked' : deposit.is_frozen ? 'frozen' : 'active'}">
                        ${deposit.is_blocked ? 'Заблокирован' : deposit.is_frozen ? 'Заморожен' : 'Активен'}
                    </span>
                </div>
                <div class="card-body">
                    <div class="amount">₸${(deposit.amount || 0).toFixed(2)}</div>
                    <div class="details">
                        <p>Процентная ставка: ${deposit.interest || 0}%</p>
                        <p>ID: ${deposit.deposit_id || 'N/A'}</p>
                    </div>
                </div>
                <div class="card-actions">
                    <button class="action-btn transfer" onclick="showTransferModal(${deposit.deposit_id})">
                        <i class="fas fa-exchange-alt"></i> Перевод
                    </button>
                    ${!deposit.is_blocked ? 
                        `<button class="action-btn ${deposit.is_frozen ? 'unfreeze' : 'freeze'}" 
                            onclick="${deposit.is_frozen ? 'unfreezeDeposit' : 'showFreezeModal'}(${deposit.deposit_id})">
                            <i class="fas fa-${deposit.is_frozen ? 'sun' : 'snowflake'}"></i>
                            ${deposit.is_frozen ? 'Разморозить' : 'Заморозить'}
                        </button>` : ''}
                    ${!deposit.is_frozen ? 
                        `<button class="action-btn ${deposit.is_blocked ? 'unblock' : 'block'}"
                            onclick="${deposit.is_blocked ? 'unblockDeposit' : 'blockDeposit'}(${deposit.deposit_id})">
                            <i class="fas fa-${deposit.is_blocked ? 'unlock' : 'lock'}"></i>
                            ${deposit.is_blocked ? 'Разблокировать' : 'Блокировать'}
                        </button>` : ''}
                </div>
            </div>
        `;
    });
    html += '</div>';
    container.innerHTML = html;
}

// Set up a function to show create deposit modal
function showCreateDepositModal() {
    // Get the modal element
    const modal = document.getElementById('create-deposit-modal');
    if (!modal) {
        // Create modal if it doesn't exist yet
        createDepositModal();
    }
    
    // Show the modal
    document.getElementById('create-deposit-modal').style.display = 'block';
}

function createDepositModal() {
    const modal = document.createElement('div');
    modal.id = 'create-deposit-modal';
    modal.className = 'modal';
    
    modal.innerHTML = `
        <div class="modal-content">
            <span class="close" onclick="document.getElementById('create-deposit-modal').style.display='none'">&times;</span>
            <h3>Создать новый депозит</h3>
            <form id="create-deposit-form">
                <div class="form-group">
                    <label for="create_bank_name">Название банка</label>
                    <input type="text" id="create_bank_name" required>
                </div>
                <div class="form-group">
                    <label for="create_amount">Сумма</label>
                    <input type="number" id="create_amount" min="0" step="0.01" required>
                </div>
                <div class="form-group">
                    <label for="create_interest">Процентная ставка (%)</label>
                    <input type="number" id="create_interest" min="0" max="100" step="0.1" required>
                </div>
                <button type="button" onclick="createDeposit()" class="primary-button">Создать депозит</button>
            </form>
        </div>
    `;
    
    document.body.appendChild(modal);
}

document.addEventListener('DOMContentLoaded', function() {
    // ...existing initialization code...
    
    // Initial data load
    if (document.getElementById('deposits-list')) {
        loadDeposits();
    }
    if (document.getElementById('loans-list')) {
        loadLoans();
    }
    
    // Refresh data every 30 seconds
    setInterval(() => {
        if (document.getElementById('deposits-list')) {
            loadDeposits();
        }
        if (document.getElementById('loans-list')) {
            loadLoans();
        }
    }, 30000);
});
