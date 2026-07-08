document.addEventListener('alpine:init', () => {
    Alpine.data('stepper', (highQueueActive = false) => ({
        highQueueActive: highQueueActive,
        step: parseInt(sessionStorage.getItem('intake_step')) || 1,
        totalSteps: 5,
        budget: sessionStorage.getItem('intake_budget') || '',
        customBudget: sessionStorage.getItem('intake_customBudget') || '',
        scopeText: sessionStorage.getItem('intake_scopeText') || '',
        scopeMax: 2000,
        nameText: sessionStorage.getItem('intake_nameText') || '',
        nameMax: 150,
        deadlineNeeded: sessionStorage.getItem('intake_deadlineNeeded') === 'true',
        deadlineSpec: sessionStorage.getItem('intake_deadlineSpec') || '',
        contactEmail: sessionStorage.getItem('intake_contactEmail') || '',
        contactInfo: sessionStorage.getItem('intake_contactInfo') || '',
        submitted: sessionStorage.getItem('intake_submitted') === 'true',
        nameTouched: false,
        scopeTouched: false,
        budgetTouched: false,
        emailTouched: false,
        deadlineTouched: false,
        submitStatus: 'idle',
        turnstileToken: '',
        turnstileId: null,
        isVerified: false,

        get scopeRemaining() {
            return this.scopeMax - this.scopeText.length;
        },

        get nameRemaining() {
            return this.nameMax - this.nameText.length;
        },

        get budgetValue() {
            if (this.budget === 'other') {
                return this.customBudget;
            }
            return this.budget;
        },

        canGoNext(currentStep) {
            if (currentStep === 1) {
                return this.nameText.trim().length >= 2;
            }

            if (currentStep === 2) {
                return this.scopeText.trim().length >= 20;
            }

            if (currentStep === 3) {
                if (this.deadlineNeeded) {
                    return this.deadlineSpec.trim().length >= 2;
                }
                return true;
            }

            if (currentStep === 4) {
                if (this.budget === '') {
                    return false;
                }
                if (this.budget === 'other') {
                    return this.customBudget.trim() !== '' && !isNaN(this.customBudget) && parseInt(this.customBudget) >= 100;
                }
                return true;
            }

            return true;
        },

        init() {
            if (this.highQueueActive) {
                this.deadlineNeeded = false;
                this.deadlineSpec = '';
                sessionStorage.removeItem('intake_deadlineNeeded');
                sessionStorage.removeItem('intake_deadlineSpec');
            }

            this.$watch('step', val => {
                sessionStorage.setItem('intake_step', val);
                if (val === 5 && !this.submitted) {
                    this.$nextTick(() => this.renderTurnstile());
                }
            });

            this.$watch('budget', val => {
                sessionStorage.setItem('intake_budget', val);
                if (val !== 'other') {
                    this.customBudget = '';
                    sessionStorage.removeItem('intake_customBudget');
                }
            });

            this.$watch('customBudget', val => sessionStorage.setItem('intake_customBudget', val));
            this.$watch('scopeText', val => sessionStorage.setItem('intake_scopeText', val));
            this.$watch('nameText', val => sessionStorage.setItem('intake_nameText', val));

            this.$watch('deadlineNeeded', val => {
                sessionStorage.setItem('intake_deadlineNeeded', val);
                if (!val) {
                    this.deadlineSpec = '';
                    sessionStorage.removeItem('intake_deadlineSpec');
                }
            });

            this.$watch('deadlineSpec', val => sessionStorage.setItem('intake_deadlineSpec', val));
            this.$watch('contactEmail', val => sessionStorage.setItem('intake_contactEmail', val));
            this.$watch('contactInfo', val => sessionStorage.setItem('intake_contactInfo', val));

            this.$watch('submitted', val => {
                sessionStorage.setItem('intake_submitted', val);
                if (val) {
                    sessionStorage.setItem('openStepper', 'true');
                }
            });

            if (this.step === 5 && !this.submitted) {
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

        resetForm() {
            this.step = 1;
            this.budget = '';
            this.customBudget = '';
            this.scopeText = '';
            this.nameText = '';
            this.deadlineNeeded = false;
            this.deadlineSpec = '';
            this.contactEmail = '';
            this.contactInfo = '';
            this.submitted = false;
            this.nameTouched = false;
            this.scopeTouched = false;
            this.budgetTouched = false;
            this.emailTouched = false;
            this.deadlineTouched = false;
            this.submitStatus = 'idle';
            this.turnstileToken = '';
            this.isVerified = false;

            if (window.turnstile && this.turnstileId !== null) {
                window.turnstile.reset(this.turnstileId);
            }

            sessionStorage.removeItem('intake_step');
            sessionStorage.removeItem('intake_budget');
            sessionStorage.removeItem('intake_customBudget');
            sessionStorage.removeItem('intake_scopeText');
            sessionStorage.removeItem('intake_nameText');
            sessionStorage.removeItem('intake_deadlineNeeded');
            sessionStorage.removeItem('intake_deadlineSpec');
            sessionStorage.removeItem('intake_contactEmail');
            sessionStorage.removeItem('intake_contactInfo');
            sessionStorage.removeItem('intake_submitted');
            sessionStorage.removeItem('openStepper');

            this.openStepper = false;
            document.getElementById('intake-form').reset();
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
        }
    }));
});
