
// Component: Dropdown Menus
class DropdownMenu {
    constructor(element) {
        this.element = element;
        this.trigger = element.querySelector('.dropdown-trigger');
        this.menu = element.querySelector('.dropdown-menu');
        this.isOpen = false;
        
        this.init();
    }

    init() {
        if (this.trigger && this.menu) {
            this.trigger.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggle();
            });

            // Close on outside click
            document.addEventListener('click', (e) => {
                if (!this.element.contains(e.target)) {
                    this.close();
                }
            });

            // Close on escape
            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && this.isOpen) {
                    this.close();
                }
            });
        }
    }

    toggle() {
        this.isOpen ? this.close() : this.open();
    }

    open() {
        this.menu.classList.remove('hidden');
        this.menu.classList.add('opacity-100', 'visible');
        this.isOpen = true;
        this.trigger.setAttribute('aria-expanded', 'true');
    }

    close() {
        this.menu.classList.add('hidden');
        this.menu.classList.remove('opacity-100', 'visible');
        this.isOpen = false;
        this.trigger.setAttribute('aria-expanded', 'false');
    }
}

// Component: Accordion
class Accordion {
    constructor(element) {
        this.element = element;
        this.items = element.querySelectorAll('.accordion-item');
        this.allowMultiple = element.hasAttribute('data-allow-multiple');
        
        this.init();
    }

    init() {
        this.items.forEach((item, index) => {
            const trigger = item.querySelector('.accordion-trigger');
            const content = item.querySelector('.accordion-content');
            
            if (trigger && content) {
                trigger.addEventListener('click', () => {
                    this.toggleItem(item, index);
                });
                
                // Set up ARIA attributes
                trigger.setAttribute('aria-controls', `accordion-content-${index}`);
                content.setAttribute('id', `accordion-content-${index}`);
                
                // Set initial state
                if (item.hasAttribute('data-open')) {
                    this.openItem(item);
                } else {
                    this.closeItem(item);
                }
            }
        });
    }

    toggleItem(item, index) {
        const isOpen = item.classList.contains('open');
        
        if (!this.allowMultiple) {
            // Close all other items
            this.items.forEach((otherItem, otherIndex) => {
                if (otherIndex !== index) {
                    this.closeItem(otherItem);
                }
            });
        }
        
        if (isOpen) {
            this.closeItem(item);
        } else {
            this.openItem(item);
        }
    }

    openItem(item) {
        const trigger = item.querySelector('.accordion-trigger');
        const content = item.querySelector('.accordion-content');
        
        item.classList.add('open');
        trigger.setAttribute('aria-expanded', 'true');
        content.style.height = content.scrollHeight + 'px';
    }

    closeItem(item) {
        const trigger = item.querySelector('.accordion-trigger');
        const content = item.querySelector('.accordion-content');
        
        item.classList.remove('open');
        trigger.setAttribute('aria-expanded', 'false');
        content.style.height = '0px';
    }
}

// Component: Modal
class Modal {
    constructor(element) {
        this.element = element;
        this.backdrop = element.querySelector('.modal-backdrop');
        this.content = element.querySelector('.modal-content');
        this.closeButtons = element.querySelectorAll('[data-modal-close]');
        this.isOpen = false;
        
        this.init();
    }

    init() {
        // Close button handlers
        this.closeButtons.forEach(button => {
            button.addEventListener('click', () => this.close());
        });

        // Backdrop click to close
        if (this.backdrop) {
            this.backdrop.addEventListener('click', (e) => {
                if (e.target === this.backdrop) {
                    this.close();
                }
            });
        }

        // Escape key to close
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.close();
            }
        });
    }

    open() {
        this.element.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        this.isOpen = true;
        
        // Focus management
        const focusableElement = this.content.querySelector('input, button, [tabindex]');
        if (focusableElement) {
            focusableElement.focus();
        }
        
        // Animation
        requestAnimationFrame(() => {
            this.element.classList.add('modal-open');
        });
    }

    close() {
        this.element.classList.remove('modal-open');
        document.body.style.overflow = '';
        this.isOpen = false;
        
        // Wait for animation to complete
        setTimeout(() => {
            this.element.classList.add('hidden');
        }, 200);
    }
}

// Component: Tabs
class Tabs {
    constructor(element) {
        this.element = element;
        this.tabList = element.querySelector('.tab-list');
        this.tabs = element.querySelectorAll('.tab');
        this.panels = element.querySelectorAll('.tab-panel');
        this.activeIndex = 0;
        
        this.init();
    }

