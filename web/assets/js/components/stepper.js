document.addEventListener('alpine:init', () => {
    Alpine.data('stepper', () => ({
        step: parseInt(sessionStorage.getItem('intake_step')) || 1,
        totalSteps: 4,
        budget: sessionStorage.getItem('intake_budget') || '',
        customBudget: sessionStorage.getItem('intake_customBudget') || '',
        scopeText: sessionStorage.getItem('intake_scopeText') || '',
        scopeMax: 2000,
        nameText: sessionStorage.getItem('intake_nameText') || '',
        nameMax: 150,
        contactEmail: sessionStorage.getItem('intake_contactEmail') || '',
        contactPhone: sessionStorage.getItem('intake_contactPhone') || '',
        submitted: sessionStorage.getItem('intake_submitted') === 'true',
        nameTouched: false,
        scopeTouched: false,
        budgetTouched: false,
        emailTouched: false,

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
            this.$watch('step', val => sessionStorage.setItem('intake_step', val));

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

            this.$watch('contactEmail', val => sessionStorage.setItem('intake_contactEmail', val));

            this.$watch('contactPhone', val => sessionStorage.setItem('intake_contactPhone', val));

            this.$watch('submitted', val => {
                sessionStorage.setItem('intake_submitted', val);

                if (val) {
                    sessionStorage.setItem('openStepper', 'true');
                }
            });
        },

        resetForm() {
            this.step = 1;
            this.budget = '';
            this.customBudget = '';
            this.scopeText = '';
            this.nameText = '';
            this.contactEmail = '';
            this.contactPhone = '';
            this.submitted = false;
            this.nameTouched = false;
            this.scopeTouched = false;
            this.budgetTouched = false;
            this.emailTouched = false;

            sessionStorage.removeItem('intake_step');
            sessionStorage.removeItem('intake_budget');
            sessionStorage.removeItem('intake_customBudget');
            sessionStorage.removeItem('intake_scopeText');
            sessionStorage.removeItem('intake_nameText');
            sessionStorage.removeItem('intake_contactEmail');
            sessionStorage.removeItem('intake_contactPhone');
            sessionStorage.removeItem('intake_submitted');
            sessionStorage.removeItem('openStepper');

            this.openStepper = false;
            document.getElementById('intake-form').reset();
        }
    }));
});
