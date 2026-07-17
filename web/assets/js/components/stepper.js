document.addEventListener('alpine:init', () => {
    Alpine.data('stepper', function (highQueueActive = false) {
        return {
            highQueueActive: highQueueActive,
            openStepper: Alpine.$persist(false).as('st_open'),
            totalSteps: 5,
            step: Alpine.$persist(1).as('st_step'),
            budget: Alpine.$persist('').as('st_budget'),
            customBudget: Alpine.$persist('').as('st_customBudget'),
            scopeText: Alpine.$persist('').as('st_scope'),
            scopeMax: 2000,
            nameText: Alpine.$persist('').as('st_name'),
            nameMax: 150,
            deadlineNeeded: Alpine.$persist(false).as('st_deadlineNeeded'),
            deadlineSpec: Alpine.$persist('').as('st_deadlineSpec'),
            contactEmail: Alpine.$persist('').as('st_email'),
            contactInfo: Alpine.$persist('').as('st_info'),
            submitted: Alpine.$persist(false).as('st_submitted'),
            nameTouched: Alpine.$persist(false).as('st_nameTouched'),
            scopeTouched: Alpine.$persist(false).as('st_scopeTouched'),
            budgetTouched: Alpine.$persist(false).as('st_budgetTouched'),
            emailTouched: Alpine.$persist(false).as('st_emailTouched'),
            deadlineTouched: Alpine.$persist(false).as('st_deadlineTouched'),
            submitStatus: Alpine.$persist('idle').as('st_submitStatus'),
            transitionWizard: {
                ['x-transition:enter']: 'transition ease-out duration-500 delay-200',
                ['x-transition:enter-start']: 'opacity-0 translate-x-4',
                ['x-transition:enter-end']: 'opacity-100 translate-x-0',
                ['x-transition:leave']: 'transition ease-in duration-200',
                ['x-transition:leave-start']: 'opacity-100 translate-x-0',
                ['x-transition:leave-end']: 'opacity-0 -translate-x-4'
            },
            ...window.tivriTurnstileMixin(),

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
                        return (
                            this.customBudget.trim() !== '' &&
                            !isNaN(this.customBudget) &&
                            parseInt(this.customBudget) >= 100
                        );
                    }
                    return true;
                }

                return true;
            },

            init() {
                window.tivriHandleLocaleChange(() => {
                    this.resetForm();
                });

                if (this.highQueueActive) {
                    this.deadlineNeeded = false;
                    this.deadlineSpec = '';
                }

                this.$watch('step', (val) => {
                    if (val === 5 && !this.submitted) {
                        this.$nextTick(() => this.renderTurnstile());
                    }
                });

                this.$watch('budget', (val) => {
                    if (val !== 'other') {
                        this.customBudget = '';
                    }
                });

                this.$watch('deadlineNeeded', (val) => {
                    if (!val) {
                        this.deadlineSpec = '';
                    }
                });

                this.$watch('submitted', (val) => {
                    if (val) {
                        this.openStepper = true;
                    }
                });

                if (this.step === 5 && !this.submitted) {
                    this.$nextTick(() => this.renderTurnstile());
                }

                this.initTurnstileListeners();
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
                this.resetTurnstile();

                this.openStepper = false;
                document.getElementById('intake-form').reset();
            },

            openStepperForm() {
                this.openStepper = true;
            },

            setStep(n) {
                this.step = n;
            },

            setBudget(b) {
                this.budget = b;
            },

            setDeadlineNeeded(val) {
                if (!this.highQueueActive) {
                    this.deadlineNeeded = val;
                }
            },

            getStepperIntroClass() {
                return !this.openStepper
                    ? 'grid-collapse grid-expand'
                    : 'grid-collapse opacity-0 pointer-events-none delay-300';
            },

            getStepperFormClass() {
                return this.openStepper
                    ? 'grid-collapse grid-expand'
                    : 'grid-collapse opacity-0 pointer-events-none delay-300';
            },

            getSubmittedClass() {
                return this.submitted
                    ? 'opacity-100 scale-100 pointer-events-auto'
                    : 'opacity-0 scale-95 pointer-events-none hidden';
            },

            getNotSubmittedClass() {
                return !this.submitted
                    ? 'opacity-100 scale-100 pointer-events-auto'
                    : 'opacity-0 scale-95 pointer-events-none hidden';
            },

            getBudgetValidationClass() {
                if (this.budgetTouched && (this.customBudget.trim() === '' || parseInt(this.customBudget, 10) < 100)) {
                    return 'opacity-100';
                }
                return 'opacity-0 hidden';
            },

            handleBudgetInput(event) {
                this.customBudget = this.customBudget.replace(/[^0-9]/g, '');
                this.budgetTouched = true;
            },

            handleHtmxAfterRequest(event) {
                if (event.detail.successful) {
                    this.submitted = true;
                }
                this.submitStatus = 'idle';
            },

            handleSubmit(event) {
                if (!this.validateTurnstile(event)) return;

                this.submitStatus = 'submitting';
            }
        };
    });
});
