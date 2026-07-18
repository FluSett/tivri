import { createReactiveState, bindRefs, delegate } from './state.js';

export function initGlobalError() {
    const container = document.getElementById('global-error-toast');
    if (!container) return;

    const refs = bindRefs(container);

    let timeoutId = null;

    const state = createReactiveState(
        'error',
        {
            showError: false,
            message: '',
            time: '',
            lastErrorTime: null
        },
        (newState) => {
            updateUI();
        }
    );

    function updateUI() {
        if (state.showError) {
            container.classList.remove('opacity-0', 'translate-y-4', 'scale-95', 'pointer-events-none');
            container.classList.add('opacity-100', 'translate-y-0', 'scale-100', 'pointer-events-auto');
        } else {
            container.classList.remove('opacity-100', 'translate-y-0', 'scale-100', 'pointer-events-auto');
            container.classList.add('opacity-0', 'translate-y-4', 'scale-95', 'pointer-events-none');
        }

        refs.time.textContent = state.time;
        refs.message.textContent = state.message;
    }

    const closeHandler = () => {
        state.showError = false;
    };

    delegate(container, 'click', '[data-ref="close"]', closeHandler);

    const errorHandler = (e) => {
        const now = new Date();
        const msg = e.detail;

        if (state.showError && state.message === msg) return;
        if (state.lastErrorTime && now - state.lastErrorTime < 1500) return;

        const pad = (num) => String(num).padStart(2, '0');

        state.lastErrorTime = now;
        state.message = msg;
        state.time = pad(now.getHours()) + ':' + pad(now.getMinutes()) + ':' + pad(now.getSeconds());
        state.showError = true;

        if (timeoutId) clearTimeout(timeoutId);
        timeoutId = setTimeout(() => {
            state.showError = false;
        }, 6000);
    };

    window.addEventListener('tivri-error', errorHandler);

    updateUI();

    return () => {
        window.removeEventListener('tivri-error', errorHandler);
        if (timeoutId) clearTimeout(timeoutId);
    };
}
