document.addEventListener('DOMContentLoaded', function() {
    // Check authorization
    const authToken = sessionStorage.getItem('authToken');
    const userRole = sessionStorage.getItem('userRole');
    
    if (authToken && (userRole === 'manager' || userRole === 'admin')) {
        document.getElementById('app-content').style.display = 'block';
        document.getElementById('auth-check').style.display = 'none';
        
        // Initialize user info
        const userID = sessionStorage.getItem('userID');
        const username = sessionStorage.getItem('username') || userID;
        
        const userInfo = document.getElementById('user-info');
        userInfo.innerHTML = `
            <div class="user-info-details">
                <span class="username">${username}</span>
                <span class="user-role manager">${userRole === 'admin' ? 'Administrator' : 'Manager'}</span>
            </div>
            <button onclick="logoutUser()">Logout</button>
        `;
        userInfo.style.display = 'flex';
        
        // Load initial data
        loadStatistics();

     
        const tabButtons = document.querySelectorAll('.tab-button');
        tabButtons.forEach(button => {
            button.addEventListener('click', function() {
                const targetTabId = this.getAttribute('data-tab');
                

                tabButtons.forEach(btn => btn.classList.remove('active'));
                this.classList.add('active');
                

                const tabContents = document.querySelectorAll('.tab-content');
                tabContents.forEach(content => content.classList.remove('active'));
                document.getElementById(targetTabId).classList.add('active');
                

                if (targetTabId === 'transactions-tab') {
                    loadTransactions();
                } else if (targetTabId === 'stats-tab') {
                    loadStatistics();
                } else if (targetTabId === 'loans-tab') {
                    loadLoans();
                }
            });
        });
        

        document.getElementById('apply-filters').addEventListener('click', function() {
            loadTransactions();
        });
        

        document.getElementById('refresh-loans').addEventListener('click', function() {
            loadLoans();
        });
        
        // Set up loan status filter
        document.getElementById('loan-status-filter').addEventListener('change', function() {
            loadLoans();
        });
        
        // Set up cancel transaction modal
        document.querySelector('#cancel-transaction-modal .close-modal').addEventListener('click', function() {
            document.getElementById('cancel-transaction-modal').style.display = 'none';
        });
        
        document.getElementById('cancel-modal-close').addEventListener('click', function() {
            document.getElementById('cancel-transaction-modal').style.display = 'none';
        });
        
        document.getElementById('confirm-cancel-transaction').addEventListener('click', function() {
            const transactionId = this.getAttribute('data-transaction-id');
            cancelTransaction(transactionId);
        });
        
        // Set up loan review modal
        document.querySelector('#review-loan-modal .close-modal').addEventListener('click', function() {
            document.getElementById('review-loan-modal').style.display = 'none';
        });
        
        document.getElementById('loan-modal-close').addEventListener('click', function() {
            document.getElementById('review-loan-modal').style.display = 'none';
        });
        
        document.getElementById('reject-loan').addEventListener('click', function() {
            const loanId = this.getAttribute('data-loan-id');
            const comment = document.getElementById('review-loan-comment').value;
            
            if (!comment) {
                showLoanNotification('error', 'Please provide a comment for the rejection');
                return;
            }
            
            reviewLoan(loanId, 'reject', comment);
        });
        
        document.getElementById('approve-loan').addEventListener('click', function() {
            const loanId = this.getAttribute('data-loan-id');
            const comment = document.getElementById('review-loan-comment').value;
            reviewLoan(loanId, 'approve', comment);
        });
    } else {
        document.getElementById('app-content').style.display = 'none';
        document.getElementById('auth-check').style.display = 'block';
    }
});

