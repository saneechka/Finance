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

    try {
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

async function register() {
    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const username = document.getElementById('register-nickname').value;
    const role = document.getElementById('register-role').value;
    const password = document.getElementById('register-password').value;
    const confirmPassword = document.getElementById('register-confirm-password').value;

   
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
                fullName: name,
                role: role || 'client' 
            })
        });

        const data = await response.json();

        if (response.ok) {
            showMessage('register-form', 'Registration successful!', 'success');
            setTimeout(() => {
                switchTab('login');
            }, 1500);
        } else {
            showMessage('register-form', data.error || 'Registration failed');
        }
    } catch (error) {
        console.error('Registration error:', error);
        showMessage('register-form', 'An error occurred. Please try again.');
    }
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
