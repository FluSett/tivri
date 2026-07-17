export function initAlpine() {
    document.addEventListener('alpine:init', function () {
        Alpine.data('layout', function () {
            return {
                mobileMenuOpen: false,
                openStepper: this.$persist(false).as('openStepper').using(sessionStorage),
                closeMobileMenu() {
                    this.mobileMenuOpen = false;
                },
                openMobileMenu() {
                    this.mobileMenuOpen = true;
                },
                getMobileMenuClass() {
                    return this.mobileMenuOpen ? '' : 'hidden';
                },
                getOverflowClass() {
                    return this.mobileMenuOpen ? 'overflow-hidden' : '';
                },
                getMobileMenuOverlayClass() {
                    return this.mobileMenuOpen
                        ? 'opacity-100 pointer-events-auto'
                        : 'opacity-0 pointer-events-none delay-300';
                },
                getMobileMenuPanelClass() {
                    return this.mobileMenuOpen ? 'translate-x-0' : 'translate-x-full';
                },
                getHeaderClass() {
                    return this.mobileMenuOpen ? 'header-hidden' : '';
                },
                setLocaleChange() {
                    sessionStorage.setItem('locale_change', 'true');
                },
                changeLanguage(lang) {
                    sessionStorage.setItem('locale_change', 'true');
                    this.mobileMenuOpen = false;
                    setTimeout(() => {
                        htmx.ajax('GET', '/api/lang?lang=' + lang, {
                            target: 'body',
                            swap: 'outerHTML transition:true'
                        });
                    }, 350);
                },
                handleLogoClick(event) {
                    if (window.location.pathname === '/' || window.location.pathname === '/admin') {
                        if (!window.location.search) {
                            event.preventDefault();
                            event.stopPropagation();
                            window.scrollTo({ top: 0, behavior: 'smooth' });
                        }
                    }
                    this.mobileMenuOpen = false;
                },
                handleDesktopLogoClick(event, path) {
                    if (window.location.pathname === path && !window.location.search) {
                        event.preventDefault();
                        event.stopPropagation();
                        window.scrollTo({ top: 0, behavior: 'smooth' });
                    }
                },
                handleResize() {
                    if (window.innerWidth >= 768) {
                        this.mobileMenuOpen = false;
                    }
                }
            };
        });

        Alpine.data('globalError', function () {
            return {
                errorMessage: '',
                errorTime: '',
                showError: false,
                lastErrorTime: null,
                timeoutId: null,
                init: function () {
                    var self = this;
                    window.addEventListener('tivri-error', function (e) {
                        var now = new Date();

                        // Prevent duplicate error spam when already displaying
                        if (self.showError && self.errorMessage === e.detail) {
                            return;
                        }

                        // Throttle error updates to prevent layout flash spam
                        if (self.lastErrorTime && now - self.lastErrorTime < 1500) {
                            return;
                        }

                        self.lastErrorTime = now;
                        var pad = function (num) {
                            return String(num).padStart(2, '0');
                        };
                        self.errorTime =
                            pad(now.getHours()) + ':' + pad(now.getMinutes()) + ':' + pad(now.getSeconds());
                        self.errorMessage = e.detail;
                        self.showError = true;

                        if (self.timeoutId) {
                            clearTimeout(self.timeoutId);
                        }
                        self.timeoutId = setTimeout(function () {
                            self.showError = false;
                        }, 6000);
                    });
                },
                closeError: function () {
                    this.showError = false;
                },
                getErrorClass: function () {
                    return this.showError
                        ? 'opacity-100 translate-y-0 scale-100 pointer-events-auto'
                        : 'opacity-0 translate-y-4 scale-95 pointer-events-none';
                }
            };
        });

        Alpine.data('expandableText', function () {
            return {
                expanded: false,
                toggle() {
                    this.expanded = !this.expanded;
                },
                getExpandedClass() {
                    return this.expanded ? '' : 'hidden';
                },
                getCollapsedClass() {
                    return !this.expanded ? '' : 'hidden';
                }
            };
        });

        Alpine.data('dropdown', function () {
            return {
                open: false,
                toggle() {
                    this.open = !this.open;
                },
                close() {
                    this.open = false;
                },
                getDropdownClass() {
                    return this.open
                        ? 'opacity-100 scale-100 pointer-events-auto'
                        : 'opacity-0 scale-95 pointer-events-none delay-200';
                },
                getIconClass() {
                    return this.open ? 'rotate-180' : '';
                }
            };
        });
    });
}

document.addEventListener(
    'click',
    function (e) {
        const link = e.target.closest('a');
        if (!link) return;

        const href = link.getAttribute('href');
        if (!href) return;

        const isRootHash = href.startsWith('/#') && window.location.pathname === '/';
        const isPureHash = href.startsWith('#');

        if (isRootHash || isPureHash) {
            const hashStr = isRootHash ? href.substring(1) : href;
            if (hashStr && hashStr !== '#') {
                const target = document.querySelector(hashStr);
                if (target) {
                    e.preventDefault();
                    e.stopPropagation(); // Hide the click from HTMX
                    target.scrollIntoView({ behavior: 'smooth' });
                    if (history.pushState) history.pushState(null, null, hashStr);
                }
            }
            return;
        }

        if (
            link.origin === window.location.origin &&
            link.pathname === window.location.pathname &&
            link.search === window.location.search &&
            !link.hash
        ) {
            e.preventDefault();
            e.stopPropagation();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }
    },
    true
); // use capture phase to run before HTMX body listener

initAlpine();
