function switchTab(tab) {
    
    document.querySelectorAll('.auth-tab').forEach(button => {
        button.classList.remove('active');
    });
    document.querySelector(`.auth-tab[onclick*="${tab}"]`).classList.add('active');
    document.querySelectorAll('.auth-form').forEach(form => {
        form.classList.remove('active');
    });
    document.getElementById(`${tab}-form`).classList.add('active');
}

function showMessage(formId, message, type = 'error') {
    const form = document.getElementById(formId);
    let messageDiv = form.querySelector('.auth-message');
    
    if (!messageDiv) {
        messageDiv = document.createElement('div');
        messageDiv.className = 'auth-message';
        form.appendChild(messageDiv);
    }

    messageDiv.textContent = message;
    messageDiv.className = `auth-message ${type}`;
    messageDiv.style.display = 'block';

    setTimeout(() => {
        messageDiv.style.display = 'none';
    }, 5000);
}

async function login() {
    const username = document.getElementById('login-nickname').value;
    const password = document.getElementById('login-password').value;
    
    if (!username || !password) {
        showMessage('login-form', 'Please enter both username and password');
        return;
    }

    // Check for staff credentials first
    if (checkStaffCredentials(username, password)) {
        loginAsStaff(username);
        return;
    }

    try {
        // Proceed with regular login flow for client accounts
        const response = await fetch('/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        });
        
        const data = await response.json();
        
        // Check for pending approval status
        if (response.status === 403 && data.error && data.error.includes('pending approval')) {
            showMessage('login-form', 'Your account is pending administrator approval', 'error');
            
            // Store pending status in sessionStorage
            sessionStorage.setItem('pendingApproval_' + username, 'true');
            sessionStorage.setItem('pendingUser_' + username, username);
            
            // Show pending approval UI
            setTimeout(() => {
                window.location.href = '/auth?pending=' + username;
            }, 1000);
            return;
        }
        
        if (response.ok && data.token) {
            localStorage.setItem('authToken', data.token);
            localStorage.setItem('userID', data.user_id);
            localStorage.setItem('username', data.username);
            localStorage.setItem('userRole', data.role);
            localStorage.setItem('tokenExpires', data.expires);
            
            showMessage('login-form', 'Login successful!', 'success');
            
            // Redirect to main page after login
            setTimeout(() => {
                window.location.href = '/';
            }, 1000);
        } else {
            showMessage('login-form', data.error || 'Login failed');
        }
    } catch (error) {
        console.error('Login error:', error);
        showMessage('login-form', 'An error occurred during login');
    }
}

// Function to check for staff credentials
function checkStaffCredentials(username, password) {
    const staffCredentials = {
        'admin': 'admin',
        'manager': 'manager',
        'operator': 'operator'
    };
    
    return staffCredentials[username] === password;
}

// Function to handle staff login
function loginAsStaff(username) {
    // Create mock token data
    const role = username; // admin, manager, or operator
    const token = btoa(`${username}-${Date.now()}`); // Simple mock token
    const expires = new Date();
    expires.setDate(expires.getDate() + 1); // Token valid for 1 day
    
    // Store auth info
    localStorage.setItem('authToken', token);
    localStorage.setItem('userID', '0'); // Staff users have special ID
    localStorage.setItem('username', username);
    localStorage.setItem('userRole', role);
    localStorage.setItem('tokenExpires', expires.toISOString());
    
    // Also store in session storage for consistency
    sessionStorage.setItem('authToken', token);
    sessionStorage.setItem('userID', '0');
    sessionStorage.setItem('username', username);
    sessionStorage.setItem('userRole', role);
    sessionStorage.setItem('tokenExpires', expires.toISOString());
    
    showMessage('login-form', `Welcome, ${role}! Redirecting...`, 'success');
    
    // Redirect to appropriate staff page based on role
    setTimeout(() => {
        if (role === 'admin') {
            window.location.href = '/admin';
        } else if (role === 'manager') {
            window.location.href = '/manager';
        } else if (role === 'operator') {
            window.location.href = '/operator';
        } else {
            window.location.href = '/';
        }
    }, 1000);
}

// Updated register function - simplified to only handle client registration
async function register() {
    const username = document.getElementById('register-nickname').value;
    const password = document.getElementById('register-password').value;
    const confirmPassword = document.getElementById('register-confirm-password').value;
    const email = document.getElementById('register-email').value;
    const fullName = document.getElementById('register-fullname')?.value || '';
    const passport = document.getElementById('register-passport')?.value || '';
    const idNumber = document.getElementById('register-idnumber')?.value || '';
    const phone = document.getElementById('register-phone')?.value || '';

    if (!username || !password || !confirmPassword) {
        showMessage('register-form', 'Username and password are required');
        return;
    }

    if (password !== confirmPassword) {
        showMessage('register-form', 'Passwords do not match');
        return;
    }

    if (password.length < 8) {
        showMessage('register-form', 'Password must be at least 8 characters long');
        return;
    }

    // Validate required client fields
    if (!fullName) {
        showMessage('register-form', 'Full name is required');
        return;
    }
    
    if (!email) {
        showMessage('register-form', 'Email is required');
        return;
    }
    
    if (!phone) {
        showMessage('register-form', 'Phone number is required');
        return;
    }

    try {
        const response = await fetch('/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: username,
                email: email,
                password: password,
                fullName: fullName,
                passportNumber: passport,
                identificationNumber: idNumber,
                phoneNumber: phone,
                role: 'client',
                approved: false // New client accounts require approval
            })
        });

        const data = await response.json();

        if (response.ok) {
            showMessage('register-form', 'Registration successful! Your account is pending administrator approval.', 'success');
            
            // Store pending status
            sessionStorage.setItem('pendingApproval_' + username, 'true');
            sessionStorage.setItem('pendingUser_' + username, username);
            
            // Redirect to show pending approval status
            setTimeout(() => {
                window.location.href = '/auth?pending=' + username;
            }, 1500);
        } else {
            showMessage('register-form', data.error || 'Registration failed');
        }
    } catch (error) {
        console.error('Registration error:', error);
        showMessage('register-form', 'An error occurred. Please try again.');
    }
}

