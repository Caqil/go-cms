class AuthManager {
    constructor() {
        this.apiBase = '/api/v1';
        this.token = localStorage.getItem('authToken');
        this.init();
    }

    init() {
        // Check if user is already logged in and redirect to dashboard
        if (this.token && window.location.pathname.includes('login.html')) {
            this.validateTokenAndRedirect();
        }

        this.setupEventListeners();
    }

    setupEventListeners() {
        const loginForm = document.getElementById('loginForm');
        if (loginForm) {
            loginForm.addEventListener('submit', (e) => this.handleLogin(e));
        }

        const registerForm = document.getElementById('registerForm');
        if (registerForm) {
            registerForm.addEventListener('submit', (e) => this.handleRegister(e));
        }

        const togglePassword = document.getElementById('togglePassword');
        if (togglePassword) {
            togglePassword.addEventListener('click', this.togglePasswordVisibility);
        }
    }

    async handleLogin(event) {
        event.preventDefault();
        
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const rememberMe = document.getElementById('remember-me').checked;

        this.setLoading(true);

        try {
            const response = await fetch(`${this.apiBase}/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password }),
            });

            const data = await response.json();

            if (response.ok) {
                // Store token
                localStorage.setItem('authToken', data.tokens.access_token);
                if (rememberMe) {
                    localStorage.setItem('refreshToken', data.tokens.refresh_token);
                }
                
                // Store user info
                localStorage.setItem('user', JSON.stringify(data.user));

                this.showAlert('Login successful! Redirecting...', 'success');
                
                // Redirect to dashboard
                setTimeout(() => {
                    window.location.href = 'index.html';
                }, 1000);
            } else {
                this.showAlert(data.error || 'Login failed', 'error');
            }
        } catch (error) {
            this.showAlert('Network error. Please try again.', 'error');
            console.error('Login error:', error);
        } finally {
            this.setLoading(false);
        }
    }

    async handleRegister(event) {
        event.preventDefault();
        
        const username = document.getElementById('username').value;
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const confirmPassword = document.getElementById('confirmPassword').value;

        if (password !== confirmPassword) {
            this.showAlert('Passwords do not match', 'error');
            return;
        }

        this.setLoading(true);

        try {
            const response = await fetch(`${this.apiBase}/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, email, password }),
            });

            const data = await response.json();

            if (response.ok) {
                this.showAlert('Registration successful! Please log in.', 'success');
                setTimeout(() => {
                    window.location.href = 'login.html';
                }, 2000);
            } else {
                this.showAlert(data.error || 'Registration failed', 'error');
            }
        } catch (error) {
            this.showAlert('Network error. Please try again.', 'error');
            console.error('Registration error:', error);
        } finally {
            this.setLoading(false);
        }
    }

    async validateTokenAndRedirect() {
        try {
            const response = await fetch(`${this.apiBase}/profile`, {
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                },
            });

            if (response.ok) {
                window.location.href = 'index.html';
            } else {
                localStorage.removeItem('authToken');
                localStorage.removeItem('refreshToken');
                localStorage.removeItem('user');
            }
        } catch (error) {
            console.error('Token validation error:', error);
            localStorage.removeItem('authToken');
        }
    }

    togglePasswordVisibility() {
        const passwordInput = document.getElementById('password');
        const toggleButton = document.getElementById('togglePassword');
        
        if (passwordInput.type === 'password') {
            passwordInput.type = 'text';
            toggleButton.innerHTML = `
                <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21"></path>
                </svg>
            `;
        } else {
            passwordInput.type = 'password';
            toggleButton.innerHTML = `
                <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
            `;
        }
    }

    setLoading(loading) {
        const loginBtn = document.getElementById('loginBtn');
        const loginText = document.getElementById('loginText');
        
        if (loginBtn && loginText) {
            if (loading) {
                loginBtn.disabled = true;
                loginText.textContent = 'Signing in...';
                loginBtn.classList.add('opacity-50', 'cursor-not-allowed');
            } else {
                loginBtn.disabled = false;
                loginText.textContent = 'Sign in';
                loginBtn.classList.remove('opacity-50', 'cursor-not-allowed');
            }
        }
    }

    showAlert(message, type) {
        const alertMessage = document.getElementById('alertMessage');
        const alertText = document.getElementById('alertText');
        
        if (alertMessage && alertText) {
            alertText.textContent = message;
            alertMessage.className = `mt-4 p-4 rounded-md ${
                type === 'success' 
                    ? 'bg-green-50 text-green-800 border border-green-200' 
                    : 'bg-red-50 text-red-800 border border-red-200'
            }`;
            alertMessage.classList.remove('hidden');

            setTimeout(() => {
                alertMessage.classList.add('hidden');
            }, 5000);
        }
    }

    logout() {
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('user');
        window.location.href = 'login.html';
    }
}

// Initialize auth manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new AuthManager();
});