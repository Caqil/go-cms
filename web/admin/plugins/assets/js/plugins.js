function pluginManager() {
    return {
        // State
        plugins: [],
        systemInfo: null,
        selectedPlugin: null,
        selectedFile: null,
        uploadProgress: 0,
        isUploading: false,
        
        // Modal states
        showUploadModal: false,
        showDetailsModal: false,
        showSystemInfo: false,
        showSettingsModal: false,
        
        // Notification
        notification: {
            show: false,
            type: 'success',
            title: '',
            message: ''
        },

        // Initialize
        async init() {
            await this.loadPlugins();
            await this.loadSystemInfo();
            this.loadNavigation();
            this.checkAuth();
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

        // Load Data
        async loadPlugins() {
            try {
                const response = await this.apiCall('/admin/plugins');
                this.plugins = response.plugins || [];
                this.systemInfo = response.system_info;
            } catch (error) {
                console.error('Failed to load plugins:', error);
                this.showNotification('error', 'Load Error', 'Failed to load plugins');
            }
        },

        async loadSystemInfo() {
            try {
                const response = await this.apiCall('/admin/system/info');
                this.systemInfo = response;
            } catch (error) {
                console.error('Failed to load system info:', error);
            }
        },

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
            if (!sidebarNav) return;

            const navHTML = menuItems.map(item => this.createNavItem(item)).join('');
            sidebarNav.innerHTML = navHTML;
        },

        createNavItem(item) {
            const hasChildren = item.children && item.children.length > 0;
            const isActive = window.location.pathname.includes(item.url.replace('/admin/', ''));
            
            let childrenHTML = '';
            if (hasChildren) {
                childrenHTML = `
                    <div class="ml-4 mt-1 space-y-1">
                        ${item.children.map(child => `
                            <a href="${child.url}" class="group flex items-center px-2 py-2 text-xs font-medium rounded-md text-gray-300 hover:text-white hover:bg-gray-700">
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
                        <svg class="mr-3 h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
                        </svg>
                        ${item.title}
                    </a>
                    ${childrenHTML}
                </div>
            `;
        },

        // Plugin Actions
        async togglePlugin(pluginName) {
            try {
                const response = await this.apiCall(`/admin/plugins/${pluginName}/toggle`, {
                    method: 'POST'
                });

                this.showNotification('success', 'Success', response.message);
                await this.loadPlugins();
            } catch (error) {
                console.error('Failed to toggle plugin:', error);
            }
        },

        async reloadPlugin(pluginName) {
            try {
                const response = await this.apiCall(`/admin/plugins/${pluginName}/reload`, {
                    method: 'POST'
                });

                this.showNotification('success', 'Success', 'Plugin reloaded successfully');
                await this.loadPlugins();
            } catch (error) {
                console.error('Failed to reload plugin:', error);
            }
        },

        async deletePlugin(pluginName) {
            if (!confirm(`Are you sure you want to delete the plugin "${pluginName}"? This action cannot be undone.`)) {
                return;
            }

            try {
                const response = await this.apiCall(`/admin/plugins/${pluginName}`, {
                    method: 'DELETE'
                });

                this.showNotification('success', 'Success', 'Plugin deleted successfully');
                await this.loadPlugins();
            } catch (error) {
                console.error('Failed to delete plugin:', error);
            }
        },

        async hotReload() {
            try {
                this.showNotification('info', 'Hot Reload', 'Reloading all plugins...');
                
                const response = await this.apiCall('/admin/system/hot-reload', {
                    method: 'POST'
                });

                this.showNotification('success', 'Success', response.message);
                await this.loadPlugins();
            } catch (error) {
                console.error('Failed to hot reload:', error);
            }
        },

        async cleanupCache() {
            try {
                const response = await this.apiCall('/admin/system/cleanup-cache', {
                    method: 'POST'
                });

                this.showNotification('success', 'Success', response.message);
            } catch (error) {
                console.error('Failed to cleanup cache:', error);
            }
        },

        // File Upload
        handleFileSelect(event) {
            const file = event.target.files[0];
            this.setSelectedFile(file);
        },

        handleFileDrop(event) {
            const file = event.dataTransfer.files[0];
            this.setSelectedFile(file);
        },

        setSelectedFile(file) {
            if (!file) return;

            // Validate file type
            if (!file.name.toLowerCase().endsWith('.zip')) {
                this.showNotification('error', 'Invalid File', 'Please select a .zip file');
                return;
            }

            // Validate file size (100MB)
            if (file.size > 100 * 1024 * 1024) {
                this.showNotification('error', 'File Too Large', 'Plugin files must be under 100MB');
                return;
            }

            this.selectedFile = file;
        },

        clearFile() {
            this.selectedFile = null;
            this.uploadProgress = 0;
            if (this.$refs.fileInput) {
                this.$refs.fileInput.value = '';
            }
        },

        async uploadPlugin() {
            if (!this.selectedFile) {
                this.showNotification('error', 'No File', 'Please select a plugin file to upload');
                return;
            }

            this.isUploading = true;
            this.uploadProgress = 0;

            try {
                const formData = new FormData();
                formData.append('plugin', this.selectedFile);

                const xhr = new XMLHttpRequest();

                // Track upload progress
                xhr.upload.addEventListener('progress', (event) => {
                    if (event.lengthComputable) {
                        this.uploadProgress = Math.round((event.loaded / event.total) * 100);
                    }
                });

                // Handle completion
                xhr.addEventListener('load', () => {
                    if (xhr.status === 200) {
                        const response = JSON.parse(xhr.responseText);
                        this.showNotification('success', 'Success', response.message);
                        this.showUploadModal = false;
                        this.clearFile();
                        this.loadPlugins();
                    } else {
                        const error = JSON.parse(xhr.responseText);
                        this.showNotification('error', 'Upload Failed', error.error || 'Unknown error occurred');
                    }
                    this.isUploading = false;
                    this.uploadProgress = 0;
                });

                // Handle errors
                xhr.addEventListener('error', () => {
                    this.showNotification('error', 'Upload Failed', 'Network error occurred');
                    this.isUploading = false;
                    this.uploadProgress = 0;
                });

                // Send the request
                xhr.open('POST', '/api/v1/admin/plugins/upload');
                xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('authToken')}`);
                xhr.send(formData);

            } catch (error) {
                console.error('Upload failed:', error);
                this.showNotification('error', 'Upload Failed', error.message);
                this.isUploading = false;
                this.uploadProgress = 0;
            }
        },

        // Modal Actions
        viewPluginDetails(plugin) {
            this.selectedPlugin = plugin;
            this.showDetailsModal = true;
        },

        async editPluginSettings(plugin) {
            try {
                const response = await this.apiCall(`/admin/plugins/${plugin.name}/settings`);
                this.selectedPlugin = { ...plugin, settings: response.settings };
                this.showSettingsModal = true;
            } catch (error) {
                console.error('Failed to load plugin settings:', error);
            }
        },

        async savePluginSettings() {
            if (!this.selectedPlugin) return;

            try {
                const settings = {};
                this.selectedPlugin.settings.forEach(setting => {
                    settings[setting.key] = setting.value;
                });

                const response = await this.apiCall(`/admin/plugins/${this.selectedPlugin.name}/settings`, {
                    method: 'PUT',
                    body: JSON.stringify(settings)
                });

                this.showNotification('success', 'Success', 'Settings saved successfully');
                this.showSettingsModal = false;
                await this.loadPlugins();
            } catch (error) {
                console.error('Failed to save plugin settings:', error);
            }
        },

        // Utility Functions
        formatDate(dateString) {
            if (!dateString) return 'Unknown';
            const date = new Date(dateString);
            return date.toLocaleDateString();
        },

        formatFileSize(bytes) {
            if (!bytes) return '0 B';
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(1024));
            return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
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

        // Plugin Testing
        async testPlugin(pluginName) {
            try {
                // Test basic plugin endpoints
                const endpoints = [
                    `/plugins/${pluginName.toLowerCase()}/`,
                    `/plugins/${pluginName.toLowerCase()}/info`,
                    `/plugins/${pluginName.toLowerCase()}/status`
                ];

                const results = [];
                for (const endpoint of endpoints) {
                    try {
                        const response = await this.apiCall(endpoint);
                        results.push({ endpoint, status: 'success', data: response });
                    } catch (error) {
                        results.push({ endpoint, status: 'error', error: error.message });
                    }
                }

                console.log('Plugin test results:', results);
                this.showNotification('info', 'Plugin Test', 'Check console for detailed results');

                return results;
            } catch (error) {
                console.error('Plugin test failed:', error);
                this.showNotification('error', 'Test Failed', error.message);
            }
        },

        // Search and Filter
        searchPlugins(query) {
            if (!query) return this.plugins;
            
            query = query.toLowerCase();
            return this.plugins.filter(plugin => 
                plugin.name.toLowerCase().includes(query) ||
                plugin.description?.toLowerCase().includes(query) ||
                plugin.author?.toLowerCase().includes(query)
            );
        },

        filterPluginsByStatus(status) {
            switch (status) {
                case 'active':
                    return this.plugins.filter(p => p.is_active);
                case 'inactive':
                    return this.plugins.filter(p => !p.is_active);
                case 'loaded':
                    return this.plugins.filter(p => p.is_loaded);
                case 'unloaded':
                    return this.plugins.filter(p => !p.is_loaded);
                default:
                    return this.plugins;
            }
        },

        // Export plugin data
        exportPluginData() {
            const data = {
                plugins: this.plugins.map(p => ({
                    name: p.name,
                    version: p.version,
                    description: p.description,
                    author: p.author,
                    is_active: p.is_active,
                    is_loaded: p.is_loaded,
                    created_at: p.created_at
                })),
                system_info: this.systemInfo,
                exported_at: new Date().toISOString()
            };

            const blob = new Blob([JSON.stringify(data, null, 2)], { 
                type: 'application/json' 
            });
            
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `plugin-data-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            this.showNotification('success', 'Export Complete', 'Plugin data downloaded successfully');
        },

        // Bulk operations
        async bulkTogglePlugins(pluginNames, activate = true) {
            const action = activate ? 'activate' : 'deactivate';
            let successCount = 0;
            let failCount = 0;

            for (const name of pluginNames) {
                try {
                    await this.togglePlugin(name);
                    successCount++;
                } catch (error) {
                    failCount++;
                    console.error(`Failed to ${action} plugin ${name}:`, error);
                }
            }

            const message = `${successCount} plugins ${action}d successfully`;
            if (failCount > 0) {
                message += `, ${failCount} failed`;
            }

            this.showNotification(
                failCount === 0 ? 'success' : 'warning',
                'Bulk Operation Complete',
                message
            );

            await this.loadPlugins();
        }
    };
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // If Alpine.js is not available, provide fallback
    if (typeof Alpine === 'undefined') {
        console.warn('Alpine.js not loaded, plugin management may not work correctly');
        return;
    }

    // Auto-start the plugin manager
    Alpine.start();
});

// Global utilities
window.pluginUtils = {
    // Generate plugin template
    generatePluginTemplate(name, description = '') {
        return {
            name: name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
            version: '1.0.0',
            description: description || `A ${name} plugin for the CMS`,
            author: 'Plugin Developer',
            main: 'main.go',
            dependencies: {
                go: '1.21',
                gin: 'v1.9.1'
            }
        };
    },

    // Validate plugin name
    validatePluginName(name) {
        const pattern = /^[a-z0-9-]+$/;
        return pattern.test(name) && name.length > 0 && name.length <= 50;
    },

    // Plugin development helper
    generateReadme(pluginName, description) {
        return `# ${pluginName.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')} Plugin

${description}

## Installation

1. Create a zip file with the plugin contents
2. Upload via admin interface or API
3. Plugin will be compiled and available immediately

## API Endpoints

- \`GET /api/v1/plugins/${pluginName}/\` - Plugin index
- \`GET /api/v1/plugins/${pluginName}/info\` - Plugin information  
- \`GET /api/v1/plugins/${pluginName}/status\` - Plugin status
- \`POST /api/v1/plugins/${pluginName}/action\` - Execute actions

## Development

Built with the Go CMS WordPress-like plugin system.
`;
    }
};