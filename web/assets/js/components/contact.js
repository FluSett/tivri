document.addEventListener('alpine:init', () => {
    Alpine.data('contact', function () {
        return {
            showForm: Alpine.$persist(false).as('dm_show'),
            submitted: Alpine.$persist(false).as('dm_submitted'),
            email: Alpine.$persist('').as('dm_email'),
            topic: Alpine.$persist('').as('dm_topic'),
            message: Alpine.$persist('').as('dm_message'),
            emailTouched: Alpine.$persist(false).as('dm_emailTouched'),
            topicTouched: Alpine.$persist(false).as('dm_topicTouched'),
            messageTouched: Alpine.$persist(false).as('dm_messageTouched'),
            submitStatus: Alpine.$persist('idle').as('dm_submitStatus'),
            ...window.tivriTurnstileMixin(),

            init() {
                window.tivriHandleLocaleChange(() => {
                    this.resetForm();
                    this.showForm = false;
                });

                this.$watch('showForm', (val) => {
                    if (val && !this.submitted) {
                        this.$nextTick(() => this.renderTurnstile());
                    }
                });

                this.$watch('submitted', (val) => {
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
