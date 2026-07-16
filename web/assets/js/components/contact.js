document.addEventListener('alpine:init', () => {
    Alpine.data('contact', () => ({
        showForm: Alpine.$persist(true).as('contact_showForm').using(sessionStorage),
        submitted: Alpine.$persist(false).as('contact_submitted').using(sessionStorage),
        email: Alpine.$persist('').as('contact_email').using(sessionStorage),
        topic: Alpine.$persist('').as('contact_topic').using(sessionStorage),
        message: Alpine.$persist('').as('contact_message').using(sessionStorage),
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
    }));
});
