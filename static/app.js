async function makeRequest(endpoint, method, data) {
    try {
        const response = await fetch(endpoint, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });
        const result = await response.json();
        return {
            status: response.status,
            data: result
        };
    } catch (error) {
        return {
            status: 500,
            data: { error: error.message }
        };
    }
}

async function createDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('create_client_id').value),
        bank_name: document.getElementById('create_bank_name').value,
        amount: parseFloat(document.getElementById('create_amount').value),
        interest: parseFloat(document.getElementById('create_interest').value)
    };

    const result = await makeRequest('/deposit/create', 'POST', data);
    document.getElementById('create_response').innerHTML = 
        `Status: ${result.status}<br>Response: ${JSON.stringify(result.data, null, 2)}`;
}

async function transferBetweenAccounts() {
    const data = {
        client_id: parseInt(document.getElementById('transfer_client_id').value),
        bank_name: document.getElementById('transfer_bank_name').value,
        from_account: parseInt(document.getElementById('from_account').value),
        to_account: parseInt(document.getElementById('to_account').value),
        amount: parseFloat(document.getElementById('transfer_amount').value)
    };

    const result = await makeRequest('/deposit/transfer', 'POST', data);
    document.getElementById('transfer_response').innerHTML = 
        `Status: ${result.status}<br>Response: ${JSON.stringify(result.data, null, 2)}`;
}

async function freezeDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('freeze_client_id').value),
        bank_name: document.getElementById('freeze_bank_name').value,
        deposit_id: parseInt(document.getElementById('freeze_deposit_id').value),
        freeze_duration: parseInt(document.getElementById('freeze_duration').value)
    };

    const result = await makeRequest('/deposit/freeze', 'POST', data);
    document.getElementById('freeze_response').innerHTML = 
        `Status: ${result.status}<br>Response: ${JSON.stringify(result.data, null, 2)}`;
}

async function blockDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('block_client_id').value),
        bank_name: document.getElementById('block_bank_name').value,
        deposit_id: parseInt(document.getElementById('block_deposit_id').value)
    };

    const result = await makeRequest('/deposit/block', 'POST', data);
    document.getElementById('block_response').innerHTML = 
        `Status: ${result.status}<br>Response: ${JSON.stringify(result.data, null, 2)}`;
}

async function unblockDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('unblock_client_id').value),
        bank_name: document.getElementById('unblock_bank_name').value,
        deposit_id: parseInt(document.getElementById('unblock_deposit_id').value)
    };

    const result = await makeRequest('/deposit/unblock', 'POST', data);
    document.getElementById('unblock_response').innerHTML = 
        `Status: ${result.status}<br>Response: ${JSON.stringify(result.data, null, 2)}`;
}

async function deleteDeposit() {
    const data = {
        client_id: parseInt(document.getElementById('delete_client_id').value),
        bank_name: document.getElementById('delete_bank_name').value
    };

    const result = await makeRequest('/deposit/delete', 'DELETE', data);
    document.getElementById('delete_response').innerHTML = 
        `Status: ${result.status}<br>Response: ${JSON.stringify(result.data, null, 2)}`;
}
