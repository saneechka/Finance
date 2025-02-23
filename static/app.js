async function makeRequest(endpoint, method, data) {
    try {
        console.log('Making request:', { endpoint, method, data }); // Debug log

        const apiEndpoint = `/api${endpoint}`;
        const response = await fetch(apiEndpoint, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        console.log('Response received:', response.status); // Debug log

        const result = await response.json();
        console.log('Response data:', result); // Debug log

        // Display response in the UI
        const elementId = endpoint.split('/')[2] + '_response'; // get operation type from endpoint
        const responseElement = document.getElementById(elementId);
        if (responseElement) {
            responseElement.style.display = 'block';
            responseElement.innerHTML = `
                <div class="operation-status ${response.ok ? 'status-success' : 'status-error'}">
                    <span class="status-indicator"></span>
                    <span>${response.ok ? 'Success' : 'Error'}</span>
                </div>
                <div class="operation-details">
                    <strong>Status:</strong> ${response.status}
                    <br>
                    <strong>Response:</strong> ${JSON.stringify(result, null, 2)}
                </div>
            `;
        }

        return { status: response.status, data: result };
    } catch (error) {
        console.error('Request error:', error); // Debug log
        const errorMessage = {
            status: 500,
            data: { error: error.message }
        };
        
        // Show error in UI
        const elementId = endpoint.split('/')[2] + '_response';
        const responseElement = document.getElementById(elementId);
        if (responseElement) {
            responseElement.style.display = 'block';
            responseElement.innerHTML = `
                <div class="operation-status status-error">
                    <span class="status-indicator"></span>
                    <span>Error</span>
                </div>
                <div class="operation-details">
                    <strong>Error:</strong> ${error.message}
                </div>
            `;
        }
        
        return errorMessage;
    }
}

async function createDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('create_client_id').value),
        bank_name: document.getElementById('create_bank_name').value,
        amount: parseFloat(document.getElementById('create_amount').value),
        interest: parseFloat(document.getElementById('create_interest').value)
    };

    console.log('Creating deposit with data:', data); // Debug log
    return await makeRequest('/deposit/create', 'POST', data);
}

async function transferBetweenAccounts() {
    const data = {
        client_id: parseInt(document.getElementById('transfer_client_id').value),
        bank_name: document.getElementById('transfer_bank_name').value,
        from_account: parseInt(document.getElementById('from_account').value),
        to_account: parseInt(document.getElementById('to_account').value),
        amount: parseFloat(document.getElementById('transfer_amount').value)
    };

    console.log('Transferring between accounts with data:', data); // Debug log
    return await makeRequest('/deposit/transfer', 'POST', data);
}

async function freezeDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('freeze_client_id').value),
        bank_name: document.getElementById('freeze_bank_name').value,
        deposit_id: parseInt(document.getElementById('freeze_deposit_id').value),
        freeze_duration: parseInt(document.getElementById('freeze_duration').value)
    };

    console.log('Freezing deposit with data:', data); // Debug log
    return await makeRequest('/deposit/freeze', 'POST', data);
}

async function blockDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('block_client_id').value),
        bank_name: document.getElementById('block_bank_name').value,
        deposit_id: parseInt(document.getElementById('block_deposit_id').value)
    };

    console.log('Blocking deposit with data:', data); // Debug log
    return await makeRequest('/deposit/block', 'POST', data);
}

async function unblockDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('unblock_client_id').value),
        bank_name: document.getElementById('unblock_bank_name').value,
        deposit_id: parseInt(document.getElementById('unblock_deposit_id').value)
    };

    console.log('Unblocking deposit with data:', data); // Debug log
    return await makeRequest('/deposit/unblock', 'POST', data);
}

async function deleteDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('delete_client_id').value),
        bank_name: document.getElementById('delete_bank_name').value
    };

    console.log('Deleting deposit with data:', data); // Debug log
    return await makeRequest('/deposit/delete', 'DELETE', data);
}
