if ('scrollRestoration' in history) {
    history.scrollRestoration = 'manual';
}

(function() {
    var footerActive = false;

    function updateScrollState() {
        var header = document.getElementById('site-header');
        if (header) {
            if (window.scrollY > 50) {
                header.classList.add('backdrop-blur-lg', 'bg-black/90', 'border-b', 'border-white/[0.08]', 'py-5', 'shadow-[0_4px_30px_rgba(0,0,0,0.8)]');
                header.classList.remove('bg-transparent', 'border-transparent', 'py-10');
            } else {
                header.classList.remove('backdrop-blur-lg', 'bg-black/90', 'border-b', 'border-white/[0.08]', 'py-5', 'shadow-[0_4px_30px_rgba(0,0,0,0.8)]');
                header.classList.add('bg-transparent', 'border-transparent', 'py-10');
            }
        }

        var footer = document.getElementById('site-footer');
        if (footer) {
            var scrollY = window.pageYOffset || window.scrollY;
            var maxScroll = document.documentElement.scrollHeight - window.innerHeight;
            var threshold = footerActive ? 155 : 135;
            var isAtBottom = (maxScroll - scrollY <= threshold);

            if (isAtBottom && !footerActive) {
                footerActive = true;
                footer.classList.add('backdrop-blur-lg', 'bg-black/90', 'border-white/[0.08]', 'pt-10', 'pb-6', 'shadow-[0_-4px_30px_rgba(0,0,0,0.8)]');
                footer.classList.remove('bg-transparent', 'border-transparent', 'pt-16', 'pb-12');
            } else if (!isAtBottom && footerActive) {
                footerActive = false;
                footer.classList.remove('backdrop-blur-lg', 'bg-black/90', 'border-white/[0.08]', 'pt-10', 'pb-6', 'shadow-[0_-4px_30px_rgba(0,0,0,0.8)]');
                footer.classList.add('bg-transparent', 'border-transparent', 'pt-16', 'pb-12');
            }
        }
    }

    window.addEventListener('scroll', updateScrollState);

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', updateScrollState);
    } else {
        updateScrollState();
    }

    document.addEventListener('htmx:beforeSwap', function(evt) {
        document.documentElement.classList.add('no-transition');
        sessionStorage.setItem('tivri_htmx_nav', 'true');

        if (evt.detail.serverResponse) {
            try {
                var parser = new DOMParser();
                var doc = parser.parseFromString(evt.detail.serverResponse, 'text/html');
                var oldHeader = document.getElementById('site-header');
                var newHeader = doc.getElementById('site-header');
                if (oldHeader && newHeader) {
                    newHeader.className = oldHeader.className;
                }

                var oldFooter = document.getElementById('site-footer');
                var newFooter = doc.getElementById('site-footer');
                if (oldFooter && newFooter) {
                    newFooter.className = oldFooter.className;
                }

                evt.detail.serverResponse = doc.documentElement.outerHTML;
            } catch (e) {
                console.error('Failed to parse server response:', e);
            }
        }
    });

    document.addEventListener('htmx:afterSwap', function() {
        updateScrollState();
        setTimeout(updateScrollState, 50);
        setTimeout(function() {
            document.documentElement.classList.remove('no-transition');
        }, 100);
    });
})();

(function() {
    function initNavObserver() {
        var sections = ['about', 'benefits', 'skills', 'portfolio', 'contact'];
        var navLinks = document.querySelectorAll('#main-nav .nav-link');

        function clearActive() {
            navLinks.forEach(function(link) {
                link.classList.remove('nav-active');
            });
        }

        function setActive(sectionId) {
            clearActive();
            navLinks.forEach(function(link) {
                if (link.getAttribute('href') === '/#' + sectionId) {
                    link.classList.add('nav-active');
                }
            });
        }

        var observerOptions = {
            root: null,
            rootMargin: '-20% 0px -60% 0px',
            threshold: 0
        };

        var observer = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    setActive(entry.target.id);
                }
            });
        }, observerOptions);

        sections.forEach(function(id) {
            var el = document.getElementById(id);
            if (el) { observer.observe(el); }
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initNavObserver);
    } else {
        initNavObserver();
    }

    document.addEventListener('htmx:afterSwap', initNavObserver);

    document.addEventListener('htmx:responseError', function(evt) {
        var errorText = evt.detail.xhr.responseText || 'An error occurred during submission.';
        window.dispatchEvent(new CustomEvent('tivri-error', { detail: errorText }));
    });
})();

document.addEventListener('alpine:init', function() {
    Alpine.data('layout', function() {
        return {
            mobileMenuOpen: false,
            openStepper: Alpine.$persist(false).as('openStepper').using(sessionStorage)
        };
    });

    Alpine.data('globalError', function() {
        return {
            errorMessage: '',
            showError: false,
            init: function() {
                var self = this;
                window.addEventListener('tivri-error', function(e) {
                    self.errorMessage = e.detail;
                    self.showError = true;
                    setTimeout(function() {
                        self.showError = false;
                    }, 6000);
                });
            }
        };
    });

    Alpine.data('dropdown', function() {
        return {
            open: false
        };
    });

});

window.tivriTurnstileMixin = function() {
    return {
        turnstileToken: '',
        turnstileId: null,
        isVerified: false,

        destroy() {
            if (window.turnstile && this.turnstileId !== null) {
                try { window.turnstile.remove(this.turnstileId); } catch(e) {}
            }
        },

        initTurnstileListeners() {
            document.addEventListener('htmx:responseError', () => {
                this.isVerified = false;
                this.turnstileToken = '';
                if (window.turnstile && this.turnstileId !== null) {
                    window.turnstile.reset(this.turnstileId);
                }
            });
        },

        renderTurnstile() {
            if (window.tivriTurnstileSiteKey && window.turnstile && this.$refs.turnstileContainer && this.turnstileId === null) {
                try {
                    this.turnstileId = window.turnstile.render(this.$refs.turnstileContainer, {
                        sitekey: window.tivriTurnstileSiteKey,
                        theme: 'dark',
                        size: 'normal',
                        language: document.documentElement.lang || 'en',
                        callback: (token) => {
                            this.turnstileToken = token;
                            this.isVerified = true;
                        },
                        'expired-callback': () => {
                            this.turnstileToken = '';
                            this.isVerified = false;
                        },
                        'error-callback': () => {
                            this.isVerified = false;
                            this.turnstileToken = '';
                            window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Security verification failed.' }));
                            return true;
                        }
                    });
                } catch (e) {
                    console.error('Turnstile render failed:', e);
                }
            }
        },

        validateTurnstile(event) {
            if (window.tivriTurnstileSiteKey && window.turnstile && !this.isVerified) {
                if (event) {
                    event.preventDefault();
                    event.stopPropagation();
                }
                window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Please complete the security check.' }));
                return false;
            }
            return true;
        },

        resetTurnstile() {
            this.turnstileToken = '';
            this.isVerified = false;
            if (window.turnstile && this.turnstileId !== null) {
                window.turnstile.reset(this.turnstileId);
            }
        }
    };
};
