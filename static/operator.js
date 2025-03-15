document.addEventListener('DOMContentLoaded', function() {
    // Check authorization
    const authToken = sessionStorage.getItem('authToken');
    const userRole = sessionStorage.getItem('userRole');
    
    if (authToken && (userRole === 'operator' || userRole === 'admin' || userRole === 'manager')) {
        document.getElementById('app-content').style.display = 'block';
        document.getElementById('auth-check').style.display = 'none';
        

        const userID = sessionStorage.getItem('userID');
        const username = sessionStorage.getItem('username') || userID;
        
        const userInfo = document.getElementById('user-info');
        userInfo.innerHTML = `
            <div class="user-info-details">
                <span class="username">${username}</span>
                <span class="user-role operator">${userRole === 'admin' ? 'Администратор' : userRole === 'manager' ? 'Менеджер' : 'Оператор'}</span>
            </div>
            <button onclick="logoutUser()">Выход</button>
        `;
        userInfo.style.display = 'flex';
        

        loadUsers();
        loadStatistics();
        loadRecentActions();


        document.getElementById('search-user-btn').addEventListener('click', function() {
            loadUsers();
        });
        

        document.getElementById('refresh-stats').addEventListener('click', function() {
            loadStatistics();
        });
        

        document.getElementById('stats-period').addEventListener('change', function() {
            loadStatistics();
        });
        

        document.getElementById('refresh-actions').addEventListener('click', function() {
            loadRecentActions();
        });
        

        document.getElementById('actions-limit').addEventListener('change', function() {
            loadRecentActions();
        });
        

        document.querySelector('#cancel-action-modal .close-modal').addEventListener('click', function() {
            document.getElementById('cancel-action-modal').style.display = 'none';
        });
        
        document.getElementById('cancel-modal-close').addEventListener('click', function() {
            document.getElementById('cancel-action-modal').style.display = 'none';
        });
        
        document.getElementById('confirm-cancel-action').addEventListener('click', function() {
            const actionId = this.getAttribute('data-action-id');
            const userId = this.getAttribute('data-user-id');
            cancelLastUserAction(userId, actionId);
        });
    } else {
        document.getElementById('app-content').style.display = 'none';
        document.getElementById('auth-check').style.display = 'block';
    }
});

async function loadUsers() {
    const tbody = document.getElementById('users-list');
    tbody.innerHTML = '<tr><td colspan="6" class="loading-row">Загрузка пользователей...</td></tr>';
    
    try {

        const searchTerm = document.getElementById('search-user').value;
        

        let queryParams = '';
        if (searchTerm) {
            queryParams = `?search=${encodeURIComponent(searchTerm)}`;
        }
        
        const response = await fetch('/operator/users' + queryParams, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Не удалось загрузить список пользователей');
        }
        
        const result = await response.json();
        
        if (!result.users || result.users.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="empty-row">Пользователи не найдены</td></tr>';
            return;
        }
        
        let tableHTML = '';
        
        result.users.forEach(user => {
            const lastAction = user.last_action ? {
                id: user.last_action.id || '-',
                type: user.last_action.type || '-',
                date: user.last_action.timestamp ? new Date(user.last_action.timestamp).toLocaleString() : '-',
                amount: user.last_action.amount ? formatCurrency(user.last_action.amount) : '-'
            } : null;
            

            const canCancelAction = user.has_cancellable_action === true;
            
            tableHTML += `
                <tr>
                    <td>${user.id}</td>
                    <td>${user.username}</td>
                    <td>${user.email || '-'}</td>
                    <td><span class="badge role-${user.role}">${user.role}</span></td>
                    <td>${lastAction ? `${lastAction.type} (${lastAction.date})` : 'Нет действий'}</td>
                    <td>
                        ${canCancelAction ? 
                            `<button class="action-btn danger small" onclick="showCancelActionModal(${user.id})">
                                <i class="fas fa-ban"></i> Отменить последнее
                            </button>` : 
                            `<span class="action-disabled" title="Нет действий, которые можно отменить">Нет действий для отмены</span>`
                        }
                    </td>
                </tr>
            `;
        });
        
        tbody.innerHTML = tableHTML;
    } catch (error) {
        console.error('Error loading users:', error);
        tbody.innerHTML = `<tr><td colspan="6" class="error-row">${error.message}</td></tr>`;
        showNotification('error', 'Не удалось загрузить список пользователей: ' + error.message);
    }
}

