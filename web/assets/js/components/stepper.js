document.addEventListener('alpine:init', () => {
    Alpine.data('stepper', function(highQueueActive = false) {
        return {
        highQueueActive: highQueueActive,
        openStepper: false,
        totalSteps: 5,
        step: this.$persist(1).as('intake_step').using(sessionStorage),
        budget: this.$persist('').as('intake_budget').using(sessionStorage),
        customBudget: this.$persist('').as('intake_customBudget').using(sessionStorage),
        scopeText: this.$persist('').as('intake_scopeText').using(sessionStorage),
        scopeMax: 2000,
        nameText: this.$persist('').as('intake_nameText').using(sessionStorage),
        nameMax: 150,
        deadlineNeeded: this.$persist(false).as('intake_deadlineNeeded').using(sessionStorage),
        deadlineSpec: this.$persist('').as('intake_deadlineSpec').using(sessionStorage),
        contactEmail: this.$persist('').as('intake_contactEmail').using(sessionStorage),
        contactInfo: this.$persist('').as('intake_contactInfo').using(sessionStorage),
        submitted: this.$persist(false).as('intake_submitted').using(sessionStorage),
        nameTouched: false,
        scopeTouched: false,
        budgetTouched: false,
        emailTouched: false,
        deadlineTouched: false,
        submitStatus: 'idle',
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
            }

            this.$watch('step', val => {
                if (val === 5 && !this.submitted) {
                    this.$nextTick(() => this.renderTurnstile());
                }
            });

            this.$watch('budget', val => {
                if (val !== 'other') {
                    this.customBudget = '';
                }
            });

            this.$watch('deadlineNeeded', val => {
                if (!val) {
                    this.deadlineSpec = '';
                }
            });

            this.$watch('submitted', val => {
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

        handleSubmit(event) {
            if (!this.validateTurnstile(event)) return;

            this.submitStatus = 'submitting';
        }
    };
    });
});
