class AdminDashboard {
    constructor() {
        this.apiBase = '/api/v1';
        this.token = localStorage.getItem('authToken');
        this.user = JSON.parse(localStorage.getItem('user') || '{}');
        
        if (!this.token) {
            window.location.href = 'login.html';
            return;
        }

        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.loadNavigation();
        await this.loadDashboardData();
        this.setupCharts();
    }

    setupEventListeners() {
        // Logout button
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', this.logout.bind(this));
        }

        // Mobile menu toggle
        const openMobileMenu = document.getElementById('open-mobile-menu');
        const closeMobileMenu = document.getElementById('close-mobile-menu');
        const mobileMenu = document.getElementById('mobile-menu');

        if (openMobileMenu && mobileMenu) {
            openMobileMenu.addEventListener('click', () => {
                mobileMenu.classList.remove('hidden');
            });
        }

        if (closeMobileMenu && mobileMenu) {
            closeMobileMenu.addEventListener('click', () => {
                mobileMenu.classList.add('hidden');
            });
        }

        // Search functionality
        const searchInput = document.getElementById('search');
        if (searchInput) {
            searchInput.addEventListener('input', this.handleSearch.bind(this));
        }
    }

    async loadNavigation() {
        try {
            const response = await this.apiCall('/admin/menu');
            const menuItems = response.menu;
            
            this.renderNavigation(menuItems);
            this.updateUserInfo();
        } catch (error) {
            console.error('Failed to load navigation:', error);
        }
    }

    renderNavigation(menuItems) {
        const sidebarNav = document.getElementById('sidebar-nav');
        const mobileNav = document.getElementById('mobile-nav');
        
        if (!sidebarNav) return;

        const navHTML = menuItems.map(item => this.createNavItem(item)).join('');
        sidebarNav.innerHTML = navHTML;
        
        if (mobileNav) {
            mobileNav.innerHTML = navHTML;
        }
    }

    createNavItem(item) {
        const hasChildren = item.children && item.children.length > 0;
        const isActive = window.location.pathname.includes(item.url.replace('/admin/', ''));
        
        return `
            <div class="nav-item">
                <a href="${item.url}" 
                   class="nav-link ${isActive ? 'bg-gray-900 text-white' : 'text-gray-300 hover:bg-gray-700 hover:text-white'} group flex items-center px-2 py-2 text-sm font-medium rounded-md"
                   ${hasChildren ? 'onclick="return false;"' : ''}>
                    <svg class="mr-3 flex-shrink-0 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        ${this.getIconPath(item.icon)}
                    </svg>
                    ${item.title}
                    ${hasChildren ? `
                        <svg class="ml-auto h-5 w-5 transform group-hover:text-gray-300 transition-colors ease-in-out duration-150" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                        </svg>
                    ` : ''}
                </a>
                ${hasChildren ? `
                    <div class="submenu mt-1 space-y-1">
                        ${item.children.map(child => `
                            <a href="${child.url}" 
                               class="group flex items-center pl-11 pr-2 py-2 text-sm font-medium text-gray-400 rounded-md hover:text-gray-300 hover:bg-gray-700">
                                ${child.title}
                            </a>
                        `).join('')}
                    </div>
                ` : ''}
            </div>
        `;
    }

    getIconPath(iconName) {
        const icons = {
            'tachometer-alt': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>',
            'puzzle-piece': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z"></path>',
            'palette': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17v4a2 2 0 002 2h4M13 13h4a2 2 0 012 2v4a2 2 0 01-2 2H9a2 2 0 01-2-2v-4a2 2 0 012-2h4z"></path>',
            'users': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"></path>',
            'cog': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>',
            default: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>'
        };
        
        return icons[iconName] || icons.default;
    }

    updateUserInfo() {
        const userName = document.getElementById('user-name');
        if (userName && this.user.username) {
            userName.textContent = this.user.username;
        }
    }

    async loadDashboardData() {
        try {
            const response = await this.apiCall('/admin/dashboard');
            this.renderStats(response.stats);
            this.renderRecentActivity(response.recent_activity || []);
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.showError('Failed to load dashboard data');
        }
    }

    renderStats(stats) {
        const statsGrid = document.getElementById('stats-grid');
        if (!statsGrid) return;

        const statsCards = [
            {
                name: 'Total Users',
                value: stats.total_users || 0,
                icon: 'users',
                color: 'blue',
                change: '+12%',
                changeType: 'increase'
            },
            {
                name: 'Active Plugins',
                value: stats.active_plugins || 0,
                icon: 'puzzle',
                color: 'green',
                change: '+3',
                changeType: 'increase'
            },
            {
                name: 'Active Theme',
                value: stats.active_theme || 'Default',
                icon: 'palette',
                color: 'purple',
                change: 'Recently changed',
                changeType: 'neutral'
            },
            {
                name: 'System Uptime',
                value: stats.system_uptime || '0h',
                icon: 'clock',
                color: 'yellow',
                change: 'Running',
                changeType: 'neutral'
            }
        ];

        statsGrid.innerHTML = statsCards.map(stat => `
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="p-5">
                    <div class="flex items-center">
                        <div class="flex-shrink-0">
                            <div class="w-8 h-8 bg-${stat.color}-500 rounded-md flex items-center justify-center">
                                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    ${this.getStatIcon(stat.icon)}
                                </svg>
                            </div>
                        </div>
                        <div class="ml-5 w-0 flex-1">
                            <dl>
                                <dt class="text-sm font-medium text-gray-500 truncate">${stat.name}</dt>
                                <dd>
                                    <div class="text-lg font-medium text-gray-900">${stat.value}</div>
                                </dd>
                            </dl>
                        </div>
                    </div>
                </div>
                <div class="bg-gray-50 px-5 py-3">
                    <div class="text-sm">
                        <span class="text-gray-600">${stat.change}</span>
                    </div>
                </div>
            </div>
        `).join('');
    }

    getStatIcon(iconName) {
        const icons = {
            users: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"></path>',
            puzzle: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z"></path>',
            palette: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17v4a2 2 0 002 2h4M13 13h4a2 2 0 012 2v4a2 2 0 01-2 2H9a2 2 0 01-2-2v-4a2 2 0 012-2h4z"></path>',
            clock: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>'
        };
        
        return icons[iconName] || icons.clock;
    }

    renderRecentActivity(activities) {
        const activityList = document.getElementById('activity-list');
        if (!activityList) return;

        if (activities.length === 0) {
            activityList.innerHTML = `
                <li class="text-center py-4 text-gray-500">
                    No recent activity to display.
                </li>
            `;
            return;
        }

        activityList.innerHTML = activities.map((activity, index) => `
            <li class="${index !== activities.length - 1 ? 'pb-8' : ''}">
                <div class="relative">
                    ${index !== activities.length - 1 ? '<span class="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200"></span>' : ''}
                    <div class="relative flex space-x-3">
                        <div>
                            <span class="h-8 w-8 rounded-full bg-${this.getActivityColor(activity.type)}-500 flex items-center justify-center ring-8 ring-white">
                                <svg class="h-4 w-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    ${this.getActivityIcon(activity.type)}
                                </svg>
                            </span>
                        </div>
                        <div class="min-w-0 flex-1 pt-1.5 flex justify-between space-x-4">
                            <div>
                                <p class="text-sm text-gray-500">${activity.description}</p>
                                <p class="mt-0.5 text-xs text-gray-400">by ${activity.user}</p>
                            </div>
                            <div class="text-right text-sm whitespace-nowrap text-gray-500">
                                <time datetime="${activity.timestamp}">${this.formatDate(activity.timestamp)}</time>
                            </div>
                        </div>
                    </div>
                </div>
            </li>
        `).join('');
    }

    getActivityColor(type) {
        const colors = {
            plugin: 'blue',
            theme: 'purple',
            user: 'green',
            system: 'gray',
            error: 'red'
        };
        return colors[type] || 'gray';
    }

    getActivityIcon(type) {
        const icons = {
            plugin: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z"></path>',
            theme: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17v4a2 2 0 002 2h4M13 13h4a2 2 0 012 2v4a2 2 0 01-2 2H9a2 2 0 01-2-2v-4a2 2 0 012-2h4z"></path>',
            user: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>',
            system: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>',
            error: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>'
        };
        return icons[type] || icons.system;
    }

    setupCharts() {
        this.setupAnalyticsChart();
        this.setupPluginChart();
    }

    setupAnalyticsChart() {
        const ctx = document.getElementById('analyticsChart');
        if (!ctx) return;

        new Chart(ctx, {
            type: 'line',
            data: {
                labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
                datasets: [{
                    label: 'Page Views',
                    data: [1200, 1900, 3000, 5000, 2300, 3200],
                    borderColor: 'rgb(59, 130, 246)',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    tension: 0.1
                }, {
                    label: 'Users',
                    data: [300, 450, 600, 800, 500, 700],
                    borderColor: 'rgb(16, 185, 129)',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    setupPluginChart() {
        const ctx = document.getElementById('pluginChart');
        if (!ctx) return;

        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Active', 'Inactive', 'Error'],
                datasets: [{
                    data: [12, 3, 1],
                    backgroundColor: [
                        'rgb(16, 185, 129)',
                        'rgb(107, 114, 128)',
                        'rgb(239, 68, 68)'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    }

    formatDate(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now - date;
        
        if (diff < 60000) return 'Just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
        return `${Math.floor(diff / 86400000)}d ago`;
    }

    handleSearch(event) {
        const query = event.target.value.toLowerCase();
        // Implement search functionality
        console.log('Search query:', query);
    }

    async apiCall(endpoint, options = {}) {
        const response = await fetch(`${this.apiBase}${endpoint}`, {
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });

        if (response.status === 401) {
            this.logout();
            throw new Error('Unauthorized');
        }

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'API call failed');
        }

        return data;
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showNotification(message, type = 'info') {
        const notification = document.getElementById('notification-panel');
        const title = document.getElementById('notification-title');
        const messageEl = document.getElementById('notification-message');
        const icon = document.getElementById('notification-icon');
        
        if (notification && title && messageEl && icon) {
            title.textContent = type === 'error' ? 'Error' : type === 'success' ? 'Success' : 'Information';
            messageEl.textContent = message;
            
            // Update icon based on type
            const iconPaths = {
                success: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>',
                error: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>',
                info: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>'
            };
            
            icon.innerHTML = iconPaths[type] || iconPaths.info;
            icon.className = `h-6 w-6 ${type === 'error' ? 'text-red-400' : type === 'success' ? 'text-green-400' : 'text-blue-400'}`;
            
            notification.classList.remove('hidden');
            
            setTimeout(() => {
                notification.classList.add('hidden');
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

// Utility functions
function hideNotification() {
    const notification = document.getElementById('notification-panel');
    if (notification) {
        notification.classList.add('hidden');
    }
}

function showLoading() {
    const spinner = document.getElementById('loading-spinner');
    if (spinner) {
        spinner.classList.remove('hidden');
    }
}

function hideLoading() {
    const spinner = document.getElementById('loading-spinner');
    if (spinner) {
        spinner.classList.add('hidden');
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new AdminDashboard();
});