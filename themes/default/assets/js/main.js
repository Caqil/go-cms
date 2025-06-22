// assets/js/main.js - Main theme JavaScript

class DefaultTheme {
    constructor() {
        this.isDarkMode = this.getStoredTheme() === 'dark' || 
                         (!this.getStoredTheme() && window.matchMedia('(prefers-color-scheme: dark)').matches);
        this.mobileMenuOpen = false;
        this.searchModalOpen = false;
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.initDarkMode();
        this.initMobileMenu();
        this.initSearch();
        this.initSmoothScroll();
        this.initLazyLoading();
        this.initAnimations();
        
        // Emit theme loaded event
        document.dispatchEvent(new CustomEvent('themeLoaded'));
    }

    setupEventListeners() {
        // Dark mode toggle
        const darkModeToggle = document.getElementById('dark-mode-toggle');
        if (darkModeToggle) {
            darkModeToggle.addEventListener('click', () => this.toggleDarkMode());
        }

        // Mobile menu toggle
        const mobileMenuToggle = document.getElementById('mobile-menu-toggle');
        if (mobileMenuToggle) {
            mobileMenuToggle.addEventListener('click', () => this.toggleMobileMenu());
        }

        // Search toggle
        const searchToggle = document.getElementById('search-toggle');
        if (searchToggle) {
            searchToggle.addEventListener('click', () => this.toggleSearchModal());
        }

        // Close modals on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeSearchModal();
                this.closeMobileMenu();
            }
        });

        // Close mobile menu on resize
        window.addEventListener('resize', () => {
            if (window.innerWidth >= 768) {
                this.closeMobileMenu();
            }
        });

        // Handle scroll events
        let ticking = false;
        window.addEventListener('scroll', () => {
            if (!ticking) {
                requestAnimationFrame(() => {
                    this.handleScroll();
                    ticking = false;
                });
                ticking = true;
            }
        });

        // Form enhancements
        this.enhanceForms();

        // External links
        this.handleExternalLinks();
    }

    // Dark Mode Functionality
    initDarkMode() {
        this.updateDarkMode();
    }

    toggleDarkMode() {
        this.isDarkMode = !this.isDarkMode;
        this.updateDarkMode();
        this.storeTheme(this.isDarkMode ? 'dark' : 'light');
        
        // Emit dark mode change event
        document.dispatchEvent(new CustomEvent('darkModeChanged', {
            detail: { isDarkMode: this.isDarkMode }
        }));
    }

    updateDarkMode() {
        if (this.isDarkMode) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    }

    getStoredTheme() {
        return localStorage.getItem('theme');
    }

    storeTheme(theme) {
        localStorage.setItem('theme', theme);
    }

    // Mobile Menu Functionality
    initMobileMenu() {
        const mobileMenu = document.getElementById('mobile-menu');
        if (mobileMenu) {
            mobileMenu.classList.add('hidden');
        }
    }

    toggleMobileMenu() {
        this.mobileMenuOpen ? this.closeMobileMenu() : this.openMobileMenu();
    }

    openMobileMenu() {
        const mobileMenu = document.getElementById('mobile-menu');
        if (mobileMenu) {
            mobileMenu.classList.remove('hidden');
            document.body.style.overflow = 'hidden';
            this.mobileMenuOpen = true;
            
            // Add animation
            requestAnimationFrame(() => {
                mobileMenu.style.opacity = '1';
                mobileMenu.style.transform = 'translateY(0)';
            });
        }
    }

    closeMobileMenu() {
        const mobileMenu = document.getElementById('mobile-menu');
        if (mobileMenu && this.mobileMenuOpen) {
            mobileMenu.classList.add('hidden');
            document.body.style.overflow = '';
            this.mobileMenuOpen = false;
            
            mobileMenu.style.opacity = '';
            mobileMenu.style.transform = '';
        }
    }

    // Search Functionality
    initSearch() {
        const searchModal = document.getElementById('search-modal');
        const searchInput = document.getElementById('search-input');
        const searchResults = document.getElementById('search-results');
        
        if (searchInput) {
            let debounceTimer;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(debounceTimer);
                debounceTimer = setTimeout(() => {
                    this.performSearch(e.target.value, searchResults);
                }, 300);
            });
        }
    }

    toggleSearchModal() {
        this.searchModalOpen ? this.closeSearchModal() : this.openSearchModal();
    }

    openSearchModal() {
        const searchModal = document.getElementById('search-modal');
        const searchInput = document.getElementById('search-input');
        
        if (searchModal) {
            searchModal.classList.remove('hidden');
            document.body.style.overflow = 'hidden';
            this.searchModalOpen = true;
            
            // Focus search input
            if (searchInput) {
                setTimeout(() => searchInput.focus(), 100);
            }
            
            // Add animation
            requestAnimationFrame(() => {
                searchModal.style.opacity = '1';
                const modalContent = searchModal.querySelector('.search-modal-content');
                if (modalContent) {
                    modalContent.style.transform = 'scale(1)';
                }
            });
        }
    }

    closeSearchModal() {
        const searchModal = document.getElementById('search-modal');
        if (searchModal && this.searchModalOpen) {
            searchModal.classList.add('hidden');
            document.body.style.overflow = '';
            this.searchModalOpen = false;
            
            // Clear search results
            const searchResults = document.getElementById('search-results');
            if (searchResults) {
                searchResults.innerHTML = '';
            }
            
            // Clear search input
            const searchInput = document.getElementById('search-input');
            if (searchInput) {
                searchInput.value = '';
            }
            
            searchModal.style.opacity = '';
            const modalContent = searchModal.querySelector('.search-modal-content');
            if (modalContent) {
                modalContent.style.transform = '';
            }
        }
    }

    async performSearch(query, resultsContainer) {
        if (!query.trim() || !resultsContainer) return;
        
        resultsContainer.innerHTML = '<div class="p-4 text-center">Searching...</div>';
        
        try {
            const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
            const data = await response.json();
            
            if (data.results && data.results.length > 0) {
                resultsContainer.innerHTML = data.results.map(result => `
                    <a href="${result.url}" class="block p-4 hover:bg-surface border-b border-border">
                        <h3 class="font-semibold text-text-primary">${result.title}</h3>
                        <p class="text-sm text-text-secondary mt-1">${result.excerpt}</p>
                        <span class="text-xs text-primary">${result.type}</span>
                    </a>
                `).join('');
            } else {
                resultsContainer.innerHTML = '<div class="p-4 text-center text-text-secondary">No results found</div>';
            }
        } catch (error) {
            console.error('Search error:', error);
            resultsContainer.innerHTML = '<div class="p-4 text-center text-error">Search error occurred</div>';
        }
    }

    // Smooth Scrolling
    initSmoothScroll() {
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', (e) => {
                e.preventDefault();
                const target = document.querySelector(anchor.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    // Lazy Loading
    initLazyLoading() {
        if ('IntersectionObserver' in window) {
            const imageObserver = new IntersectionObserver((entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        img.src = img.dataset.src;
                        img.classList.remove('lazy');
                        observer.unobserve(img);
                    }
                });
            });

            document.querySelectorAll('img[data-src]').forEach(img => {
                imageObserver.observe(img);
            });
        }
    }

    // Animations
    initAnimations() {
        if ('IntersectionObserver' in window) {
            const animationObserver = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        entry.target.classList.add('animate-in');
                    }
                });
            }, {
                threshold: 0.1
            });

            document.querySelectorAll('.animate-on-scroll').forEach(el => {
                animationObserver.observe(el);
            });
        }
    }

    // Scroll Handler
    handleScroll() {
        const header = document.querySelector('header');
        if (header) {
            if (window.scrollY > 100) {
                header.classList.add('scrolled');
            } else {
                header.classList.remove('scrolled');
            }
        }

        // Update reading progress for blog posts
        this.updateReadingProgress();
    }

    updateReadingProgress() {
        const progressBar = document.getElementById('reading-progress');
        if (progressBar) {
            const article = document.querySelector('article');
            if (article) {
                const articleHeight = article.offsetHeight;
                const articleTop = article.offsetTop;
                const scrolled = window.scrollY - articleTop;
                const progress = Math.min(Math.max(scrolled / articleHeight, 0), 1);
                progressBar.style.width = `${progress * 100}%`;
            }
        }
    }

    // Form Enhancements
    enhanceForms() {
        // Add floating labels
        document.querySelectorAll('.form-group').forEach(group => {
            const input = group.querySelector('input, textarea, select');
            const label = group.querySelector('label');
            
            if (input && label) {
                input.addEventListener('focus', () => {
                    group.classList.add('focused');
                });
                
                input.addEventListener('blur', () => {
                    if (!input.value) {
                        group.classList.remove('focused');
                    }
                });
                
                if (input.value) {
                    group.classList.add('focused');
                }
            }
        });

        // Form validation
        document.querySelectorAll('form[data-validate]').forEach(form => {
            form.addEventListener('submit', (e) => {
                if (!this.validateForm(form)) {
                    e.preventDefault();
                }
            });
        });
    }

    validateForm(form) {
        let isValid = true;
        const requiredFields = form.querySelectorAll('[required]');
        
        requiredFields.forEach(field => {
            if (!field.value.trim()) {
                this.showFieldError(field, 'This field is required');
                isValid = false;
            } else {
                this.clearFieldError(field);
            }
        });

        // Email validation
        const emailFields = form.querySelectorAll('input[type="email"]');
        emailFields.forEach(field => {
            if (field.value && !this.isValidEmail(field.value)) {
                this.showFieldError(field, 'Please enter a valid email address');
                isValid = false;
            }
        });

        return isValid;
    }

    showFieldError(field, message) {
        this.clearFieldError(field);
        field.classList.add('error');
        
        const errorEl = document.createElement('div');
        errorEl.className = 'field-error';
        errorEl.textContent = message;
        field.parentNode.appendChild(errorEl);
    }

    clearFieldError(field) {
        field.classList.remove('error');
        const errorEl = field.parentNode.querySelector('.field-error');
        if (errorEl) {
            errorEl.remove();
        }
    }

    isValidEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    }

    // External Links
    handleExternalLinks() {
        document.querySelectorAll('a[href^="http"]').forEach(link => {
            if (!link.href.includes(window.location.hostname)) {
                link.setAttribute('target', '_blank');
                link.setAttribute('rel', 'noopener noreferrer');
                
                // Add external link icon
                if (!link.querySelector('.external-icon')) {
                    const icon = document.createElement('svg');
                    icon.className = 'external-icon inline w-4 h-4 ml-1';
                    icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path>';
                    link.appendChild(icon);
                }
            }
        });
    }

    // Utility Methods
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
    }

    throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // Public API
    getTheme() {
        return {
            isDarkMode: this.isDarkMode,
            mobileMenuOpen: this.mobileMenuOpen,
            searchModalOpen: this.searchModalOpen
        };
    }

    setDarkMode(isDark) {
        this.isDarkMode = isDark;
        this.updateDarkMode();
        this.storeTheme(isDark ? 'dark' : 'light');
    }
}

// Initialize theme when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.defaultTheme = new DefaultTheme();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DefaultTheme;
}

