document.addEventListener('alpine:init', () => {
    Alpine.data('contact', () => ({
        showForm: sessionStorage.getItem('contact_showForm') === 'true',
        submitted: sessionStorage.getItem('contact_submitted') === 'true',
        email: sessionStorage.getItem('contact_email') || '',
        topic: sessionStorage.getItem('contact_topic') || '',
        message: sessionStorage.getItem('contact_message') || '',
        emailTouched: false,
        topicTouched: false,
        messageTouched: false,

        init() {
            this.$watch('showForm', val => sessionStorage.setItem('contact_showForm', val));

            this.$watch('email', val => sessionStorage.setItem('contact_email', val || ''));

            this.$watch('topic', val => sessionStorage.setItem('contact_topic', val || ''));

            this.$watch('message', val => sessionStorage.setItem('contact_message', val || ''));

            this.$watch('submitted', val => sessionStorage.setItem('contact_submitted', val));
        },

        resetForm() {
            this.showForm = false;
            this.submitted = false;
            this.email = '';
            this.topic = '';
            this.message = '';
            this.emailTouched = false;
            this.topicTouched = false;
            this.messageTouched = false;

            sessionStorage.removeItem('contact_showForm');
            sessionStorage.removeItem('contact_email');
            sessionStorage.removeItem('contact_topic');
            sessionStorage.removeItem('contact_message');
            sessionStorage.removeItem('contact_submitted');
        }
    }));
});
