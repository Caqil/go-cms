function adminDashboard() {
    return {
        // State
        user: JSON.parse(localStorage.getItem('user') || '{}'),
        stats: {},
        recentActivity: [],
        mobileMenuOpen: false,
        pluginChart: null,

        // Notification
        notification: {
            show: false,
            type: 'success',
            title: '',
            message: ''
        },

        // Computed properties
        get userInitials() {
            if (!this.user.username) return 'U';
            return this.user.username.charAt(0).toUpperCase();
        },

        // Initialize
        async init() {
            this.checkAuth();
            await this.loadDashboardData();
            await this.loadNavigation();
            this.setupCharts();
            this.setupEventListeners();
        },

        // Authentication
        checkAuth() {
            const token = localStorage.getItem('authToken');
            if (!token) {
                window.location.href = 'login.html';
                return;
            }
        },

        getHeaders() {
            return {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
                'Content-Type': 'application/json'
            };
        },

        // API Methods
        async apiCall(endpoint, options = {}) {
            try {
                const response = await fetch(`/api/v1${endpoint}`, {
                    headers: this.getHeaders(),
                    ...options
                });

                if (!response.ok) {
                    if (response.status === 401) {
                        localStorage.removeItem('authToken');
                        localStorage.removeItem('user');
                        window.location.href = 'login.html';
                        return;
                    }
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }

                return await response.json();
            } catch (error) {
                console.error('API call failed:', error);
                this.showNotification('error', 'API Error', error.message);
                throw error;
            }
        },

        // Load Dashboard Data
        async loadDashboardData() {
            try {
                const response = await this.apiCall('/admin/dashboard');
                this.stats = response.stats || {};
                this.recentActivity = response.activity || [];
                
                // Update charts after data loads
                this.$nextTick(() => {
                    this.updatePluginChart();
                });
            } catch (error) {
                console.error('Failed to load dashboard data:', error);
                this.showNotification('error', 'Load Error', 'Failed to load dashboard data');
            }
        },

        // Navigation
        async loadNavigation() {
            try {
                const response = await this.apiCall('/admin/menu');
                this.renderNavigation(response.menu);
            } catch (error) {
                console.error('Failed to load navigation:', error);
            }
        },

        renderNavigation(menuItems) {
            const sidebarNav = document.getElementById('sidebar-nav');
            const mobileNav = document.getElementById('mobile-nav');
            
            if (!sidebarNav) return;

            const navHTML = menuItems.map(item => this.createNavItem(item)).join('');
            sidebarNav.innerHTML = navHTML;
            
            if (mobileNav) {
                mobileNav.innerHTML = navHTML;
            }
        },

        createNavItem(item) {
            const hasChildren = item.children && item.children.length > 0;
            const isActive = this.isActiveRoute(item.url);
            
            let childrenHTML = '';
            if (hasChildren) {
                childrenHTML = `
                    <div class="ml-4 mt-1 space-y-1">
                        ${item.children.map(child => `
                            <a href="${child.url}" 
                               class="group flex items-center px-2 py-2 text-xs font-medium rounded-md ${this.isActiveRoute(child.url) ? 'bg-gray-700 text-white' : 'text-gray-300 hover:text-white hover:bg-gray-700'}">
                                ${child.title}
                            </a>
                        `).join('')}
                    </div>
                `;
            }

            return `
                <div class="nav-item">
                    <a href="${item.url}" 
                       class="nav-link ${isActive ? 'bg-gray-900 text-white' : 'text-gray-300 hover:bg-gray-700 hover:text-white'} group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                        ${this.getNavIcon(item.icon)}
                        ${item.title}
                    </a>
                    ${childrenHTML}
                </div>
            `;
        },

        getNavIcon(iconName) {
            const icons = {
                'dashboard': `<svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                </svg>`,
                'plugins': `<svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
                </svg>`,
                'palette': `<svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h4a2 2 0 012-2v-1a2 2 0 00-2-2H7m0-3.5a.5.5 0 01.5-.5h1a.5.5 0 01.5.5v1a.5.5 0 01-.5.5h-1a.5.5 0 01-.5-.5v-1z"></path>
                </svg>`,
                'users': `<svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"></path>
                </svg>`,
                'cog': `<svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                </svg>`,
                'default': `<svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>`
            };
            
            return icons[iconName] || icons.default;
        },

        isActiveRoute(url) {
            const currentPath = window.location.pathname;
            if (url === '/admin' || url === '/admin/') {
                return currentPath === '/admin/' || currentPath === '/admin/index.html' || currentPath.endsWith('/admin');
            }
            return currentPath.includes(url.replace('/admin/', ''));
        },

        // Charts
        setupCharts() {
            this.setupPluginChart();
        },

        setupPluginChart() {
            const ctx = document.getElementById('pluginChart');
            if (!ctx) return;

            this.pluginChart = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: ['Active', 'Inactive', 'Loaded', 'Unloaded'],
                    datasets: [{
                        data: [0, 0, 0, 0],
                        backgroundColor: [
                            '#10B981', // green-500
                            '#EF4444', // red-500
                            '#3B82F6', // blue-500
                            '#6B7280'  // gray-500
                        ],
                        borderWidth: 2,
                        borderColor: '#FFFFFF'
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom',
                            labels: {
                                padding: 20,
                                usePointStyle: true
                            }
                        }
                    },
                    cutout: '60%'
                }
            });
        },

        updatePluginChart() {
            if (!this.pluginChart) return;

            const active = this.stats.active_plugins || 0;
            const total = this.stats.total_plugins || 0;
            const inactive = total - active;

            this.pluginChart.data.datasets[0].data = [active, inactive];
            this.pluginChart.data.labels = ['Active', 'Inactive'];
            this.pluginChart.update();
        },

        // Event Listeners
        setupEventListeners() {
            // Search functionality
            const searchInput = document.getElementById('search-field');
            if (searchInput) {
                searchInput.addEventListener('input', this.handleSearch.bind(this));
            }

            // Auto-refresh dashboard every 5 minutes
            setInterval(() => {
                this.refreshDashboard();
            }, 5 * 60 * 1000);
        },

        handleSearch(event) {
            const query = event.target.value.toLowerCase();
            
            // Simple search implementation
            // In a real app, you'd implement more sophisticated search
            console.log('Searching for:', query);
            
            // Could filter dashboard content, navigate to search results, etc.
            if (query.includes('plugin')) {
                this.showNotification('info', 'Search Tip', 'Visit the Plugins page for plugin management');
            }
        },

        // Actions
        async refreshDashboard() {
            try {
                await this.loadDashboardData();
                this.showNotification('success', 'Refreshed', 'Dashboard data updated');
            } catch (error) {
                console.error('Failed to refresh dashboard:', error);
            }
        },

        logout() {
            localStorage.removeItem('authToken');
            localStorage.removeItem('user');
            window.location.href = 'login.html';
        },

        // Utility Functions
        formatDate(dateString) {
            if (!dateString) return 'Unknown';
            const date = new Date(dateString);
            const now = new Date();
            const diffMs = now - date;
            const diffMins = Math.floor(diffMs / 60000);
            const diffHours = Math.floor(diffMs / 3600000);
            const diffDays = Math.floor(diffMs / 86400000);

            if (diffMins < 1) return 'Just now';
            if (diffMins < 60) return `${diffMins}m ago`;
            if (diffHours < 24) return `${diffHours}h ago`;
            if (diffDays < 7) return `${diffDays}d ago`;
            
            return date.toLocaleDateString();
        },

        getActivityIcon(type) {
            const iconClasses = {
                'plugin': 'bg-blue-500',
                'user': 'bg-green-500',
                'theme': 'bg-purple-500',
                'system': 'bg-yellow-500',
                'error': 'bg-red-500',
                'default': 'bg-gray-500'
            };
            
            return iconClasses[type] || iconClasses.default;
        },

        showNotification(type, title, message) {
            this.notification = {
                show: true,
                type,
                title,
                message
            };

            // Auto-hide after 5 seconds
            setTimeout(() => {
                this.notification.show = false;
            }, 5000);
        },

        // Quick Actions
        navigateToPlugins() {
            window.location.href = 'plugins.html';
        },

        navigateToThemes() {
            window.location.href = 'themes.html';
        },

        navigateToUsers() {
            window.location.href = 'users.html';
        },

        navigateToSettings() {
            window.location.href = 'settings.html';
        },

        // System Status
        async checkSystemStatus() {
            try {
                const response = await this.apiCall('/admin/system/info');
                return {
                    online: true,
                    ...response
                };
            } catch (error) {
                return {
                    online: false,
                    error: error.message
                };
            }
        },

        // Plugin Quick Actions
        async quickPluginAction(action) {
            try {
                switch (action) {
                    case 'hot-reload':
                        await this.apiCall('/admin/system/hot-reload', { method: 'POST' });
                        this.showNotification('success', 'Success', 'All plugins reloaded');
                        break;
                    case 'cleanup':
                        await this.apiCall('/admin/system/cleanup-cache', { method: 'POST' });
                        this.showNotification('success', 'Success', 'Cache cleaned up');
                        break;
                    default:
                        console.warn('Unknown action:', action);
                }
            } catch (error) {
                console.error('Quick action failed:', error);
            }
        },

        // Dashboard Widgets
        async loadRecentPlugins() {
            try {
                const response = await this.apiCall('/admin/plugins');
                const plugins = response.plugins || [];
                
                // Get recently uploaded plugins
                return plugins
                    .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
                    .slice(0, 5);
            } catch (error) {
                console.error('Failed to load recent plugins:', error);
                return [];
            }
        },

        // Export functions
        exportDashboardData() {
            const data = {
                stats: this.stats,
                recent_activity: this.recentActivity,
                user: this.user,
                exported_at: new Date().toISOString()
            };

            const blob = new Blob([JSON.stringify(data, null, 2)], { 
                type: 'application/json' 
            });
            
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `dashboard-data-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            this.showNotification('success', 'Export Complete', 'Dashboard data downloaded');
        }
    };
}

// Global utilities for the admin interface
window.adminUtils = {
    // Format file sizes
    formatFileSize(bytes) {
        if (!bytes) return '0 B';
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    },

    // Format numbers with commas
    formatNumber(num) {
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    },

    // Validate email addresses
    isValidEmail(email) {
        const pattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return pattern.test(email);
    },

    // Generate secure passwords
    generatePassword(length = 12) {
        const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*';
        let password = '';
        for (let i = 0; i < length; i++) {
            password += charset.charAt(Math.floor(Math.random() * charset.length));
        }
        return password;
    },

    // Copy text to clipboard
    async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (error) {
            // Fallback for older browsers
            const textArea = document.createElement('textarea');
            textArea.value = text;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
            return true;
        }
    },

    // Debounce function for search inputs
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },

    // Theme switcher
    toggleTheme() {
        const html = document.documentElement;
        const isDark = html.classList.contains('dark');
        
        if (isDark) {
            html.classList.remove('dark');
            localStorage.setItem('theme', 'light');
        } else {
            html.classList.add('dark');
            localStorage.setItem('theme', 'dark');
        }
    },

    // Initialize theme from localStorage
    initTheme() {
        const savedTheme = localStorage.getItem('theme');
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        
        if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
            document.documentElement.classList.add('dark');
        }
    }
};

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', function() {
    window.adminUtils.initTheme();
    
    // Initialize Alpine.js if available
    if (typeof Alpine !== 'undefined') {
        Alpine.start();
    }
});

// Handle offline/online status
window.addEventListener('online', () => {
    console.log('Connection restored');
    // Could show a notification or retry failed requests
});

window.addEventListener('offline', () => {
    console.log('Connection lost');
    // Could show an offline indicator
});