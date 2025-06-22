// Reusable UI Components
class UIComponents {
    static createModal(id, title, content, options = {}) {
        const modal = document.createElement('div');
        modal.id = id;
        modal.className = 'fixed inset-0 z-50 overflow-y-auto hidden';
        modal.innerHTML = `
            <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onclick="UIComponents.closeModal('${id}')"></div>
                <span class="hidden sm:inline-block sm:align-middle sm:h-screen">&#8203;</span>
                <div class="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle ${options.size || 'sm:max-w-lg'} sm:w-full sm:p-6">
                    <div class="sm:flex sm:items-start">
                        <div class="mt-3 text-center sm:mt-0 sm:text-left w-full">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">
                                ${title}
                            </h3>
                            <div class="modal-content">
                                ${content}
                            </div>
                        </div>
                    </div>
                    <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                        ${options.buttons || `
                            <button type="button" onclick="UIComponents.closeModal('${id}')" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:mt-0 sm:w-auto sm:text-sm">
                                Close
                            </button>
                        `}
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        return modal;
    }

    static showModal(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.remove('hidden');
            document.body.style.overflow = 'hidden';
        }
    }

    static closeModal(id) {
        const modal = document.getElementById(id);
        if (modal) {
            modal.classList.add('hidden');
            document.body.style.overflow = 'auto';
        }
    }

    static createCard(title, content, actions = '') {
        return `
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="px-4 py-5 sm:p-6">
                    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">${title}</h3>
                    <div class="card-content">
                        ${content}
                    </div>
                    ${actions ? `<div class="mt-4 flex space-x-2">${actions}</div>` : ''}
                </div>
            </div>
        `;
    }

    static createButton(text, type = 'primary', onclick = '', icon = '') {
        const colors = {
            primary: 'bg-primary-600 hover:bg-primary-700 text-white',
            secondary: 'bg-gray-600 hover:bg-gray-700 text-white',
            success: 'bg-green-600 hover:bg-green-700 text-white',
            danger: 'bg-red-600 hover:bg-red-700 text-white',
            warning: 'bg-yellow-600 hover:bg-yellow-700 text-white',
            outline: 'border border-gray-300 text-gray-700 bg-white hover:bg-gray-50'
        };

        return `
            <button type="button" onclick="${onclick}" 
                class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium ${colors[type]} focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed">
                ${icon ? `<svg class="-ml-1 mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">${icon}</svg>` : ''}
                ${text}
            </button>
        `;
    }

    static createAlert(message, type = 'info', dismissible = true) {
        const colors = {
            success: 'bg-green-50 border-green-200 text-green-800',
            error: 'bg-red-50 border-red-200 text-red-800',
            warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
            info: 'bg-blue-50 border-blue-200 text-blue-800'
        };

        const icons = {
            success: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>',
            error: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>',
            warning: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z"></path>',
            info: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>'
        };

        return `
            <div class="rounded-md border p-4 ${colors[type]} ${dismissible ? 'alert-dismissible' : ''}" role="alert">
                <div class="flex">
                    <div class="flex-shrink-0">
                        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            ${icons[type]}
                        </svg>
                    </div>
                    <div class="ml-3">
                        <p class="text-sm font-medium">${message}</p>
                    </div>
                    ${dismissible ? `
                        <div class="ml-auto pl-3">
                            <div class="-mx-1.5 -my-1.5">
                                <button type="button" onclick="this.closest('.alert-dismissible').remove()" 
                                    class="inline-flex rounded-md p-1.5 text-current hover:bg-black hover:bg-opacity-10 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-current focus:ring-current">
                                    <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                                    </svg>
                                </button>
                            </div>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
    }

    static createFormField(label, type, name, value = '', required = false, options = {}) {
        const fieldId = `field-${name}`;
        let inputHtml = '';

        switch (type) {
            case 'textarea':
                inputHtml = `<textarea id="${fieldId}" name="${name}" rows="${options.rows || 3}" 
                    class="block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm" 
                    ${required ? 'required' : ''} ${options.placeholder ? `placeholder="${options.placeholder}"` : ''}>${value}</textarea>`;
                break;
            case 'select':
                inputHtml = `
                    <select id="${fieldId}" name="${name}" 
                        class="block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm" 
                        ${required ? 'required' : ''}>
                        ${options.options ? options.options.map(opt => 
                            `<option value="${opt.value}" ${opt.value === value ? 'selected' : ''}>${opt.label}</option>`
                        ).join('') : ''}
                    </select>
                `;
                break;
            case 'checkbox':
                inputHtml = `
                    <div class="flex items-center">
                        <input id="${fieldId}" name="${name}" type="checkbox" value="1" 
                            class="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded" 
                            ${value ? 'checked' : ''}>
                        <label for="${fieldId}" class="ml-2 block text-sm text-gray-900">${label}</label>
                    </div>
                `;
                return `<div class="form-field">${inputHtml}</div>`;
            default:
                inputHtml = `<input type="${type}" id="${fieldId}" name="${name}" value="${value}" 
                    class="block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm" 
                    ${required ? 'required' : ''} ${options.placeholder ? `placeholder="${options.placeholder}"` : ''}
                    ${options.min ? `min="${options.min}"` : ''} ${options.max ? `max="${options.max}"` : ''}
                    ${options.step ? `step="${options.step}"` : ''}>`;
        }

        return `
            <div class="form-field">
                <label for="${fieldId}" class="block text-sm font-medium text-gray-700 mb-1">
                    ${label} ${required ? '<span class="text-red-500">*</span>' : ''}
                </label>
                ${inputHtml}
                ${options.description ? `<p class="mt-1 text-sm text-gray-500">${options.description}</p>` : ''}
            </div>
        `;
    }

    static createTabContainer(tabs, activeTab = 0) {
        const tabNav = tabs.map((tab, index) => `
            <button type="button" onclick="UIComponents.switchTab(${index})" 
                class="tab-button ${index === activeTab ? 'active' : ''} px-4 py-2 text-sm font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
                data-tab="${index}">
                ${tab.title}
            </button>
        `).join('');

        const tabContent = tabs.map((tab, index) => `
            <div class="tab-content ${index === activeTab ? 'active' : 'hidden'}" data-tab="${index}">
                ${tab.content}
            </div>
        `).join('');

        return `
            <div class="tab-container">
                <div class="tab-nav border-b border-gray-200 mb-6">
                    <nav class="-mb-px flex space-x-8">
                        ${tabNav}
                    </nav>
                </div>
                <div class="tab-content-container">
                    ${tabContent}
                </div>
            </div>
        `;
    }

    static switchTab(activeIndex) {
        // Hide all tab contents
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.add('hidden');
            content.classList.remove('active');
        });

        // Remove active class from all tab buttons
        document.querySelectorAll('.tab-button').forEach(button => {
            button.classList.remove('active', 'border-primary-500', 'text-primary-600');
            button.classList.add('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
        });

        // Show active tab content
        const activeContent = document.querySelector(`.tab-content[data-tab="${activeIndex}"]`);
        if (activeContent) {
            activeContent.classList.remove('hidden');
            activeContent.classList.add('active');
        }

        // Style active tab button
        const activeButton = document.querySelector(`.tab-button[data-tab="${activeIndex}"]`);
        if (activeButton) {
            activeButton.classList.add('active', 'border-primary-500', 'text-primary-600');
            activeButton.classList.remove('border-transparent', 'text-gray-500', 'hover:text-gray-700', 'hover:border-gray-300');
        }
    }

    static createLoadingSpinner(text = 'Loading...') {
        return `
            <div class="flex items-center justify-center py-8">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mr-3"></div>
                <span class="text-gray-600">${text}</span>
            </div>
        `;
    }

    static createEmptyState(title, description, actionText = '', actionOnclick = '', icon = '') {
        return `
            <div class="text-center py-12">
                <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    ${icon || '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>'}
                </svg>
                <h3 class="mt-2 text-sm font-medium text-gray-900">${title}</h3>
                <p class="mt-1 text-sm text-gray-500">${description}</p>
                ${actionText ? `
                    <div class="mt-6">
                        <button type="button" onclick="${actionOnclick}" 
                            class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                            ${actionText}
                        </button>
                    </div>
                ` : ''}
            </div>
        `;
    }

    static showNotification(message, type = 'success', duration = 5000) {
        const notification = document.createElement('div');
        notification.className = 'fixed top-4 right-4 z-50 max-w-sm w-full';
        
        const colors = {
            success: 'bg-green-50 border-green-200 text-green-800',
            error: 'bg-red-50 border-red-200 text-red-800',
            warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
            info: 'bg-blue-50 border-blue-200 text-blue-800'
        };

        notification.innerHTML = `
            <div class="rounded-md border p-4 shadow-lg ${colors[type]} notification-slide-in">
                <div class="flex">
                    <div class="flex-shrink-0">
                        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                    </div>
                    <div class="ml-3">
                        <p class="text-sm font-medium">${message}</p>
                    </div>
                    <div class="ml-auto pl-3">
                        <button type="button" onclick="this.closest('div').parentElement.remove()" 
                            class="inline-flex rounded-md p-1.5 text-current hover:bg-black hover:bg-opacity-10 focus:outline-none">
                            <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(notification);

        // Auto remove after duration
        setTimeout(() => {
            if (notification.parentNode) {
                notification.classList.add('notification-slide-out');
                setTimeout(() => notification.remove(), 300);
            }
        }, duration);
    }

    static formatBytes(bytes, decimals = 2) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const dm = decimals < 0 ? 0 : decimals;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
    }

    static formatDate(date, options = {}) {
        const defaultOptions = { 
            year: 'numeric', 
            month: 'short', 
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        };
        return new Date(date).toLocaleDateString('en-US', { ...defaultOptions, ...options });
    }

    static debounce(func, wait, immediate = false) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                timeout = null;
                if (!immediate) func(...args);
            };
            const callNow = immediate && !timeout;
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
            if (callNow) func(...args);
        };
    }
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    .notification-slide-in {
        animation: slideInRight 0.3s ease-out;
    }
    
    .notification-slide-out {
        animation: slideOutRight 0.3s ease-in;
    }
    
    @keyframes slideInRight {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    
    @keyframes slideOutRight {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
    
    .tab-button.active {
        border-bottom: 2px solid #3b82f6;
        color: #3b82f6;
    }
    
    .tab-button {
        border-bottom: 2px solid transparent;
        color: #6b7280;
        transition: all 0.2s;
    }
    
    .tab-button:hover:not(.active) {
        color: #374151;
        border-bottom-color: #d1d5db;
    }
`;
document.head.appendChild(style);

// Make UIComponents globally available
window.UIComponents = UIComponents;