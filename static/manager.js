document.addEventListener('DOMContentLoaded', async function() {
    // Check authentication and role
    const token = sessionStorage.getItem('authToken');
    const role = sessionStorage.getItem('userRole');
    
    if (!token) {
        document.getElementById('auth-check').style.display = 'block';
        document.getElementById('app-content').style.display = 'none';
        return;
    }
    
    if (role !== 'manager') {
        showFeedback('error', 'You need manager privileges to access this page');
        setTimeout(() => {
            window.location.href = '/';
        }, 2000);
        return;
    }
    
    document.getElementById('auth-check').style.display = 'none';
    document.getElementById('app-content').style.display = 'block';
    
    // Display user info
    const username = sessionStorage.getItem('username');
    document.getElementById('user-info').innerHTML = `
        <div class="user-info-details">
            <span class="username">${username}</span>
            <span class="user-role manager">Manager</span>
        </div>
        <button onclick="logoutUser()">Logout</button>
    `;
    document.getElementById('user-info').style.display = 'flex';
    
    // Initialize tabs
    const tabButtons = document.querySelectorAll('.tab-button');
    const tabContents = document.querySelectorAll('.tab-content');
    
    tabButtons.forEach(button => {
        button.addEventListener('click', () => {
            // Remove active class from all tabs
            tabButtons.forEach(btn => btn.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Add active class to clicked tab
            button.classList.add('active');
            const tabContent = document.getElementById(button.dataset.tab);
            tabContent.classList.add('active');
            
            // Load data for the active tab
            if (button.dataset.tab === 'stats-tab') {
                loadStatistics();
            } else if (button.dataset.tab === 'transactions-tab') {
                loadTransactions();
            } else if (button.dataset.tab === 'loans-tab') {
                loadLoans();
            }
        });
    });
    
    // Initial data load
    loadStatistics();
    
    // Setup event listeners for filter buttons
    document.getElementById('apply-filters').addEventListener('click', loadTransactions);
    document.getElementById('refresh-loans').addEventListener('click', loadLoans);
    
    // Setup event listeners for loan status filter
    document.getElementById('loan-status-filter').addEventListener('change', loadLoans);
    
    // Setup modal close buttons
    document.querySelectorAll('.close-modal').forEach(button => {
        button.addEventListener('click', () => {
            document.querySelectorAll('.modal').forEach(modal => {
                modal.style.display = 'none';
            });
        });
    });
    
    document.getElementById('cancel-modal-close').addEventListener('click', () => {
        document.getElementById('cancel-transaction-modal').style.display = 'none';
    });
    
    document.getElementById('loan-modal-close').addEventListener('click', () => {
        document.getElementById('review-loan-modal').style.display = 'none';
    });
    
    // Setup transaction cancellation
    document.getElementById('confirm-cancel-transaction').addEventListener('click', cancelTransaction);
    
    // Setup loan review actions
    document.getElementById('approve-loan').addEventListener('click', () => reviewLoan('approve'));
    document.getElementById('reject-loan').addEventListener('click', () => reviewLoan('reject'));
});

// Load transaction statistics (same as operator)
async function loadStatistics() {
    try {
        const response = await fetch('/manager/transactions/statistics', {
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
            }
        });
        
        if (!response.ok) throw new Error('Failed to load statistics');
        
        const data = await response.json();
        
        document.getElementById('total-transactions').textContent = data.total_transactions.toLocaleString();
        document.getElementById('total-amount').textContent = data.total_amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
        document.getElementById('active-users').textContent = data.active_users.toLocaleString();
        document.getElementById('avg-transaction').textContent = data.avg_transaction.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    } catch (error) {
        console.error('Error loading statistics:', error);
        showFeedback('error', `Failed to load statistics: ${error.message}`);
    }
}