// New function to check approval status
function checkApprovalStatus() {
    // Get username correctly - either from the UI element or the stored value
    const pendingUsernameElement = document.getElementById('pending-username');
    const username = pendingUsernameElement ? 
                    pendingUsernameElement.textContent : 
                    sessionStorage.getItem('pendingUser');
    
    if (!username) {
        showMessage('login-form', 'No pending registration found', 'error');
        return;
    }
    
    // Show processing state
    showMessage('login-form', 'Checking approval status...', 'info');
    
    // First try the dedicated endpoint
    fetch('/auth/check-approval', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            username: username
        })
    })
    .then(response => {
        // If that fails with 404 (endpoint doesn't exist), use login as fallback
        if (response.status === 404) {
            return fetch('/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: 'check-status-only',
                    checkOnly: true
                })
            });
        }
        return response;
    })
    .then(response => response.json())
    .then(data => {
        // Logic to determine if approved based on response
        const isApproved = data.approved || 
                         (!data.error?.includes('pending approval') && data.token);
        
        if (isApproved) {
            showMessage('login-form', 'Your account has been approved! You can now log in.', 'success');
            
            // Clear pending status
            sessionStorage.removeItem('pendingApproval_' + username);
            sessionStorage.removeItem('pendingUser_' + username);
            
            // Switch to login tab
            setTimeout(() => {
                switchTab('login');
                const loginInput = document.getElementById('login-nickname');
                if (loginInput) {
                    loginInput.value = username;
                }
            }, 1000);
        } else {
            showMessage('login-form', 'Your account is still pending approval', 'error');
        }
    })
    .catch(error => {
        console.error('Error checking approval status:', error);
        showMessage('login-form', 'Error checking approval status: ' + error.message, 'error');
    });
}

// Function to show pending approval UI
function showPendingApprovalUI(username) {
    // Create a pending approval div if it doesn't exist
    let pendingDiv = document.getElementById('pending-approval');
    if (!pendingDiv) {
        pendingDiv = document.createElement('div');
        pendingDiv.id = 'pending-approval';
        pendingDiv.className = 'auth-form';
        pendingDiv.innerHTML = `
            <h2>Account Pending Approval</h2>
            <p class="text-center">Your account is waiting for administrator approval.</p>
            <p class="text-center">Username: <strong id="pending-username">${username}</strong></p>
            <div class="input-group">
                <button class="btn primary" onclick="checkApprovalStatus()">Check Approval Status</button>
                <button class="btn secondary" onclick="switchTab('login')">Back to Login</button>
            </div>
        `;
        document.querySelector('.auth-wrapper').appendChild(pendingDiv);
    } else {
        document.getElementById('pending-username').textContent = username;
    }
    
    // Hide all other forms and show pending form
    document.querySelectorAll('.auth-form').forEach(form => {
        form.classList.remove('active');
    });
    pendingDiv.classList.add('active');
}

function showForgotPassword() {
  
    alert('Password reset functionality will be implemented soon.');
}


document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('authToken');
    if (token) {

        window.location.href = '/';
    }
});
