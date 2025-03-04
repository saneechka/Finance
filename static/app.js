

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

      
        const response = await fetch(endpoint, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

    
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
        client_id: parseInt(document.getElementById('create_client_id').value),
        bank_name: document.getElementById('create_bank_name').value,
        amount: parseFloat(document.getElementById('create_amount').value),
        interest: parseFloat(document.getElementById('create_interest').value)
    };
    

    if (!data.client_id || !data.bank_name || !data.amount || data.interest === undefined) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/create', 'POST', data);
}


async function transferBetweenAccounts() {
    const data = {
        client_id: parseInt(document.getElementById('transfer_client_id').value),
        bank_name: document.getElementById('transfer_bank_name').value,
        from_account: parseInt(document.getElementById('from_account').value),
        to_account: parseInt(document.getElementById('to_account').value),
        amount: parseFloat(document.getElementById('transfer_amount').value),
        deposit_id: parseInt(document.getElementById('transfer_deposit_id').value)
    };
    
    // Basic validation
    if (!data.client_id || !data.bank_name || !data.from_account || 
        !data.to_account || !data.amount || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/transfer', 'POST', data);
}

// Freeze deposit
async function freezeDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('freeze_client_id').value),
        bank_name: document.getElementById('freeze_bank_name').value,
        deposit_id: parseInt(document.getElementById('freeze_deposit_id').value),
        freeze_duration: parseInt(document.getElementById('freeze_duration').value)
    };
    
    // Basic validation
    if (!data.client_id || !data.bank_name || !data.deposit_id || !data.freeze_duration) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/freeze', 'POST', data);
}

// Block deposit
async function blockDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('block_client_id').value),
        bank_name: document.getElementById('block_bank_name').value,
        deposit_id: parseInt(document.getElementById('block_deposit_id').value)
    };
    
    // Basic validation
    if (!data.client_id || !data.bank_name || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/block', 'POST', data);
}

// Unblock deposit
async function unblockDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('unblock_client_id').value),
        bank_name: document.getElementById('unblock_bank_name').value,
        deposit_id: parseInt(document.getElementById('unblock_deposit_id').value)
    };
    
    // Basic validation
    if (!data.client_id || !data.bank_name || !data.deposit_id) {
        showFeedback('error', 'Пожалуйста, заполните все поля');
        return;
    }
    
    return await makeRequest('/deposit/unblock', 'POST', data);
}

// Delete deposit
async function deleteDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('delete_client_id').value),
        bank_name: document.getElementById('delete_bank_name').value,
        deposit_id: parseInt(document.getElementById('delete_deposit_id').value)
    };
    
  
    if (!data.client_id || !data.bank_name || !data.deposit_id) {
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
        '/deposit/delete': 'deleteDeposit'
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
        '/deposit/delete': 'Депозит успешно удалён'
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
    }
    
    feedback.innerHTML = `
        <div class="feedback-icon">${icon}</div>
        <div class="feedback-text">${message}</div>
        <button class="feedback-close">&times;</button>
    `;
    
    document.body.appendChild(feedback);
    

    setTimeout(() => {
        feedback.classList.add('visible');
    }, 10);
    
 
    const closeButton = feedback.querySelector('.feedback-close');
    closeButton.addEventListener('click', () => {
        feedback.classList.add('fade-out');
        setTimeout(() => {
            if (feedback.parentNode) {
                document.body.removeChild(feedback);
            }
        }, 300);
    });
    
   
    if (type === 'success') {
        setTimeout(() => {
            if (feedback.parentNode) {
                feedback.classList.add('fade-out');
                setTimeout(() => {
                    if (feedback.parentNode) {
                        document.body.removeChild(feedback);
                    }
                }, 300);
            }
        }, 5000);
    }
}


document.addEventListener('DOMContentLoaded', () => {
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
});