// Load transactions with filters (same as operator)
async function loadTransactions() {
    const usernameFilter = document.getElementById('username-filter').value;
    const typeFilter = document.getElementById('type-filter').value;
    const dateFilter = document.getElementById('date-filter').value;
    
    let url = '/manager/transactions?';
    if (usernameFilter) url += `username=${encodeURIComponent(usernameFilter)}&`;
    if (typeFilter) url += `type=${encodeURIComponent(typeFilter)}&`;
    if (dateFilter) url += `date=${encodeURIComponent(dateFilter)}&`;
    
    try {
        document.getElementById('transactions-list').innerHTML = `
            <tr>
                <td colspan="6" class="loading-row">Loading transactions...</td>
            </tr>
        `;
        
        const response = await fetch(url, {
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
            }
        });
        
        if (!response.ok) throw new Error('Failed to load transactions');
        
        const data = await response.json();
        const transactions = data.transactions || [];
        
        if (transactions.length === 0) {
            document.getElementById('transactions-list').innerHTML = `
                <tr>
                    <td colspan="6" class="empty-row">No transactions found</td>
                </tr>
            `;
            return;
        }
        
        let html = '';
        transactions.forEach(tx => {
            const date = new Date(tx.timestamp);
            html += `
                <tr>
                    <td>${tx.id}</td>
                    <td>${tx.username || `User #${tx.user_id}`}</td>
                    <td><span class="badge ${tx.type}">${tx.type}</span></td>
                    <td>${tx.amount ? `${tx.amount.toFixed(2)}` : '-'}</td>
                    <td>${date.toLocaleString()}</td>
                    <td>
                        ${tx.can_cancel ? `
                            <button class="action-btn small danger" onclick="showCancelModal(${tx.id})">
                                <i class="fas fa-ban"></i> Cancel
                            </button>
                        ` : `
                            <span class="action-disabled">
                                <i class="fas fa-ban"></i> Unavailable
                            </span>
                        `}
                    </td>
                </tr>
            `;
        });
        
        document.getElementById('transactions-list').innerHTML = html;
    } catch (error) {
        console.error('Error loading transactions:', error);
        document.getElementById('transactions-list').innerHTML = `
            <tr>
                <td colspan="6" class="error-row">Error: ${error.message}</td>
            </tr>
        `;
    }
}

// Show transaction cancellation confirmation modal
function showCancelModal(transactionId) {
    document.getElementById('cancel-transaction-id').textContent = transactionId;
    document.getElementById('cancel-transaction-modal').style.display = 'block';
}

// Cancel transaction (same as operator)
async function cancelTransaction() {
    const transactionId = document.getElementById('cancel-transaction-id').textContent;
    
    try {
        const response = await fetch('/manager/transactions/cancel', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ transaction_id: parseInt(transactionId) })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            document.getElementById('cancel-transaction-modal').style.display = 'none';
            showFeedback('success', 'Transaction cancelled successfully');
            loadTransactions();
        } else {
            showFeedback('error', data.error || 'Failed to cancel transaction');
        }
    } catch (error) {
        console.error('Error cancelling transaction:', error);
        showFeedback('error', `Failed to cancel transaction: ${error.message}`);
    }
}

// Load loans based on selected status
async function loadLoans() {
    const statusFilter = document.getElementById('loan-status-filter').value;
    
    try {
        document.getElementById('loans-list').innerHTML = `
            <tr>
                <td colspan="9" class="loading-row">Loading loans...</td>
            </tr>
        `;
        
        const response = await fetch(`/manager/loans/pending?status=${statusFilter}`, {
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
            }
        });
        
        if (!response.ok) throw new Error('Failed to load loans');
        
        const data = await response.json();
        const loans = data.loans || [];
        
        if (loans.length === 0) {
            document.getElementById('loans-list').innerHTML = `
                <tr>
                    <td colspan="9" class="empty-row">No ${statusFilter} loans found</td>
                </tr>
            `;
            return;
        }
        
        let html = '';
        loans.forEach(loan => {
            const createdDate = new Date(loan.created_at).toLocaleDateString();
            const loanType = loan.type === 'standard' ? 'Standard Loan' : 'Installment Plan';
            
            html += `
                <tr>
                    <td>${loan.id}</td>
                    <td>${loan.username || `User #${loan.user_id}`}</td>
                    <td>${loanType}</td>
                    <td>${loan.amount.toFixed(2)}</td>
                    <td>${loan.term_months} months</td>
                    <td>${loan.interest_rate}%</td>
                    <td>${loan.total_payable.toFixed(2)}</td>
                    <td>${createdDate}</td>
                    <td>
                        ${statusFilter === 'pending' ? `
                            <button class="action-btn small primary" onclick="showLoanReviewModal(${JSON.stringify(loan).replace(/"/g, '&quot;')})">
                                <i class="fas fa-search"></i> Review
                            </button>
                        ` : `
                            <button class="action-btn small secondary" onclick="viewLoanDetails(${loan.id})">
                                <i class="fas fa-eye"></i> Details
                            </button>
                        `}
                    </td>
                </tr>
            `;
        });
        
        document.getElementById('loans-list').innerHTML = html;
    } catch (error) {
        console.error('Error loading loans:', error);
        document.getElementById('loans-list').innerHTML = `
            <tr>
                <td colspan="9" class="error-row">Error: ${error.message}</td>
            </tr>
        `;
    }
}

// Show loan review modal
function showLoanReviewModal(loan) {
    // If loan is passed as a string (from onclick), parse it
    if (typeof loan === 'string') {
        loan = JSON.parse(loan);
    }
    
    const loanType = loan.type === 'standard' ? 'Standard Loan' : 'Installment Plan';
    
    document.getElementById('review-loan-id').textContent = loan.id;
    document.getElementById('review-loan-user').textContent = loan.username || `User #${loan.user_id}`;
    document.getElementById('review-loan-type').textContent = loanType;
    document.getElementById('review-loan-amount').textContent = `${loan.amount.toFixed(2)}`;
    document.getElementById('review-loan-term').textContent = `${loan.term_months} months`;
    document.getElementById('review-loan-interest').textContent = `${loan.interest_rate}%`;
    document.getElementById('review-loan-monthly').textContent = `${loan.monthly_payment.toFixed(2)}`;
    document.getElementById('review-loan-total').textContent = `${loan.total_payable.toFixed(2)}`;
    document.getElementById('review-loan-comment').value = '';
    
    document.getElementById('review-loan-modal').style.display = 'block';
}

// View loan details (for non-pending loans)
function viewLoanDetails(loanId) {
    // This would typically show a modal with loan details
    // For now, just show a temporary notification
    showFeedback('info', `Viewing details for loan #${loanId}`);
}

// Review loan (approve/reject)
async function reviewLoan(action) {
    const loanId = parseInt(document.getElementById('review-loan-id').textContent);
    const comment = document.getElementById('review-loan-comment').value;
    
    // Require comment for rejection
    if (action === 'reject' && !comment.trim()) {
        showFeedback('error', 'Please provide a reason for rejection');
        return;
    }
    
    // Show processing notification
    const notificationId = showLoanNotification('processing', `Processing loan ${action} request...`);
    
    try {
        const response = await fetch('/manager/loans/review', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                loan_id: loanId,
                action: action,
                comment: comment
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            document.getElementById('review-loan-modal').style.display = 'none';
            // Update notification to success
            updateLoanNotification(notificationId, 'success', `Loan ${action === 'approve' ? 'approved' : 'rejected'} successfully`);
            showFeedback('success', `Loan ${action}d successfully`);
            loadLoans();
        } else {
            // Update notification to error
            updateLoanNotification(notificationId, 'error', data.error || `Failed to ${action} loan`);
            showFeedback('error', data.error || `Failed to ${action} loan`);
        }
    } catch (error) {
        console.error(`Error ${action}ing loan:`, error);
        // Update notification to error
        updateLoanNotification(notificationId, 'error', `Failed to ${action} loan: ${error.message}`);
        showFeedback('error', `Failed to ${action} loan: ${error.message}`);
    }
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
