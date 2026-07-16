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
        ...window.tivriTurnstileMixin(),

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

            this.initTurnstileListeners();
        },

        handleSubmit(event) {
            if (!this.validateTurnstile(event)) return;
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
            this.resetTurnstile();
        }
    };
    });
});