async function loadStatistics() {
    try {
        const response = await fetch('/manager/transactions/statistics', {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to load statistics');
        }
        
        const stats = await response.json();
        
        document.getElementById('total-transactions').textContent = stats.total_transactions || 0;
        document.getElementById('total-amount').textContent = formatCurrency(stats.total_amount || 0);
        document.getElementById('active-users').textContent = stats.active_users || 0;
        document.getElementById('avg-transaction').textContent = formatCurrency(stats.avg_transaction || 0);
    } catch (error) {
        console.error('Error loading statistics:', error);
        showLoanNotification('error', 'Failed to load statistics: ' + error.message);
    }
}

async function loadTransactions() {
    const tbody = document.getElementById('transactions-list');
    tbody.innerHTML = '<tr><td colspan="6" class="loading-row">Loading transactions...</td></tr>';
    
    try {
        // Get filter values
        const username = document.getElementById('username-filter').value;
        const type = document.getElementById('type-filter').value;
        const date = document.getElementById('date-filter').value;
        
        // Build query string
        let queryParams = [];
        if (username) queryParams.push(`username=${encodeURIComponent(username)}`);
        if (type) queryParams.push(`type=${encodeURIComponent(type)}`);
        if (date) queryParams.push(`date=${encodeURIComponent(date)}`);
        
        const queryString = queryParams.length > 0 ? '?' + queryParams.join('&') : '';
        
        const response = await fetch('/manager/transactions' + queryString, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to load transactions');
        }
        
        const result = await response.json();
        
        if (!result.transactions || result.transactions.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="empty-row">No transactions found</td></tr>';
            return;
        }
        
        let tableHTML = '';
        
        result.transactions.forEach(tx => {
            // Format date
            const date = new Date(tx.timestamp).toLocaleString();
            
            // Format amount
            const amount = tx.amount ? formatCurrency(tx.amount) : '—';
            
            // Determine if already cancelled and by whom
            const isCancelled = tx.cancelled === true;
            let cancelInfo = '';
            
            if (isCancelled && tx.cancelled_by) {
                const cancelTime = new Date(tx.cancel_time || Date.now()).toLocaleString();
                cancelInfo = `Cancelled by ID ${tx.cancelled_by} on ${cancelTime}`;
            }
            
            // Determine if transaction is cancellable (e.g., not already cancelled, not too old)
            const canCancel = !isCancelled && isTransactionCancellable(tx);
            
            tableHTML += `
                <tr>
                    <td>${tx.id}</td>
                    <td>${tx.username || '—'} (#${tx.user_id})</td>
                    <td><span class="badge ${tx.type}">${tx.type}</span></td>
                    <td>${amount}</td>
                    <td>${date}</td>
                    <td>
                        ${isCancelled ? 
                            `<span class="action-disabled" title="${cancelInfo}">Cancelled</span>` : 
                            canCancel ?
                                `<button class="action-btn danger small" onclick="showCancelModal(${tx.id})">
                                    <i class="fas fa-ban"></i> Cancel
                                </button>` :
                                `<span class="action-disabled" title="This transaction can no longer be cancelled">Not cancellable</span>`
                        }
                    </td>
                </tr>
            `;
        });
        
        tbody.innerHTML = tableHTML;
    } catch (error) {
        console.error('Error loading transactions:', error);
        tbody.innerHTML = `<tr><td colspan="6" class="error-row">${error.message}</td></tr>`;
        showLoanNotification('error', 'Failed to load transactions: ' + error.message);
    }
}

async function loadLoans() {
    const tbody = document.getElementById('loans-list');
    tbody.innerHTML = '<tr><td colspan="9" class="loading-row">Loading loans...</td></tr>';
    
    try {
        const status = document.getElementById('loan-status-filter').value || 'pending';
        
        const response = await fetch(`/manager/loans/pending?status=${status}`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to load loans');
        }
        
        const result = await response.json();
        
        if (!result.loans || result.loans.length === 0) {
            tbody.innerHTML = `<tr><td colspan="9" class="empty-row">No ${status} loans found</td></tr>`;
            return;
        }
        
        let tableHTML = '';
        
        result.loans.forEach(loan => {
            // Format date
            const date = new Date(loan.created_at).toLocaleString();
            
            // Format amounts
            const amount = formatCurrency(loan.amount);
            const totalPayable = formatCurrency(loan.total_payable);
            
            // Determine if needs review (pending status)
            const needsReview = loan.needs_review === true || loan.status === 'pending';
            
            tableHTML += `
                <tr>
                    <td>${loan.id}</td>
                    <td>${loan.username || '—'} (#${loan.user_id})</td>
                    <td>${loan.type}</td>
                    <td>${amount}</td>
                    <td>${loan.term} months</td>
                    <td>${loan.interest_rate}%</td>
                    <td>${totalPayable}</td>
                    <td>${date}</td>
                    <td>
                        ${needsReview ? 
                            `<button class="action-btn primary small" onclick="showReviewModal(${loan.id})">
                                <i class="fas fa-search"></i> Review
                            </button>` : 
                            `<span class="status-badge ${loan.status.toLowerCase()}">${loan.status}</span>`
                        }
                    </td>
                </tr>
            `;
        });
        
        tbody.innerHTML = tableHTML;
    } catch (error) {
        console.error('Error loading loans:', error);
        tbody.innerHTML = `<tr><td colspan="9" class="error-row">${error.message}</td></tr>`;
        showLoanNotification('error', 'Failed to load loans: ' + error.message);
    }
}

function showCancelModal(transactionId) {
    // Update the modal content with transaction details
    document.getElementById('cancel-transaction-id').textContent = transactionId;
    document.getElementById('confirm-cancel-transaction').setAttribute('data-transaction-id', transactionId);
    
    // Show an additional warning about the consequences of cancellation
    const warningElement = document.createElement('p');
    warningElement.className = 'warning';
    warningElement.innerHTML = `<i class="fas fa-exclamation-triangle"></i> Cancelling transaction #${transactionId} will revert its effects. This action cannot be undone.`;
    
    // Replace any existing warning
    const existingWarning = document.querySelector('#cancel-transaction-modal .warning');
    if (existingWarning) {
        existingWarning.replaceWith(warningElement);
    } else {
        const modalContent = document.querySelector('#cancel-transaction-modal .modal-content');
        modalContent.insertBefore(warningElement, document.querySelector('#cancel-transaction-modal .modal-actions'));
    }
    
    // Display the modal
    document.getElementById('cancel-transaction-modal').style.display = 'block';
}

async function cancelTransaction(transactionId) {
    try {
        // Show a processing notification while the cancellation is being processed
        const notificationId = showLoanNotification('processing', `Cancelling transaction #${transactionId}...`);
        
        // Make the API request to cancel the transaction
        const response = await fetch('/manager/transactions/cancel', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                transaction_id: parseInt(transactionId)
            })
        });
        
        // Parse the response
        const result = await response.json();
        
        if (!response.ok) {
            // Handle API errors (e.g., transaction not found, already cancelled, etc.)
            const errorMessage = result.error || 'Failed to cancel transaction';
            updateLoanNotification(notificationId, 'error', errorMessage);
            throw new Error(errorMessage);
        }
        
        // If successful, update the notification and reload the transactions list
        updateLoanNotification(notificationId, 'success', `Transaction #${transactionId} cancelled successfully`);
        
        // Close the modal
        document.getElementById('cancel-transaction-modal').style.display = 'none';
        
        // Refresh the transactions list to show the updated status
        loadTransactions();
        
        // Also refresh the statistics since cancellation affects them
        loadStatistics();
        
        // Return success for potential additional handling
        return { success: true, message: `Transaction #${transactionId} cancelled successfully` };
    } catch (error) {
        console.error('Error cancelling transaction:', error);
        showLoanNotification('error', error.message || 'Failed to cancel transaction');
        
        // Return failure for potential additional handling
        return { success: false, error: error.message || 'Failed to cancel transaction' };
    }
}

