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
