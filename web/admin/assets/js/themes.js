class ThemeManager {
    constructor() {
        this.apiBase = '/api/v1';
        this.token = localStorage.getItem('authToken');
        this.themes = [];
        this.activeTheme = null;
        this.currentCustomization = {};
        
        if (!this.token) {
            window.location.href = '../login.html';
            return;
        }

        this.init();
    }

    async init() {
        this.setupEventListeners();
        await this.loadThemes();
        this.renderThemes();
    }

    setupEventListeners() {
        // Theme activation
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-action="activate"]')) {
                const themeName = e.target.dataset.theme;
                this.activateTheme(themeName);
            }
            
            if (e.target.matches('[data-action="customize"]')) {
                const themeName = e.target.dataset.theme;
                this.showCustomization(themeName);
            }
            
            if (e.target.matches('[data-action="delete"]')) {
                const themeName = e.target.dataset.theme;
                this.deleteTheme(themeName);
            }
        });

        // Customization form
        const customizationForm = document.getElementById('customization-form');
        if (customizationForm) {
            customizationForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.saveCustomization();
            });
        }

        // Color picker changes
        document.addEventListener('input', (e) => {
            if (e.target.matches('input[type="color"]')) {
                this.previewColorChange(e.target);
            }
        });

        // Font changes
        document.addEventListener('change', (e) => {
            if (e.target.matches('select[name*="font"]')) {
                this.previewFontChange(e.target);
            }
        });
    }

    async loadThemes() {
        try {
            this.showLoading();
            const response = await this.apiCall('/themes');
            this.themes = response.themes || [];
            this.activeTheme = response.active || null;
            this.hideLoading();
        } catch (error) {
            console.error('Failed to load themes:', error);
            this.showError('Failed to load themes: ' + error.message);
            this.hideLoading();
        }
    }

    renderThemes() {
        const themesGrid = document.getElementById('themes-grid');
        if (!themesGrid) return;

        if (this.themes.length === 0) {
            themesGrid.innerHTML = UIComponents.createEmptyState(
                'No themes found',
                'Install themes to customize your site appearance',
                'Install Theme',
                'themeManager.showInstallModal()',
                '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17v4a2 2 0 002 2h4M13 13h4a2 2 0 012 2v4a2 2 0 01-2 2H9a2 2 0 01-2-2v-4a2 2 0 012-2h4z"></path>'
            );
            return;
        }

        themesGrid.innerHTML = this.themes.map(theme => this.createThemeCard(theme)).join('');
    }

    createThemeCard(theme) {
        const isActive = theme.name === this.activeTheme || theme.is_active;
        
        return `
            <div class="bg-white overflow-hidden shadow rounded-lg border ${isActive ? 'border-primary-500 ring-2 ring-primary-200' : 'border-gray-200'} theme-card" data-theme="${theme.name}">
                <div class="aspect-w-16 aspect-h-9 bg-gray-200">
                    ${theme.screenshot ? 
                        `<img src="${theme.screenshot}" alt="${theme.name} screenshot" class="w-full h-48 object-cover">` :
                        `<div class="w-full h-48 flex items-center justify-center bg-gradient-to-br from-primary-50 to-primary-100">
                            <svg class="w-12 h-12 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17v4a2 2 0 002 2h4M13 13h4a2 2 0 012 2v4a2 2 0 01-2 2H9a2 2 0 01-2-2v-4a2 2 0 012-2h4z"></path>
                            </svg>
                        </div>`
                    }
                    ${isActive ? 
                        `<div class="absolute top-2 right-2">
                            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                                Active
                            </span>
                        </div>` : ''
                    }
                </div>
                
                <div class="px-4 py-4">
                    <div class="flex items-start justify-between">
                        <div class="flex-1">
                            <h3 class="text-lg font-medium text-gray-900">${theme.name}</h3>
                            <p class="text-sm text-gray-500">v${theme.version}</p>
                            <p class="mt-2 text-sm text-gray-600">${theme.description || 'No description available'}</p>
                        </div>
                    </div>
                    
                    <div class="mt-4 flex items-center text-sm text-gray-500">
                        <svg class="flex-shrink-0 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                        </svg>
                        By ${theme.author || 'Unknown'}
                    </div>
                    
                    ${theme.tags && theme.tags.length > 0 ? `
                        <div class="mt-3 flex flex-wrap gap-1">
                            ${theme.tags.map(tag => 
                                `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">${tag}</span>`
                            ).join('')}
                        </div>
                    ` : ''}
                    
                    <div class="mt-6 flex flex-wrap gap-2">
                        ${!isActive ? `
                            <button data-action="activate" data-theme="${theme.name}" 
                                class="inline-flex items-center px-3 py-1.5 border border-transparent shadow-sm text-xs font-medium rounded text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                                <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                                </svg>
                                Activate
                            </button>
                        ` : ''}
                        
                        <button data-action="customize" data-theme="${theme.name}" 
                            class="inline-flex items-center px-3 py-1.5 border border-gray-300 shadow-sm text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                            <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                            </svg>
                            Customize
                        </button>
                        
                        <button onclick="themeManager.previewTheme('${theme.name}')" 
                            class="inline-flex items-center px-3 py-1.5 border border-gray-300 shadow-sm text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                            <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                            </svg>
                            Preview
                        </button>
                        
                        ${!isActive ? `
                            <button data-action="delete" data-theme="${theme.name}" 
                                class="inline-flex items-center px-3 py-1.5 border border-transparent shadow-sm text-xs font-medium rounded text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500">
                                <svg class="-ml-0.5 mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                                </svg>
                                Delete
                            </button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }

    async activateTheme(themeName) {
        try {
            await this.apiCall(`/themes/${themeName}/activate`, {
                method: 'POST'
            });

            this.activeTheme = themeName;
            this.renderThemes();
            this.showSuccess(`Theme "${themeName}" activated successfully`);
        } catch (error) {
            console.error('Failed to activate theme:', error);
            this.showError('Failed to activate theme: ' + error.message);
        }
    }

    async showCustomization(themeName) {
        try {
            const response = await this.apiCall(`/themes/${themeName}/customization`);
            this.currentCustomization = response.customization || {};
            
            // Create customization modal or redirect to customization page
            window.location.href = `customize.html?theme=${themeName}`;
        } catch (error) {
            console.error('Failed to load customization:', error);
            this.showError('Failed to load customization options: ' + error.message);
        }
    }

    async deleteTheme(themeName) {
        if (!confirm(`Are you sure you want to delete the theme "${themeName}"? This action cannot be undone.`)) {
            return;
        }

        try {
            await this.apiCall(`/themes/${themeName}`, {
                method: 'DELETE'
            });

            this.themes = this.themes.filter(t => t.name !== themeName);
            this.renderThemes();
            this.showSuccess('Theme deleted successfully');
        } catch (error) {
            console.error('Failed to delete theme:', error);
            this.showError('Failed to delete theme: ' + error.message);
        }
    }

    previewTheme(themeName) {
        // Open theme preview in new tab/window
        const previewUrl = `${window.location.origin}/?theme_preview=${themeName}`;
        window.open(previewUrl, '_blank');
    }

    showInstallModal() {
        const modal = UIComponents.createModal('theme-install-modal', 'Install Theme', `
            <form id="theme-install-form" class="space-y-4">
                ${UIComponents.createFormField('Theme File', 'file', 'theme_file', '', true, {
                    description: 'Upload a theme package (.zip file)'
                })}
                ${UIComponents.createFormField('Theme Name', 'text', 'theme_name', '', true, {
                    placeholder: 'Enter theme name'
                })}
                ${UIComponents.createFormField('Description', 'textarea', 'description', '', false, {
                    placeholder: 'Optional theme description',
                    rows: 3
                })}
            </form>
        `, {
            buttons: `
                <button type="button" onclick="themeManager.installTheme()" 
                    class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-primary-600 text-base font-medium text-white hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:ml-3 sm:w-auto sm:text-sm">
                    Install
                </button>
                <button type="button" onclick="UIComponents.closeModal('theme-install-modal')" 
                    class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:mt-0 sm:w-auto sm:text-sm">
                    Cancel
                </button>
            `
        });

        UIComponents.showModal('theme-install-modal');
    }

    async installTheme() {
        const form = document.getElementById('theme-install-form');
        const formData = new FormData(form);

        try {
            const response = await fetch(`${this.apiBase}/themes/install`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                },
                body: formData
            });

            const result = await response.json();

            if (response.ok) {
                UIComponents.closeModal('theme-install-modal');
                this.showSuccess('Theme installed successfully');
                await this.loadThemes();
                this.renderThemes();
            } else {
                throw new Error(result.error || 'Installation failed');
            }
        } catch (error) {
            console.error('Failed to install theme:', error);
            this.showError('Failed to install theme: ' + error.message);
        }
    }

    previewColorChange(input) {
        const colorName = input.name;
        const colorValue = input.value;
        
        // Apply preview styles
        document.documentElement.style.setProperty(`--preview-${colorName}`, colorValue);
    }

    previewFontChange(select) {
        const fontProperty = select.name;
        const fontValue = select.value;
        
        // Apply preview styles
        document.documentElement.style.setProperty(`--preview-${fontProperty}`, fontValue);
    }

    async saveCustomization() {
        try {
            const form = document.getElementById('customization-form');
            const formData = new FormData(form);
            const customization = {};

            // Process form data
            for (let [key, value] of formData.entries()) {
                if (key.startsWith('color_')) {
                    if (!customization.colors) customization.colors = {};
                    customization.colors[key.replace('color_', '')] = value;
                } else if (key.startsWith('font_')) {
                    if (!customization.fonts) customization.fonts = {};
                    customization.fonts[key.replace('font_', '')] = value;
                } else {
                    customization[key] = value;
                }
            }

            const themeName = new URLSearchParams(window.location.search).get('theme');
            
            await this.apiCall(`/themes/${themeName}/customization`, {
                method: 'PUT',
                body: JSON.stringify(customization)
            });

            this.showSuccess('Customization saved successfully');
        } catch (error) {
            console.error('Failed to save customization:', error);
            this.showError('Failed to save customization: ' + error.message);
        }
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

// Initialize theme manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.themeManager = new ThemeManager();
});