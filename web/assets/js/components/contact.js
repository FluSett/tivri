document.addEventListener('alpine:init', () => {
    Alpine.data('contact', function() {
        return {
        showForm: this.$persist(true).as('contact_showForm').using(sessionStorage),
        submitted: this.$persist(false).as('contact_submitted').using(sessionStorage),
        email: this.$persist('').as('contact_email').using(sessionStorage),
        topic: this.$persist('').as('contact_topic').using(sessionStorage),
        message: this.$persist('').as('contact_message').using(sessionStorage),
        emailTouched: false,
        topicTouched: false,
        messageTouched: false,
        submitStatus: 'idle',
        turnstileToken: '',
        turnstileId: null,
        isVerified: false,

        init() {
            this.$watch('showForm', val => {
                if (val && !this.submitted) {
                    this.$nextTick(() => this.renderTurnstile());
                }
            });

            this.$watch('submitted', val => {
                if (!val && this.showForm) {
                    this.$nextTick(() => {
                        this.turnstileId = null;
                        this.renderTurnstile();
                    });
                }
            });

            if (this.showForm && !this.submitted) {
                this.$nextTick(() => this.renderTurnstile());
            }

            document.addEventListener('htmx:responseError', () => {
                this.submitStatus = 'idle';
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
                            this.submitStatus = 'idle';
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

        handleSubmit(event) {
            if (window.tivriTurnstileSiteKey && window.turnstile && !this.isVerified) {
                event.preventDefault();
                event.stopPropagation();
                window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Please complete the security check.' }));
                return;
            }

            this.submitStatus = 'submitting';
        },

        resetForm() {
            this.showForm = true;
            this.submitted = false;
            this.email = '';
            this.topic = '';
            this.message = '';
            this.emailTouched = false;
            this.topicTouched = false;
            this.messageTouched = false;
            this.submitStatus = 'idle';
            this.turnstileToken = '';
            this.isVerified = false;

            if (window.turnstile && this.turnstileId !== null) {
                window.turnstile.reset(this.turnstileId);
            }
        }
    };
    });
});