async function loadStatistics() {
    try {

        document.getElementById('total-transactions').innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        document.getElementById('total-amount').innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        document.getElementById('active-users').innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        document.getElementById('avg-transaction').innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        

        const period = document.getElementById('stats-period').value;
        

        const response = await fetch(`/operator/statistics?period=${period}`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Не удалось загрузить статистику операций');
        }
        
        const result = await response.json();
        
        if (!result.statistics) {
            throw new Error('Некорректный формат данных статистики');
        }
        

        const stats = result.statistics;
        document.getElementById('total-transactions').textContent = stats.total_transactions.toLocaleString();
        document.getElementById('total-amount').textContent = formatCurrency(stats.total_amount);
        document.getElementById('active-users').textContent = stats.active_users.toLocaleString();
        document.getElementById('avg-transaction').textContent = formatCurrency(stats.avg_transaction);
        

        if (result.by_type) {
            displayTransactionTypes(result.by_type);
        }
        
    } catch (error) {
        console.error('Error loading statistics:', error);
        

        document.getElementById('total-transactions').textContent = '—';
        document.getElementById('total-amount').textContent = '—';
        document.getElementById('active-users').textContent = '—';
        document.getElementById('avg-transaction').textContent = '—';
        
        showNotification('error', 'Не удалось загрузить статистику: ' + error.message);
    }
}

