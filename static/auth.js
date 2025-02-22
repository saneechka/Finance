function switchTab(tab) {
    // Update tab buttons
    document.querySelectorAll('.auth-tab').forEach(button => {
        button.classList.remove('active');
    });
    document.querySelector(`.auth-tab[onclick*="${tab}"]`).classList.add('active');

    // Update form visibility
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
    const nickname = document.getElementById('login-nickname').value;
    const password = document.getElementById('login-password').value;

    // Skip validation and always treat as successful
    showMessage('login-form', 'Login successful!', 'success');
    
    // Set a dummy token
    localStorage.setItem('token', 'dummy_token');
    
    // Redirect to requests page
    setTimeout(() => {
        window.location.href = '/api/deposit/requests';
    }, 1000);
}

async function register() {
    const nickname = document.getElementById('register-nickname').value;
    const role = document.getElementById('register-role').value;
    const password = document.getElementById('register-password').value;
    const confirmPassword = document.getElementById('register-confirm-password').value;

    // Validation
    if (!nickname || !role || !password || !confirmPassword) {
        showMessage('register-form', 'Please fill in all fields');
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

    try {
        const response = await fetch('/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                nickname,
                role,
                password
            })
        });

        const data = await response.json();

        if (response.ok) {
            showMessage('register-form', 'Registration successful!', 'success');
            setTimeout(() => {
                switchTab('login');
            }, 1000);
        } else {
            showMessage('register-form', data.error || 'Registration failed');
        }
    } catch (error) {
        showMessage('register-form', 'An error occurred. Please try again.');
    }
}

function showForgotPassword() {
    // Implement forgot password functionality
    alert('Password reset functionality will be implemented soon.');
}
