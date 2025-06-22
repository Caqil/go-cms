class PluginManager {
    constructor() {
        this.apiBase = '/api/v1';
        this.token = localStorage.getItem('authToken');
        this.plugins = [];
        this.filteredPlugins = [];
        this.currentPlugin = null;
        
        if (!this.token) {
            window.location.href = '../login.html';
            return;
        }

        this.init();
    }

    async init() {
        this.setupEventListeners();
        await this.loadPlugins();
        this.renderPlugins();
    }

    setupEventListeners() {
        // Search functionality
        const searchInput = document.getElementById('search');
        if (searchInput) {
            searchInput.addEventListener('input', UIComponents.debounce((e) => {
                this.filterPlugins(e.target.value);
            }, 300));
        }

        // Status filter
        const statusFilter = document.getElementById('status-filter');
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                this.filterPluginsByStatus(e.target.value);
            });
        }

        // Settings modal
        const saveSettingsBtn = document.getElementById('save-settings');
        if (saveSettingsBtn) {
            saveSettingsBtn.addEventListener('click', () => this.savePluginSettings());
        }

        const cancelSettingsBtn = document.getElementById('cancel-settings');
        if (cancelSettingsBtn) {
            cancelSettingsBtn.addEventListener('click', () => this.closeSettingsModal());
        }
    }

    async loadPlugins() {
        try {
            this.showLoading();
            const response = await this.apiCall('/admin/plugins');
            this.plugins = response.plugins || [];
            this.filteredPlugins = [...this.plugins];
            this.hideLoading();
        } catch (error) {
            console.error('Failed to load plugins:', error);
            this.showError('Failed to load plugins: ' + error.message);
            this.hideLoading();
        }
    }

    renderPlugins() {
        const pluginsGrid = document.getElementById('plugins-grid');
        const emptyState = document.getElementById('empty-state');
        const loadingState = document.getElementById('loading-state');

        if (!pluginsGrid) return;

        // Hide loading state
        if (loadingState) loadingState.style.display = 'none';

        if (this.filteredPlugins.length === 0) {
            pluginsGrid.innerHTML = '';
            if (emptyState) emptyState.classList.remove('hidden');
            return;
        }

        if (emptyState) emptyState.classList.add('hidden');

        pluginsGrid.innerHTML = this.filteredPlugins.map(plugin => this.createPluginCard(plugin)).join('');
    }

    createPluginCard(plugin) {
        const statusBadge = plugin.is_active 
            ? '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">Active</span>'
            : '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">Inactive</span>';

        return `
            <div class="bg-white overflow-hidden shadow rounded-lg border border-gray-200 hover:shadow-md transition-shadow plugin-card" data-plugin="${plugin.name}">
                <div class="px-4 py-5 sm:p-6">
                    <div class="flex items-start justify-between">
                        <div class="flex-1">
                            <h3 class="text-lg font-medium text-gray-900 mb-1">${plugin.name}</h3>
                            <p class="text-sm text-gray-500 mb-2">v${plugin.version}</p>
                            ${statusBadge}
                        </div>
                        <div class="flex-shrink-0">
                            <div class="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center">
                                <svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z"></path>
                                </svg>
                            </div>
                        </div>
                    </div>
                    
                    <div class="mt-4">
                        <p class="text-sm text-gray-600">${plugin.description || 'No description available'}</p>
                    </div>
                    
                    <div class="mt-4 flex items-center text-sm text-gray-500">
                        <svg class="flex-shrink-0 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                        </svg>
                        By ${plugin.author || 'Unknown'}
                    </div>
                    
                    <div class="mt-6 flex flex-wrap gap-2">
                        <button onclick="pluginManager.showPluginSettings('${plugin.name}')" 
                            class="inline-flex items-center px-3 py-1.5 border border-gray-300 shadow-sm text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                            <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                            </svg>
                            Settings
                        </button>
                        
                        <button onclick="pluginManager.togglePlugin('${plugin.name}', ${!plugin.is_active})" 
                            class="inline-flex items-center px-3 py-1.5 border border-transparent shadow-sm text-xs font-medium rounded text-white ${plugin.is_active ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'} focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                            <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="${plugin.is_active ? 'M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z' : 'M14.828 14.828a4 4 0 01-5.656 0M9 10h1.586a1 1 0 01.707.293l2.414 2.414a1 1 0 00.707.293H15M9 10v4a2 2 0 002 2h2a2 2 0 002-2v-4M9 10V9a2 2 0 012-2h2a2 2 0 012 2v1'}"></path>
                            </svg>
                            ${plugin.is_active ? 'Deactivate' : 'Activate'}
                        </button>
                        
                        <button onclick="pluginManager.deletePlugin('${plugin.name}')" 
                            class="inline-flex items-center px-3 py-1.5 border border-transparent shadow-sm text-xs font-medium rounded text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500">
                            <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                            </svg>
                            Delete
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    async showPluginSettings(pluginName) {
        try {
            const response = await this.apiCall(`/admin/plugins/${pluginName}/settings`);
            const settings = response.settings || [];
            
            this.currentPlugin = pluginName;
            this.renderSettingsModal(pluginName, settings);
        } catch (error) {
            console.error('Failed to load plugin settings:', error);
            this.showError('Failed to load plugin settings: ' + error.message);
        }
    }

    renderSettingsModal(pluginName, settings) {
        const modalTitle = document.getElementById('modal-title');
        const settingsForm = document.getElementById('settings-form');
        
        if (modalTitle) modalTitle.textContent = `${pluginName} Settings`;
        
        if (settingsForm) {
            settingsForm.innerHTML = settings.map(setting => 
                this.createSettingField(setting)
            ).join('');
        }

        UIComponents.showModal('settings-modal');
    }

    createSettingField(setting) {
        switch (setting.type) {
            case 'boolean':
                return UIComponents.createFormField(
                    setting.label,
                    'checkbox',
                    setting.key,
                    setting.value,
                    setting.required,
                    { description: setting.description }
                );
            case 'select':
                return UIComponents.createFormField(
                    setting.label,
                    'select',
                    setting.key,
                    setting.value,
                    setting.required,
                    {
                        description: setting.description,
                        options: setting.options?.map(opt => ({ value: opt, label: opt })) || []
                    }
                );
            case 'number':
                return UIComponents.createFormField(
                    setting.label,
                    'number',
                    setting.key,
                    setting.value,
                    setting.required,
                    { description: setting.description }
                );
            default:
                return UIComponents.createFormField(
                    setting.label,
                    'text',
                    setting.key,
                    setting.value,
                    setting.required,
                    { description: setting.description }
                );
        }
    }

    async savePluginSettings() {
        if (!this.currentPlugin) return;

        try {
            const form = document.getElementById('settings-form');
            const formData = new FormData(form);
            const settings = {};

            for (let [key, value] of formData.entries()) {
                const field = form.querySelector(`[name="${key}"]`);
                if (field) {
                    if (field.type === 'checkbox') {
                        settings[key] = field.checked;
                    } else if (field.type === 'number') {
                        settings[key] = parseFloat(value) || 0;
                    } else {
                        settings[key] = value;
                    }
                }
            }

            await this.apiCall(`/admin/plugins/${this.currentPlugin}/settings`, {
                method: 'PUT',
                body: JSON.stringify(settings)
            });

            this.closeSettingsModal();
            this.showSuccess('Settings saved successfully');
        } catch (error) {
            console.error('Failed to save settings:', error);
            this.showError('Failed to save settings: ' + error.message);
        }
    }

    closeSettingsModal() {
        UIComponents.closeModal('settings-modal');
        this.currentPlugin = null;
    }

    async togglePlugin(pluginName, activate) {
        try {
            await this.apiCall(`/admin/plugins/${pluginName}/toggle`, {
                method: 'POST'
            });

            // Update plugin status in local data
            const plugin = this.plugins.find(p => p.name === pluginName);
            if (plugin) {
                plugin.is_active = activate;
            }

            this.filterPlugins(); // Re-apply current filters
            this.showSuccess(`Plugin ${activate ? 'activated' : 'deactivated'} successfully`);
        } catch (error) {
            console.error('Failed to toggle plugin:', error);
            this.showError('Failed to toggle plugin: ' + error.message);
        }
    }

    async deletePlugin(pluginName) {
        if (!confirm(`Are you sure you want to delete the plugin "${pluginName}"? This action cannot be undone.`)) {
            return;
        }

        try {
            await this.apiCall(`/admin/plugins/${pluginName}`, {
                method: 'DELETE'
            });

            // Remove plugin from local data
            this.plugins = this.plugins.filter(p => p.name !== pluginName);
            this.filterPlugins(); // Re-apply current filters
            this.showSuccess('Plugin deleted successfully');
        } catch (error) {
            console.error('Failed to delete plugin:', error);
            this.showError('Failed to delete plugin: ' + error.message);
        }
    }

    filterPlugins(searchQuery = '') {
        const statusFilter = document.getElementById('status-filter')?.value || '';
        
        this.filteredPlugins = this.plugins.filter(plugin => {
            const matchesSearch = !searchQuery || 
                plugin.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                (plugin.description && plugin.description.toLowerCase().includes(searchQuery.toLowerCase())) ||
                (plugin.author && plugin.author.toLowerCase().includes(searchQuery.toLowerCase()));
            
            const matchesStatus = !statusFilter || 
                (statusFilter === 'active' && plugin.is_active) ||
                (statusFilter === 'inactive' && !plugin.is_active);
            
            return matchesSearch && matchesStatus;
        });

        this.renderPlugins();
    }

    filterPluginsByStatus(status) {
        const searchInput = document.getElementById('search');
        const searchQuery = searchInput ? searchInput.value : '';
        this.filterPlugins(searchQuery);
    }

    showLoading() {
        const loadingState = document.getElementById('loading-state');
        if (loadingState) loadingState.style.display = 'grid';
    }

    hideLoading() {
        const loadingState = document.getElementById('loading-state');
        if (loadingState) loadingState.style.display = 'none';
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
            localStorage.removeItem('authToken');
            window.location.href = '../login.html';
            throw new Error('Unauthorized');
        }

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'API call failed');
        }

        return data;
    }

    showError(message) {
        UIComponents.showNotification(message, 'error');
    }

    showSuccess(message) {
        UIComponents.showNotification(message, 'success');
    }
}

// Initialize plugin manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.pluginManager = new PluginManager();
});