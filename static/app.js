async function makeRequest(endpoint, method, data) {
    try {
        const button = document.querySelector(`button[onclick="${getOperationName(endpoint)}()"]`);
        button.classList.add('processing');
        
        if (!navigator.onLine) {
            throw new Error('Нет подключения к интернету');
        }

        const apiEndpoint = `/deposit${endpoint}`;
        const response = await fetch(apiEndpoint, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        }).catch(error => {
            throw new Error(`Ошибка сети: ${error.message}`);
        });

        let result;
        try {
            result = await response.json();
        } catch (e) {
            throw new Error('Некорректный формат ответа от сервера');
        }
        
        if (response.ok) {
            const successMessage = getSuccessMessage(endpoint);
            showFeedback('success', successMessage);
        } else {
            const errorMsg = result.error || result.message || getErrorMessage(endpoint, response.status);
            showFeedback('error', errorMsg);
        }

        return { status: response.status, data: result };
    } catch (error) {
        const errorMsg = getNetworkErrorMessage(error);
        showFeedback('error', errorMsg);
        return { status: 500, data: { error: error.message } };
    } finally {
        const button = document.querySelector(`button[onclick="${getOperationName(endpoint)}()"]`);
        if (button) button.classList.remove('processing');
    }
}

function getNetworkErrorMessage(error) {
    const networkErrors = {
        'Failed to fetch': 'Не удалось подключиться к серверу',
        'NetworkError': 'Проблема с сетевым подключением',
        'Нет подключения к интернету': 'Отсутствует подключение к интернету',
        'Некорректный формат ответа от сервера': 'Сервер вернул некорректные данные'
    };

    for (const [errorText, message] of Object.entries(networkErrors)) {
        if (error.message.includes(errorText)) {
            return message;
        }
    }
    
    return `Ошибка операции: ${error.message}`;
}

function getErrorMessage(endpoint, statusCode) {
    const commonErrors = {
        400: 'Неверные параметры запроса',
        401: 'Необходима авторизация',
        403: 'Доступ запрещен',
        404: 'Ресурс не найден',
        408: 'Время ожидания истекло',
        429: 'Слишком много запросов',
        500: 'Внутренняя ошибка сервера',
        502: 'Сервер недоступен',
        503: 'Сервис временно недоступен',
        504: 'Время ожидания ответа истекло'
    };

    const specificErrors = {
        '/create': {
            400: 'Неверные данные для создания депозита',
            409: 'Депозит с такими параметрами уже существует'
        },
        '/deposit/transfer': {
            400: 'Неверные параметры перевода',
            404: 'Один из счетов не найден',
            409: 'Недостаточно средств или счет заблокирован'
        },
        '/deposit/freeze': {
            400: 'Неверные параметры заморозки',
            404: 'Депозит не найден',
            409: 'Депозит уже заморожен или заблокирован'
        },
        '/deposit/block': {
            400: 'Неверные параметры блокировки',
            404: 'Депозит не найден',
            409: 'Депозит уже заблокирован'
        },
        '/deposit/unblock': {
            400: 'Неверные параметры разблокировки',
            404: 'Депозит не найден',
            403: 'Недостаточно прав для разблокировки'
        },
        '/deposit/delete': {
            400: 'Неверные параметры для удаления',
            404: 'Депозит не найден',
            409: 'Депозит не может быть удален: есть активные операции'
        }
    };

    return (specificErrors[endpoint] && specificErrors[endpoint][statusCode]) || 
           commonErrors[statusCode] || 
           'Неизвестная ошибка при выполнении операции';
}

function getOperationName(endpoint) {
    const operations = {
        '/create': 'createDeposit',
        '/transfer': 'transferBetweenAccounts',
        '/freeze': 'freezeDeposit',
        '/block': 'blockDeposit',
        '/unblock': 'unblockDeposit',
        '/delete': 'deleteDeposit'
    };
    return operations[endpoint];
}

function getSuccessMessage(endpoint) {
    const messages = {
        '/create': 'Депозит успешно создан! 🎉',
        '/transfer': 'Перевод успешно выполнен! 💸',
        ' /freeze': 'Депозит успешно заморожен! ❄️',
        ' /block': 'Депозит успешно заблокирован! 🔒',
        ' /unblock': 'Депозит успешно разблокирован! 🔓',
        ' /delete': 'Депозит успешно удален! 🗑️'
    };
    return messages[endpoint];
}

function showFeedback(type, message) {
    const feedbackDiv = document.createElement('div');
    feedbackDiv.className = `feedback-message ${type}`;
    
    const icons = {
        processing: '⏳',
        success: '✅',
        error: '❌'
    };

    feedbackDiv.innerHTML = `
        <span class="feedback-icon">${icons[type]}</span>
        <span class="feedback-text">${message}</span>
        <button onclick="this.parentElement.remove()" class="feedback-close">×</button>
    `;

    // Remove any existing feedback after 3 seconds
    const existingFeedback = document.querySelector('.feedback-message');
    if (existingFeedback) {
        existingFeedback.remove();
    }

    document.body.appendChild(feedbackDiv);
    
    // Auto-hide after 3 seconds for success/error messages
    if (type !== 'processing') {
        setTimeout(() => {
            feedbackDiv.classList.add('fade-out');
            setTimeout(() => feedbackDiv.remove(), 500);
        }, 3000);
    }
}

