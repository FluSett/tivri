import { createReactiveState, bindRefs, delegate } from './state.js';

export function initGlobalSuccess() {
    const container = document.getElementById('global-success-toast');
    if (!container) return;

    const refs = bindRefs(container);

    let timeoutId = null;

    const state = createReactiveState(
        'success',
        {
            showSuccess: false,
            message: '',
            title: 'Success',
            time: '',
            lastSuccessTime: null
        },
        (newState) => {
            updateUI();
        }
    );

    function updateUI() {
        if (state.showSuccess) {
            container.classList.remove('opacity-0', 'translate-y-4', 'scale-95', 'pointer-events-none');
            container.classList.add('opacity-100', 'translate-y-0', 'scale-100', 'pointer-events-auto');
        } else {
            container.classList.remove('opacity-100', 'translate-y-0', 'scale-100', 'pointer-events-auto');
            container.classList.add('opacity-0', 'translate-y-4', 'scale-95', 'pointer-events-none');
        }

        refs.time.textContent = state.time;
        if (refs.title) refs.title.textContent = state.title;
        refs.message.textContent = state.message;
    }

    const closeHandler = () => {
        state.showSuccess = false;
    };

    delegate(container, 'click', '[data-ref="close"]', closeHandler);

    const successHandler = (e) => {
        const now = new Date();
        const payload = e.detail;
        const msg = typeof payload === 'string' ? payload : payload.message;
        const title = typeof payload === 'string' ? 'Success' : payload.title || 'Success';

        if (state.showSuccess && state.message === msg) return;
        if (state.lastSuccessTime && now - state.lastSuccessTime < 1500) return;

        const pad = (num) => String(num).padStart(2, '0');

        state.lastSuccessTime = now;
        state.title = title;
        state.message = msg;
        state.time = pad(now.getHours()) + ':' + pad(now.getMinutes()) + ':' + pad(now.getSeconds());
        state.showSuccess = true;

        if (timeoutId) clearTimeout(timeoutId);
        timeoutId = setTimeout(() => {
            state.showSuccess = false;
        }, 6000);
    };

    window.addEventListener('tivri-success', successHandler);

    updateUI();

    return () => {
        window.removeEventListener('tivri-success', successHandler);
        if (timeoutId) clearTimeout(timeoutId);
    };
}