// Helper function to determine if a transaction can be cancelled
function isTransactionCancellable(transaction) {
    // Implement business rules for transaction cancellation
    if (transaction.cancelled) {
        return false; // Already cancelled
    }

    // Check transaction type - some types might not be cancellable
    const uncancellableTypes = ['system', 'fee', 'interest', 'correction'];
    if (uncancellableTypes.includes(transaction.type)) {
        return false;
    }
    
    // Check transaction age - e.g., only cancel transactions less than 30 days old
    const transactionDate = new Date(transaction.timestamp);
    const now = new Date();
    const daysDifference = (now - transactionDate) / (1000 * 60 * 60 * 24);
    
    if (daysDifference > 30) {
        return false; // Too old to cancel
    }
    
    return true;
}

async function showReviewModal(loanId) {
    try {
        // Show loading state
        document.getElementById('review-loan-id').textContent = 'Loading...';
        document.getElementById('review-loan-user').textContent = 'Loading...';
        document.getElementById('review-loan-type').textContent = 'Loading...';
        document.getElementById('review-loan-amount').textContent = 'Loading...';
        document.getElementById('review-loan-term').textContent = 'Loading...';
        document.getElementById('review-loan-interest').textContent = 'Loading...';
        document.getElementById('review-loan-monthly').textContent = 'Loading...';
        document.getElementById('review-loan-total').textContent = 'Loading...';
        document.getElementById('review-loan-comment').value = '';
        
        document.getElementById('review-loan-modal').style.display = 'block';
        
        // Set loan ID on buttons
        document.getElementById('reject-loan').setAttribute('data-loan-id', loanId);
        document.getElementById('approve-loan').setAttribute('data-loan-id', loanId);
        
        // Fetch loan details
        const response = await fetch(`/loan/${loanId}`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to load loan details');
        }
        
        const result = await response.json();
        const loan = result.loan;
        
        // Get username
        const userResponse = await fetch(`/users/${loan.user_id}`, {
            method: 'GET',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            }
        }).catch(() => ({ ok: false }));
        
        let username = 'User #' + loan.user_id;
        if (userResponse && userResponse.ok) {
            const userData = await userResponse.json();
            username = userData.username || username;
        }
        
        // Fill in loan details
        document.getElementById('review-loan-id').textContent = loan.id;
        document.getElementById('review-loan-user').textContent = username;
        document.getElementById('review-loan-type').textContent = loan.type;
        document.getElementById('review-loan-amount').textContent = formatCurrency(loan.amount);
        document.getElementById('review-loan-term').textContent = `${loan.term} months`;
        document.getElementById('review-loan-interest').textContent = `${loan.interest_rate}%`;
        document.getElementById('review-loan-monthly').textContent = formatCurrency(loan.monthly_payment);
        document.getElementById('review-loan-total').textContent = formatCurrency(loan.total_payable);
    } catch (error) {
        console.error('Error loading loan details:', error);
        showLoanNotification('error', error.message);
        document.getElementById('review-loan-modal').style.display = 'none';
    }
}