async function createDeposit() {
    const button = document.querySelector('button[onclick="createDeposit()"]');
    button.classList.add('processing');
    
    const data = {
        client_id: parseInt(document.getElementById('create_client_id').value),
        bank_name: document.getElementById('create_bank_name').value,
        amount: parseFloat(document.getElementById('create_amount').value),
        interest: parseFloat(document.getElementById('create_interest').value)
    };

    try {
        const result = await makeRequest('/create', 'POST', data);
        if (result.status >= 200 && result.status < 300) {
            showFeedback('success', 'Депозит успешно создан! 🎉');
            // Clear form
            ['create_client_id', 'create_bank_name', 'create_amount', 'create_interest']
                .forEach(id => document.getElementById(id).value = '');
        } else {
            showFeedback('error', 'Ошибка при создании депозита');
        }
    } finally {
        button.classList.remove('processing');
    }
}

async function transferBetweenAccounts() {
    const button = document.querySelector('button[onclick="transferBetweenAccounts()"]');
    button.classList.add('processing');
    
    const data = {
        client_id: parseInt(document.getElementById('transfer_client_id').value),
        bank_name: document.getElementById('transfer_bank_name').value,
        from_account: parseInt(document.getElementById('from_account').value),
        to_account: parseInt(document.getElementById('to_account').value),
        amount: parseFloat(document.getElementById('transfer_amount').value)
    };

    try {
        const result = await makeRequest('/deposit/transfer', 'POST', data);
        if (result.status >= 200 && result.status < 300) {
            showFeedback('success', 'Перевод успешно выполнен! 💸');
        } else {
            let errorMsg = 'Ошибка при выполнении перевода';
            switch (result.status) {
                case 400:
                    errorMsg = 'Неверные данные для перевода';
                    break;
                case 404:
                    errorMsg = 'Один из счетов не найден';
                    break;
                case 403:
                    errorMsg = 'Недостаточно прав для выполнения перевода';
                    break;
                case 409:
                    errorMsg = 'Недостаточно средств на счете';
                    break;
            }
            showFeedback('error', errorMsg);
        }
    } finally {
        button.classList.remove('processing');
    }
}

async function freezeDeposit() {
    const button = document.querySelector('button[onclick="freezeDeposit()"]');
    button.classList.add('processing');
    
    const data = {
        client_id: parseInt(document.getElementById('freeze_client_id').value),
        bank_name: document.getElementById('freeze_bank_name').value,
        freeze_duration: parseInt(document.getElementById('freeze_duration').value)
    };

    try {
        const result = await makeRequest('/deposit/freeze', 'POST', data);
        if (result.status >= 200 && result.status < 300) {
            showFeedback('success', 'Депозит успешно заморожен! ❄️');
        } else {
            let errorMsg = 'Ошибка при заморозке депозита';
            switch (result.status) {
                case 400:
                    errorMsg = 'Неверные параметры заморозки';
                    break;
                case 404:
                    errorMsg = 'Депозит не найден';
                    break;
                case 409:
                    errorMsg = 'Депозит уже заморожен или заблокирован';
                    break;
            }
            showFeedback('error', errorMsg);
        }
    } finally {
        button.classList.remove('processing');
    }
}

async function blockDeposit() {
    const button = document.querySelector('button[onclick="blockDeposit()"]');
    button.classList.add('processing');
    
    const data = {
        client_id: parseInt(document.getElementById('block_client_id').value),
        bank_name: document.getElementById('block_bank_name').value
    };

    try {
        const result = await makeRequest('/deposit/block', 'POST', data);
        if (result.status >= 200 && result.status < 300) {
            showFeedback('success', 'Депозит успешно заблокирован! 🔒');
        } else {
            let errorMsg = 'Ошибка при блокировке депозита';
            switch (result.status) {
                case 400:
                    errorMsg = 'Неверные параметры блокировки';
                    break;
                case 404:
                    errorMsg = 'Депозит не найден';
                    break;
                case 409:
                    errorMsg = 'Депозит уже заблокирован';
                    break;
            }
            showFeedback('error', errorMsg);
        }
    } finally {
        button.classList.remove('processing');
    }
}

async function unblockDeposit() {
    const button = document.querySelector('button[onclick="unblockDeposit()"]');
    button.classList.add('processing');
    
    const data = {
        client_id: parseInt(document.getElementById('unblock_client_id').value),
        bank_name: document.getElementById('unblock_bank_name').value
    };

    try {
        const result = await makeRequest('/deposit/unblock', 'POST', data);
        if (result.status >= 200 && result.status < 300) {
            showFeedback('success', 'Депозит успешно разблокирован! 🔓');
        } else {
            let errorMsg = 'Ошибка при разблокировке депозита';
            switch (result.status) {
                case 400:
                    errorMsg = 'Неверные параметры разблокировки';
                    break;
                case 404:
                    errorMsg = 'Депозит не найден';
                    break;
                case 403:
                    errorMsg = 'Недостаточно прав для разблокировки';
                    break;
            }
            showFeedback('error', errorMsg);
        }
    } finally {
        button.classList.remove('processing');
    }
}

async function deleteDeposit() {
    const button = document.querySelector('button[onclick="deleteDeposit()"]');
    button.classList.add('processing');
    
    const data = {
        client_id: parseInt(document.getElementById('delete_client_id').value),
        bank_name: document.getElementById('delete_bank_name').value
    };

    try {
        const result = await makeRequest('/deposit/delete', 'DELETE', data);
        if (result.status >= 200 && result.status < 300) {
            showFeedback('success', 'Депозит успешно удален! 🗑️');
        } else {
            let errorMsg = 'Ошибка при удалении депозита';
            switch (result.status) {
                case 400:
                    errorMsg = 'Неверные параметры для удаления';
                    break;
                case 404:
                    errorMsg = 'Депозит не найден';
                    break;
                case 409:
                    errorMsg = 'Депозит не может быть удален: активные транзакции или блокировка';
                    break;
            }
            showFeedback('error', errorMsg);
        }
    } finally {
        button.classList.remove('processing');
    }
}
