document.addEventListener('DOMContentLoaded', function() {
    // Make operation sections collapsible
    const sections = document.querySelectorAll('.operation-section h3');
    
    sections.forEach(section => {
        section.addEventListener('click', function() {
            const parent = this.parentElement;
            parent.classList.toggle('collapsed');
        });
    });
    
    // Initially collapse all sections except the first one
    const allSections = document.querySelectorAll('.operation-section');
    if (allSections.length > 1) {
        for (let i = 1; i < allSections.length; i++) {
            allSections[i].classList.add('collapsed');
        }
    }
    
    // Pre-fill form fields if we have user info
    const userID = localStorage.getItem('userID');
    if (userID) {
        const clientIDFields = document.querySelectorAll('input[id$="_client_id"]');
        clientIDFields.forEach(field => {
            field.value = userID;
            // Make read-only if not admin
            if (localStorage.getItem('userRole') !== 'admin') {
                field.readOnly = true;
                field.classList.add('readonly');
            }
        });
    }

    // Handle tab switching
    const tabButtons = document.querySelectorAll('.tab-button');
    tabButtons.forEach(button => {
        button.addEventListener('click', function() {
            // Remove active class from all buttons
            tabButtons.forEach(btn => btn.classList.remove('active'));
            
            // Add active class to current button
            this.classList.add('active');
            
            // Hide all tab content
            const tabContents = document.querySelectorAll('.tab-content');
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Show the corresponding tab content
            const targetTabId = this.getAttribute('data-tab');
            document.getElementById(targetTabId).classList.add('active');
            
            // Load user deposits when the my-deposits tab is activated
            if (targetTabId === 'my-deposits') {
                loadUserDeposits();
            }
        });
    });

    // Load deposits tab by default
    if (document.querySelector('.tab-button')) {
        document.querySelector('.tab-button').click();
    }
});

// Check if token is about to expire and refresh it periodically
setInterval(async function() {
    const token = localStorage.getItem('authToken');
    const tokenExpires = localStorage.getItem('tokenExpires');
    
    if (token && tokenExpires) {
        const expiresDate = new Date(tokenExpires);
        const now = new Date();
        
        // If token expires in less than 5 minutes, refresh it
        if ((expiresDate - now) < (5 * 60 * 1000)) {
            await refreshToken();
        }
    }
}, 60000); // Check every minute

// Format timestamp for better readability
function formatTimestamp(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleString();
}

// Clear all response containers
function clearAllResponses() {
    const responseContainers = document.querySelectorAll('.response-container');
    responseContainers.forEach(container => {
        container.innerHTML = '';
        container.classList.remove('visible');
    });
    showFeedback('success', 'All responses cleared');
}

// Clear request log
function clearRequestLog() {
    const requestLog = document.getElementById('request-log-entries');
    if (requestLog) {
        requestLog.innerHTML = '';
        showFeedback('success', 'Request log cleared');
    }
}