async function loadRecentActions() {
    const tbody = document.getElementById('recent-actions-list');
    tbody.innerHTML = '<tr><td colspan="6" class="loading-row">Загрузка действий...</td></tr>';
    
    try {

        const limit = document.getElementById('actions-limit').value;
        

        const response = await fetch(`/operator/recent-actions?limit=${limit}`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Не удалось загрузить список действий');
        }
        
        const result = await response.json();
        
        if (!result.recent_actions || result.recent_actions.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="empty-row">Действия не найдены</td></tr>';
            return;
        }
        
        let tableHTML = '';
        
        result.recent_actions.forEach(action => {

            const actionDate = new Date(action.timestamp).toLocaleString();
            const actionAmount = action.amount ? formatCurrency(action.amount) : '-';
            const canCancel = action.can_cancel;
            
            tableHTML += `
                <tr>
                    <td>${action.id}</td>
                    <td>${action.username || `Пользователь #${action.user_id}`}</td>
                    <td>${translateTransactionType(action.type)}</td>
                    <td>${actionAmount}</td>
                    <td>${actionDate}</td>
                    <td>
                        ${canCancel ? 
                            `<button class="action-btn danger small" onclick="showCancelActionModal(${action.user_id})">
                                <i class="fas fa-ban"></i> Отменить
                            </button>` : 
                            `<span class="action-disabled" title="Действие не может быть отменено">Нельзя отменить</span>`
                        }
                    </td>
                </tr>
            `;
        });
        
        tbody.innerHTML = tableHTML;
    } catch (error) {
        console.error('Error loading recent actions:', error);
        tbody.innerHTML = `<tr><td colspan="6" class="error-row">${error.message}</td></tr>`;
        showNotification('error', 'Не удалось загрузить список действий: ' + error.message);
    }
}

function displayTransactionTypes(typeCounts) {
    const container = document.getElementById('transaction-types');
    
    if (!typeCounts || Object.keys(typeCounts).length === 0) {
        container.innerHTML = '<div class="no-data">Нет данных для отображения</div>';
        return;
    }
    

    const typeArray = Object.entries(typeCounts).sort((a, b) => b[1] - a[1]);
    

    const total = typeArray.reduce((sum, [_, count]) => sum + count, 0);
    

    let html = '<div class="type-chart">';
    
    typeArray.forEach(([type, count]) => {
        const percentage = total > 0 ? (count / total * 100).toFixed(1) : 0;
        const translatedType = translateTransactionType(type);
        
        html += `
            <div class="type-row">
                <div class="type-label">${translatedType}</div>
                <div class="type-bar-container">
                    <div class="type-bar" style="width: ${percentage}%"></div>
                </div>
                <div class="type-value">${count} (${percentage}%)</div>
            </div>
        `;
    });
    
    html += '</div>';
    container.innerHTML = html;
}

function translateTransactionType(type) {
    const translations = {
        'create': 'Создание депозита',
        'transfer': 'Перевод',
        'freeze': 'Заморозка',
        'block': 'Блокировка',
        'unblock': 'Разблокировка',
        'delete': 'Удаление',
        'cancel_transfer': 'Отмена перевода',
        'cancel_freeze': 'Отмена заморозки',
        'cancel_block': 'Отмена блокировки',
        'cancel_unblock': 'Отмена разблокировки'
    };
    
    return translations[type] || type;
}

async function showCancelActionModal(userId) {
    try {

        document.getElementById('cancel-user-name').textContent = 'Загрузка...';
        document.getElementById('cancel-action-id').textContent = '...';
        document.getElementById('cancel-action-type').textContent = '...';
        document.getElementById('cancel-action-date').textContent = '...';
        document.getElementById('cancel-action-amount').textContent = '...';
        

        document.getElementById('cancel-action-modal').style.display = 'block';
        

        const response = await fetch(`/operator/users/${userId}/last-action`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Не удалось получить информацию о последнем действии');
        }
        
        const result = await response.json();
        
        if (!result.action) {
            showNotification('error', 'У пользователя нет действий для отмены');
            document.getElementById('cancel-action-modal').style.display = 'none';
            return;
        }
        
        // Check if action is of type 'delete'
        if (result.action.type === 'delete') {
            showNotification('error', 'Операции удаления не могут быть отменены');
            document.getElementById('cancel-action-modal').style.display = 'none';
            return;
        }
        
        // Display the action details in the modal
        document.getElementById('cancel-user-name').textContent = result.username || `Пользователь #${userId}`;
        document.getElementById('cancel-action-id').textContent = result.action.id;
        document.getElementById('cancel-action-type').textContent = translateTransactionType(result.action.type);
        document.getElementById('cancel-action-date').textContent = new Date(result.action.timestamp).toLocaleString();
        document.getElementById('cancel-action-amount').textContent = result.action.amount ? formatCurrency(result.action.amount) : '-';
        
        // Set action ID on cancel button
        document.getElementById('confirm-cancel-action').setAttribute('data-action-id', result.action.id);
        document.getElementById('confirm-cancel-action').setAttribute('data-user-id', userId);
        
    } catch (error) {
        console.error('Error getting last action:', error);
        document.getElementById('cancel-action-modal').style.display = 'none';
        showNotification('error', error.message);
    }
}

async function cancelLastUserAction(userId, actionId) {
    try {
        // Show a processing notification
        const notificationId = showNotification('processing', `Отмена действия #${actionId}...`);
        
        // Make the API request
        const response = await fetch('/operator/cancel-action', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: parseInt(userId),
                action_id: parseInt(actionId)
            })
        });
        
        const result = await response.json();
        
        if (!response.ok) {
            // Handle API errors
            const errorMessage = result.error || 'Не удалось отменить действие';
            updateNotification(notificationId, 'error', errorMessage);
            throw new Error(errorMessage);
        }
        
        // If successful, update the notification
        updateNotification(notificationId, 'success', `Действие #${actionId} успешно отменено`);
        
        // Close the modal
        document.getElementById('cancel-action-modal').style.display = 'none';
        
        // Refresh the user list and statistics
        loadUsers();
        loadStatistics();
        loadRecentActions(); // Also refresh recent actions
        
    } catch (error) {
        console.error('Error cancelling action:', error);
        // Don't show another notification here since updateNotification already shows the error
    }
}

