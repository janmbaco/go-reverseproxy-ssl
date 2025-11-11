// Main JavaScript utilities for Reverse Proxy Config UI

// Global utility functions
window.Utils = {
    // Show a temporary message notification
    showMessage: function(message, type = 'info', duration = 5000) {
        // Remove existing messages
        const existing = document.querySelector('.message-notification');
        if (existing) existing.remove();

        // Create message element
        const msgDiv = document.createElement('div');
        msgDiv.className = `message-notification ${type}`;
        msgDiv.innerHTML = `
            <i class="fas ${this.getIconForType(type)}"></i>
            <span>${message}</span>
            <button onclick="this.parentElement.remove()" class="message-close">&times;</button>
        `;

        // Add to page
        document.body.appendChild(msgDiv);

        // Auto remove after duration
        if (duration > 0) {
            setTimeout(() => {
                if (msgDiv.parentElement) msgDiv.remove();
            }, duration);
        }

        return msgDiv;
    },

    // Get appropriate icon for message type
    getIconForType: function(type) {
        const icons = {
            'success': 'fa-check-circle',
            'error': 'fa-exclamation-circle',
            'warning': 'fa-exclamation-triangle',
            'info': 'fa-info-circle'
        };
        return icons[type] || 'fa-info-circle';
    },

    // Show success message
    showSuccess: function(message, duration) {
        return this.showMessage(message, 'success', duration);
    },

    // Show error message
    showError: function(message, duration) {
        return this.showMessage(message, 'error', duration);
    },

    // Show warning message
    showWarning: function(message, duration) {
        return this.showMessage(message, 'warning', duration);
    },

    // Show info message
    showInfo: function(message, duration) {
        return this.showMessage(message, 'info', duration);
    },

    // Confirm dialog utility
    confirm: function(message, title = 'Confirm') {
        return new Promise((resolve) => {
            // Create modal elements
            const modal = document.createElement('div');
            modal.className = 'modal';
            modal.innerHTML = `
                <div class="modal-content">
                    <div class="modal-header">
                        <h2><i class="fas fa-question-circle"></i> ${title}</h2>
                    </div>
                    <div class="modal-body">
                        <p>${message}</p>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-secondary" onclick="this.closest('.modal').remove(); resolve(false)">
                            <i class="fas fa-times"></i> Cancel
                        </button>
                        <button class="btn btn-primary" onclick="this.closest('.modal').remove(); resolve(true)">
                            <i class="fas fa-check"></i> Confirm
                        </button>
                    </div>
                </div>
            `;

            // Add to page and show
            document.body.appendChild(modal);
            modal.style.display = 'flex';
            modal.offsetHeight; // Trigger reflow
            modal.classList.add('modal-visible');

            // Make resolve function available to buttons
            modal.querySelector('.btn-secondary').onclick = () => {
                modal.remove();
                resolve(false);
            };
            modal.querySelector('.btn-primary').onclick = () => {
                modal.remove();
                resolve(true);
            };
        });
    },

    // Loading state utility
    setLoading: function(element, loading = true, text = 'Loading...') {
        if (loading) {
            element.disabled = true;
            element.dataset.originalText = element.innerHTML;
            element.innerHTML = `<i class="fas fa-spinner fa-spin"></i> ${text}`;
        } else {
            element.disabled = false;
            if (element.dataset.originalText) {
                element.innerHTML = element.dataset.originalText;
            }
        }
    },

    // Form validation utility
    validateForm: function(form) {
        const inputs = form.querySelectorAll('input[required], select[required], textarea[required]');
        let isValid = true;

        inputs.forEach(input => {
            if (!input.value.trim()) {
                input.classList.add('error');
                isValid = false;
            } else {
                input.classList.remove('error');
            }
        });

        return isValid;
    },

    // AJAX utility
    ajax: function(url, options = {}) {
        const defaultOptions = {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        };

        const config = { ...defaultOptions, ...options };

        if (config.body && typeof config.body === 'object') {
            config.body = JSON.stringify(config.body);
        }

        return fetch(url, config)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            });
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    // Add loading class to body initially
    document.body.classList.add('loaded');

    // Global error handler for unhandled promise rejections
    window.addEventListener('unhandledrejection', function(event) {
        console.error('Unhandled promise rejection:', event.reason);
        Utils.showError('An unexpected error occurred. Please try again.');
    });

    // Global error handler for JavaScript errors
    window.addEventListener('error', function(event) {
        console.error('JavaScript error:', event.error);
        Utils.showError('A JavaScript error occurred. Please refresh the page.');
    });

    // Initialize form enhancements (but don't override existing functionality)
    initializeForms();
});

// Initialize form enhancements without interfering with existing modals
function initializeForms() {
    // Only enhance forms that don't have custom JavaScript handlers
    const forms = document.querySelectorAll('form:not(.no-auto-enhance)');
    forms.forEach(form => {
        // Only add submit handler if it doesn't already have one
        if (!form.hasAttribute('data-has-handler')) {
            form.addEventListener('submit', handleFormSubmit);
            form.setAttribute('data-has-handler', 'true');
        }
    });
}

// Handle form submission with loading states
async function handleFormSubmit(e) {
    const form = e.target;

    // Don't interfere with forms that have custom handlers
    if (form.classList.contains('custom-handler')) {
        return;
    }

    e.preventDefault();

    if (!Utils.validateForm(form)) {
        Utils.showError('Please correct the errors in the form.');
        return;
    }

    const submitBtn = form.querySelector('button[type="submit"]');
    if (submitBtn) {
        Utils.setLoading(submitBtn, true, 'Saving...');
    }

    try {
        const method = form.method || 'POST';
        const url = form.action || window.location.pathname;

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(Object.fromEntries(new FormData(form)))
        });

        const result = await response.json();

        if (response.ok && result.success) {
            Utils.showSuccess('Configuration saved successfully!');
            setTimeout(() => {
                window.location.href = result.redirect || '/';
            }, 1000);
        } else {
            throw new Error(result.error || 'Save failed');
        }
    } catch (error) {
        console.error('Form submission error:', error);
        Utils.showError('Error saving: ' + error.message);
    } finally {
        if (submitBtn) {
            Utils.setLoading(submitBtn, false);
        }
    }
}