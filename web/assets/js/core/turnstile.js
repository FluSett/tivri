// Initialize Turnstile configuration from data attribute and fallback error handler
window.tivriTurnstileSiteKey = document.body ? document.body.dataset.turnstileSitekey : '';
window.tivriTurnstileOnError = function (code) {
    console.warn('Turnstile challenge error:', code);
    return true;
};

window.tivriTurnstileMixin = function () {
    return {
        turnstileToken: '',
        turnstileId: null,
        isVerified: false,

        destroy() {
            if (window.turnstile && this.turnstileId !== null) {
                try {
                    window.turnstile.remove(this.turnstileId);
                } catch (e) {}
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
            if (
                window.tivriTurnstileSiteKey &&
                window.turnstile &&
                this.$refs.turnstileContainer &&
                this.turnstileId === null
            ) {
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
                            window.dispatchEvent(
                                new CustomEvent('tivri-error', { detail: 'Security verification failed.' })
                            );
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