function showNotification(type, message) {
    // Check if a notification with the same message already exists
    const existingNotifications = document.querySelectorAll('.loan-notification');
    for (let notif of existingNotifications) {
        const content = notif.querySelector('.notification-content');
        if (content && content.textContent === message) {
            // Return the ID of the existing notification instead of creating a duplicate
            return notif.id;
        }
    }
    
    // Generate a unique ID for the notification
    const id = 'notification-' + Date.now();
    
    // Create notification element
    const notificationElement = document.createElement('div');
    notificationElement.id = id;
    notificationElement.className = `loan-notification ${type}`;
    
    // Determine icon based on type
    let icon;
    switch (type) {
        case 'success': icon = '<i class="fas fa-check-circle notification-icon"></i>'; break;
        case 'error': icon = '<i class="fas fa-exclamation-circle notification-icon"></i>'; break;
        case 'processing': icon = '<i class="fas fa-spinner fa-spin notification-icon"></i>'; break;
        default: icon = '<i class="fas fa-info-circle notification-icon"></i>';
    }
    
    // Set notification content
    notificationElement.innerHTML = `
        ${icon}
        <div class="notification-content">${message}</div>
        <button class="notification-close" onclick="closeNotification('${id}')">&times;</button>
    `;
    
    // Add notification to container
    const container = document.getElementById('loan-notifications');
    container.appendChild(notificationElement);
    
    // Make it visible with animation
    setTimeout(() => {
        notificationElement.classList.add('visible');
    }, 10);
    
    // Auto-close success notifications after 5 seconds
    if (type === 'success') {
        setTimeout(() => {
            closeNotification(id);
        }, 5000);
    }
    
    return id;
}

function updateNotification(id, type, message) {
    const notificationElement = document.getElementById(id);
    if (!notificationElement) return;
    
    // Update class
    notificationElement.className = `loan-notification ${type} visible`;
    
    // Determine icon based on type
    let icon;
    switch (type) {
        case 'success': icon = '<i class="fas fa-check-circle notification-icon"></i>'; break;
        case 'error': icon = '<i class="fas fa-exclamation-circle notification-icon"></i>'; break;
        case 'processing': icon = '<i class="fas fa-spinner fa-spin notification-icon"></i>'; break;
        default: icon = '<i class="fas fa-info-circle notification-icon"></i>';
    }
    
    // Update content
    notificationElement.innerHTML = `
        ${icon}
        <div class="notification-content">${message}</div>
        <button class="notification-close" onclick="closeNotification('${id}')">&times;</button>
    `;
    
    // Auto-close success notifications after 5 seconds
    if (type === 'success') {
        setTimeout(() => {
            closeNotification(id);
        }, 5000);
    }
}

function closeNotification(id) {
    const notification = document.getElementById(id);
    if (notification) {
        notification.classList.remove('visible');
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'RUB',
        minimumFractionDigits: 2
    }).format(amount);
}

function logoutUser() {
    // Clear session storage
    sessionStorage.clear();
    // Redirect to login page
    window.location.href = '/auth';
}

function clearAllCookies() {
    const cookies = document.cookie.split("; ");
    for (let c = 0; c < cookies.length; c++) {
        const d = window.location.hostname.split(".");
        while (d.length > 0) {
            const cookieBase = encodeURIComponent(cookies[c].split(";")[0].split("=")[0]) + '=; expires=Thu, 01-Jan-1970 00:00:01 GMT; domain=' + d.join('.') + ' ;path=';
            const p = location.pathname.split('/');
            document.cookie = cookieBase + '/';
            while (p.length > 0) {
                document.cookie = cookieBase + p.join('/');
                p.pop();
            }
            d.shift();
        }
    }
    showNotification('success', 'Все cookie успешно удалены');
}