// Load user deposits
async function loadUserDeposits() {
    const depositsContainer = document.getElementById('deposits-list');
    depositsContainer.innerHTML = '<div class="loading">Loading your deposits...</div>';
    
    try {
        const response = await fetch('/deposit/list', {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        if (response.ok) {
            if (data.deposits && data.deposits.length > 0) {
                depositsContainer.innerHTML = '';
                
                data.deposits.forEach(deposit => {
                    const depositCard = createDepositCard(deposit);
                    depositsContainer.appendChild(depositCard);
                });
            } else {
                depositsContainer.innerHTML = '<div class="no-deposits">You have no deposits yet. Create one from the Deposit Management tab.</div>';
            }
        } else {
            depositsContainer.innerHTML = `<div class="error-message">Error: ${data.message || 'Failed to load deposits'}</div>`;
        }
    } catch (error) {
        depositsContainer.innerHTML = `<div class="error-message">Error: ${error.message}</div>`;
    }
}

// Create a deposit card element
function createDepositCard(deposit) {
    const card = document.createElement('div');
    card.className = 'deposit-card';
    card.dataset.id = deposit.id;
    
    const statusClass = deposit.blocked ? 'status-blocked' : (deposit.frozen_until ? 'status-frozen' : 'status-active');
    
    card.innerHTML = `
        <div class="deposit-header">
            <span class="deposit-bank">${deposit.bank_name}</span>
            <span class="deposit-status ${statusClass}">
                ${deposit.blocked ? 'Blocked' : (deposit.frozen_until ? 'Frozen' : 'Active')}
            </span>
        </div>
        <div class="deposit-body">
            <div class="deposit-amount">$${deposit.amount.toFixed(2)}</div>
            <div class="deposit-details">
                <div class="deposit-info">
                    <span class="info-label">Interest:</span>
                    <span class="info-value">${deposit.interest_rate}%</span>
                </div>
                <div class="deposit-info">
                    <span class="info-label">ID:</span>
                    <span class="info-value">${deposit.id}</span>
                </div>
                ${deposit.frozen_until ? `
                <div class="deposit-info">
                    <span class="info-label">Frozen until:</span>
                    <span class="info-value">${formatTimestamp(deposit.frozen_until)}</span>
                </div>` : ''}
                <div class="deposit-info">
                    <span class="info-label">Created:</span>
                    <span class="info-value">${formatTimestamp(deposit.created_at)}</span>
                </div>
            </div>
        </div>
        <div class="deposit-actions">
            <button class="action-btn transfer-btn" onclick="showTransferModal(${deposit.id})">Transfer</button>
            ${!deposit.blocked ? 
                `<button class="action-btn ${deposit.frozen_until ? 'unfreeze-btn' : 'freeze-btn'}" 
                    onclick="${deposit.frozen_until ? 'unfreezeDeposit' : 'showFreezeModal'}(${deposit.id})">
                    ${deposit.frozen_until ? 'Unfreeze' : 'Freeze'}
                </button>` : ''}
            ${!deposit.frozen_until ? 
                `<button class="action-btn ${deposit.blocked ? 'unblock-btn' : 'block-btn'}" 
                    onclick="${deposit.blocked ? 'unblockDepositCard' : 'blockDepositCard'}(${deposit.id})">
                    ${deposit.blocked ? 'Unblock' : 'Block'}
                </button>` : ''}
            <button class="action-btn delete-btn" onclick="confirmDeleteDeposit(${deposit.id})">Delete</button>
        </div>
    `;
    
    return card;
}

// Show transfer modal
function showTransferModal(depositId) {
    // Create modal if it doesn't exist
    let modal = document.getElementById('transfer-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'transfer-modal';
        modal.className = 'modal';
        modal.innerHTML = `
            <div class="modal-content">
                <span class="close-modal">&times;</span>
                <h3>Transfer Funds</h3>
                <div class="form-group">
                    <label for="modal_to_account">To Account ID</label>
                    <input type="number" id="modal_to_account" placeholder="Enter destination account ID">
                </div>
                <div class="form-group">
                    <label for="modal_transfer_amount">Amount</label>
                    <input type="number" id="modal_transfer_amount" placeholder="Enter amount to transfer">
                </div>
                <button id="confirm-transfer">Transfer</button>
            </div>
        `;
        document.body.appendChild(modal);
        
        // Add event listener to close button
        modal.querySelector('.close-modal').addEventListener('click', function() {
            modal.style.display = 'none';
        });
    }
    
    // Set the deposit ID and show the modal
    modal.dataset.depositId = depositId;
    modal.style.display = 'block';
    
    // Handle transfer confirmation
    document.getElementById('confirm-transfer').onclick = function() {
        const toAccount = document.getElementById('modal_to_account').value;
        const amount = document.getElementById('modal_transfer_amount').value;
        transferFromDeposit(depositId, toAccount, amount);
        modal.style.display = 'none';
    };
}

// Show freeze modal
function showFreezeModal(depositId) {
    // Create modal if it doesn't exist
    let modal = document.getElementById('freeze-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'freeze-modal';
        modal.className = 'modal';
        modal.innerHTML = `
            <div class="modal-content">
                <span class="close-modal">&times;</span>
                <h3>Freeze Deposit</h3>
                <div class="form-group">
                    <label for="modal_freeze_duration">Freeze Duration (hours)</label>
                    <input type="number" id="modal_freeze_duration" placeholder="Enter freeze duration in hours">
                </div>
                <button id="confirm-freeze">Freeze</button>
            </div>
        `;
        document.body.appendChild(modal);
        
        // Add event listener to close button
        modal.querySelector('.close-modal').addEventListener('click', function() {
            modal.style.display = 'none';
        });
    }
    
    // Set the deposit ID and show the modal
    modal.dataset.depositId = depositId;
    modal.style.display = 'block';
    
    // Handle freeze confirmation
    document.getElementById('confirm-freeze').onclick = function() {
        const duration = document.getElementById('modal_freeze_duration').value;
        freezeDepositCard(depositId, duration);
        modal.style.display = 'none';
    };
}

// Functions to handle deposit actions from cards
async function transferFromDeposit(depositId, toAccount, amount) {
    // Get the bank name from the deposit card
    const depositCard = document.querySelector(`.deposit-card[data-id="${depositId}"]`);
    const bankName = depositCard.querySelector('.deposit-bank').textContent;
    
    try {
        const response = await fetch('/deposit/transfer', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                bank_name: bankName,
                deposit_id: depositId,
                to_account: parseInt(toAccount),
                amount: parseFloat(amount)
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showFeedback('success', 'Transfer successful');
            loadUserDeposits(); // Reload deposits to show updated amounts
        } else {
            showFeedback('error', `Transfer failed: ${data.message}`);
        }
    } catch (error) {
        showFeedback('error', `Error: ${error.message}`);
    }
}

async function freezeDepositCard(depositId, duration) {
    // Get the bank name from the deposit card
    const depositCard = document.querySelector(`.deposit-card[data-id="${depositId}"]`);
    const bankName = depositCard.querySelector('.deposit-bank').textContent;
    
    try {
        const response = await fetch('/deposit/freeze', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                bank_name: bankName,
                deposit_id: depositId,
                freeze_duration: parseInt(duration)
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showFeedback('success', 'Deposit frozen successfully');
            loadUserDeposits(); // Reload deposits to reflect changes
        } else {
            showFeedback('error', `Failed to freeze deposit: ${data.message}`);
        }
    } catch (error) {
        showFeedback('error', `Error: ${error.message}`);
    }
}

async function unfreezeDeposit(depositId) {
    // This would require a backend endpoint to unfreeze a deposit
    showFeedback('warning', 'Unfreeze functionality not implemented yet');
}

async function blockDepositCard(depositId) {
    // Get the bank name from the deposit card
    const depositCard = document.querySelector(`.deposit-card[data-id="${depositId}"]`);
    const bankName = depositCard.querySelector('.deposit-bank').textContent;
    
    try {
        const response = await fetch('/deposit/block', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                bank_name: bankName,
                deposit_id: depositId
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showFeedback('success', 'Deposit blocked successfully');
            loadUserDeposits(); // Reload deposits to reflect changes
        } else {
            showFeedback('error', `Failed to block deposit: ${data.message}`);
        }
    } catch (error) {
        showFeedback('error', `Error: ${error.message}`);
    }
}

async function unblockDepositCard(depositId) {
    // Get the bank name from the deposit card
    const depositCard = document.querySelector(`.deposit-card[data-id="${depositId}"]`);
    const bankName = depositCard.querySelector('.deposit-bank').textContent;
    
    try {
        const response = await fetch('/deposit/unblock', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                bank_name: bankName,
                deposit_id: depositId
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showFeedback('success', 'Deposit unblocked successfully');
            loadUserDeposits(); // Reload deposits to reflect changes
        } else {
            showFeedback('error', `Failed to unblock deposit: ${data.message}`);
        }
    } catch (error) {
        showFeedback('error', `Error: ${error.message}`);
    }
}

function confirmDeleteDeposit(depositId) {
    if (confirm('Are you sure you want to delete this deposit? This action cannot be undone.')) {
        deleteDepositCard(depositId);
    }
}

async function deleteDepositCard(depositId) {
    // Get the bank name from the deposit card
    const depositCard = document.querySelector(`.deposit-card[data-id="${depositId}"]`);
    const bankName = depositCard.querySelector('.deposit-bank').textContent;
    
    try {
        const response = await fetch('/deposit/delete', {
            method: 'DELETE',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                bank_name: bankName,
                deposit_id: depositId
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showFeedback('success', 'Deposit deleted successfully');
            depositCard.remove(); // Remove the card from the UI
            
            // If no deposits left, show the no deposits message
            if (document.querySelectorAll('.deposit-card').length === 0) {
                document.getElementById('deposits-list').innerHTML = 
                    '<div class="no-deposits">You have no deposits yet. Create one from the Deposit Management tab.</div>';
            }
        } else {
            showFeedback('error', `Failed to delete deposit: ${data.message}`);
        }
    } catch (error) {
        showFeedback('error', `Error: ${error.message}`);
    }
}

// Display feedback messages
function showFeedback(type, message) {
    // Implementation of the feedback display method would go here
    console.log(`${type}: ${message}`);
    // This should be filled in with appropriate feedback UI code
}