    init() {
        this.tabs.forEach((tab, index) => {
            tab.addEventListener('click', (e) => {
                e.preventDefault();
                this.setActiveTab(index);
            });
            
            // Keyboard navigation
            tab.addEventListener('keydown', (e) => {
                switch (e.key) {
                    case 'ArrowLeft':
                        e.preventDefault();
                        this.setActiveTab(index > 0 ? index - 1 : this.tabs.length - 1);
                        break;
                    case 'ArrowRight':
                        e.preventDefault();
                        this.setActiveTab(index < this.tabs.length - 1 ? index + 1 : 0);
                        break;
                    case 'Home':
                        e.preventDefault();
                        this.setActiveTab(0);
                        break;
                    case 'End':
                        e.preventDefault();
                        this.setActiveTab(this.tabs.length - 1);
                        break;
                }
            });
        });
        
        // Set initial active tab
        const initialActive = this.element.querySelector('.tab.active');
        if (initialActive) {
            this.activeIndex = Array.from(this.tabs).indexOf(initialActive);
        }
        this.setActiveTab(this.activeIndex);
    }

    setActiveTab(index) {
        // Remove active class from all tabs and panels
        this.tabs.forEach(tab => tab.classList.remove('active'));
        this.panels.forEach(panel => panel.classList.remove('active'));
        
        // Add active class to selected tab and panel
        this.tabs[index].classList.add('active');
        this.panels[index].classList.add('active');
        
        // Update ARIA attributes
        this.tabs[index].setAttribute('aria-selected', 'true');
        this.tabs[index].focus();
        
        this.activeIndex = index;
    }
}

// Component: Toast Notifications
class ToastManager {
    constructor() {
        this.container = null;
        this.toasts = [];
        this.init();
    }

    init() {
        // Create toast container if it doesn't exist
        this.container = document.getElementById('toast-container');
        if (!this.container) {
            this.container = document.createElement('div');
            this.container.id = 'toast-container';
            this.container.className = 'fixed top-4 right-4 z-50 space-y-2';
            document.body.appendChild(this.container);
        }
    }

    show(message, type = 'info', duration = 5000) {
        const toast = this.createToast(message, type);
        this.container.appendChild(toast);
        this.toasts.push(toast);
        
        // Show animation
        requestAnimationFrame(() => {
            toast.classList.add('toast-show');
        });
        
        // Auto dismiss
        if (duration > 0) {
            setTimeout(() => {
                this.dismiss(toast);
            }, duration);
        }
        
        return toast;
    }

    createToast(message, type) {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type} bg-surface border border-border rounded-lg shadow-lg p-4 max-w-sm transform translate-x-full transition-transform duration-300`;
        
        const icon = this.getIcon(type);
        const color = this.getColor(type);
        
        toast.innerHTML = `
            <div class="flex items-start">
                <div class="flex-shrink-0">
                    <svg class="w-5 h-5 ${color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        ${icon}
                    </svg>
                </div>
                <div class="ml-3 flex-1">
                    <p class="text-sm text-text-primary">${message}</p>
                </div>
                <div class="ml-4 flex-shrink-0">
                    <button class="toast-close text-text-secondary hover:text-text-primary">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
            </div>
        `;
        
        // Close button handler
        const closeButton = toast.querySelector('.toast-close');
        closeButton.addEventListener('click', () => this.dismiss(toast));
        
        return toast;
    }

    dismiss(toast) {
        toast.classList.remove('toast-show');
        toast.classList.add('translate-x-full');
        
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
            const index = this.toasts.indexOf(toast);
            if (index > -1) {
                this.toasts.splice(index, 1);
            }
        }, 300);
    }

    getIcon(type) {
        const icons = {
            success: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>',
            error: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>',
            warning: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z"></path>',
            info: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>'
        };
        return icons[type] || icons.info;
    }

    getColor(type) {
        const colors = {
            success: 'text-success',
            error: 'text-error',
            warning: 'text-warning',
            info: 'text-primary'
        };
        return colors[type] || colors.info;
    }
}

// Initialize components when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Initialize dropdowns
    document.querySelectorAll('.dropdown').forEach(element => {
        new DropdownMenu(element);
    });

    // Initialize accordions
    document.querySelectorAll('.accordion').forEach(element => {
        new Accordion(element);
    });

    // Initialize modals
    document.querySelectorAll('.modal').forEach(element => {
        new Modal(element);
    });

    // Initialize tabs
    document.querySelectorAll('.tabs').forEach(element => {
        new Tabs(element);
    });

    // Initialize toast manager
    window.toastManager = new ToastManager();
    
    // Modal triggers
    document.querySelectorAll('[data-modal-target]').forEach(trigger => {
        trigger.addEventListener('click', (e) => {
            e.preventDefault();
            const targetId = trigger.getAttribute('data-modal-target');
            const modal = document.getElementById(targetId);
            if (modal && modal.modalInstance) {
                modal.modalInstance.open();
            }
        });
    });
});

// Add CSS for components
const componentStyles = `
    .toast-show {
        transform: translateX(0);
    }
    
    .modal-open .modal-backdrop {
        opacity: 1;
    }
    
    .modal-open .modal-content {
        transform: scale(1);
        opacity: 1;
    }
    
    .accordion-content {
        overflow: hidden;
        transition: height 0.3s ease;
    }
    
    .dropdown-menu {
        transition: opacity 0.2s ease, visibility 0.2s ease;
    }
`;

// Inject component styles
const style = document.createElement('style');
style.textContent = componentStyles;
document.head.appendChild(style);