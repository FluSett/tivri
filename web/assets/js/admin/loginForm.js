export function loginForm() {
    return {
        username: Alpine.$persist('').as('adm_log_user'),
        password: Alpine.$persist('').as('adm_log_pass'),
        ...(window.tivriTurnstileMixin ? window.tivriTurnstileMixin() : {}),
        init() {
            window.tivriHandleLocaleChange(() => {
                this.username = '';
                this.password = '';
            });
            if (this.initTurnstileListeners) {
                this.initTurnstileListeners();
                setTimeout(() => this.renderTurnstile(), 50);
            }
        }
    };
}