async function reviewLoan(loanId, action, comment) {
    try {
        const notificationId = showLoanNotification('processing', `Processing loan ${action}...`);
        
        const response = await fetch('/manager/loans/review', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + sessionStorage.getItem('authToken'),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                loan_id: parseInt(loanId),
                action: action,
                comment: comment
            })
        });
        
        const result = await response.json();
        
        if (!response.ok) {
            throw new Error(result.error || `Failed to ${action} loan`);
        }
        
        updateLoanNotification(
            notificationId, 
            'success', 
            `Loan #${loanId} ${action === 'approve' ? 'approved' : 'rejected'} successfully`
        );
        
        document.getElementById('review-loan-modal').style.display = 'none';
        loadLoans();
    } catch (error) {
        console.error(`Error ${action}ing loan:`, error);
        showLoanNotification('error', error.message);
    }
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2
    }).format(amount);
}

// Show loan operation notification banner
function showLoanNotification(type, message) {
    // Check if notifications container exists, if not create it
    let notificationsContainer = document.getElementById('loan-notifications');
    if (!notificationsContainer) {
        notificationsContainer = document.createElement('div');
        notificationsContainer.id = 'loan-notifications';
        document.body.appendChild(notificationsContainer);
    }
    
    // Create notification element
    const notificationId = 'loan-notification-' + Date.now();
    const notification = document.createElement('div');
    notification.id = notificationId;
    notification.className = `loan-notification ${type}`;
    
    // Set icon based on notification type
    let icon = '';
    switch (type) {
        case 'success':
            icon = '<i class="fas fa-check-circle"></i>';
            break;
        case 'error':
            icon = '<i class="fas fa-exclamation-circle"></i>';
            break;
        case 'processing':
            icon = '<i class="fas fa-circle-notch fa-spin"></i>';
            break;
        default:
            icon = '<i class="fas fa-info-circle"></i>';
    }
    
    notification.innerHTML = `
        <div class="notification-icon">${icon}</div>
        <div class="notification-content">${message}</div>
        <button class="notification-close" onclick="closeLoanNotification('${notificationId}')">
            <i class="fas fa-times"></i>
        </button>
    `;
    
    // Add to notifications container
    notificationsContainer.appendChild(notification);
    
    // Animate in
    setTimeout(() => notification.classList.add('visible'), 10);
    
    // Auto dismiss success notifications after 5 seconds
    if (type === 'success') {
        setTimeout(() => closeLoanNotification(notificationId), 5000);
    }
    
    return notificationId;
}

// Update existing loan notification
function updateLoanNotification(notificationId, type, message) {
    const notification = document.getElementById(notificationId);
    if (!notification) return;
    
    // Update class
    notification.className = `loan-notification ${type} visible`;
    
    // Update icon
    let icon = '';
    switch (type) {
    
        case 'success':
            icon = '<i class="fas fa-check-circle"></i>';
            break;
        case 'error':
            icon = '<i class="fas fa-exclamation-circle"></i>';
            break;
        case 'processing':
            icon = '<i class="fas fa-circle-notch fa-spin"></i>';
            break;
        default:
            icon = '<i class="fas fa-info-circle"></i>';
    }
    
    // Update content
    notification.querySelector('.notification-icon').innerHTML = icon;
    notification.querySelector('.notification-content').textContent = message;
    
    // Auto dismiss success notifications after 5 seconds
    if (type === 'success') {
        setTimeout(() => closeLoanNotification(notificationId), 5000);
    }
}

// Close loan notification
function closeLoanNotification(notificationId) {
    const notification = document.getElementById(notificationId);
    if (!notification) return;
    
    // Animate out
    notification.classList.remove('visible');
    
    // Remove after animation completes
    setTimeout(() => {
        if (notification && notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 300);
}

// Make the closeLoanNotification function globally accessible
window.closeLoanNotification = closeLoanNotification;
