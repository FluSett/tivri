import { createReactiveState, bindRefs } from '../core/state.js';
import { initTurnstile, destroyTurnstile } from '../core/turnstile.js';
import { toggleClasses } from '../core/dom.js';

// Removed STORAGE_MAP

export function initLoginForm() {
    const refs = bindRefs(document.body);
    if (!refs.form) return;

    const state = createReactiveState(
        'login_form',
        {
            username: '',
            password: '',
            turnstileToken: '',
            turnstileId: null,
            isVerified: false
        },
        { ephemeralKeys: ['turnstileId', 'turnstileToken', 'isVerified'] },
        (newState) => {
            if (refs.usernameInput && refs.usernameInput.value !== newState.username) {
                refs.usernameInput.value = newState.username;
            }
            if (refs.passwordInput && refs.passwordInput.value !== newState.password) {
                refs.passwordInput.value = newState.password;
            }

            if (refs.submitBtn) {
                const valid = newState.username.trim() && newState.password.trim();
                refs.submitBtn.disabled = !valid;
                toggleClasses(
                    refs.submitBtn,
                    valid,
                    ['hover:bg-primary/80', 'hover:shadow-[0_0_15px_rgba(255,51,102,0.4)]', 'cursor-pointer'],
                    ['opacity-50', 'cursor-not-allowed']
                );
            }
        }
    );

    function renderTurnstile() {
        if (state.turnstileId !== null) return;

        state.turnstileId = initTurnstile(
            refs.turnstileContainer,
            (token) => {
                state.turnstileToken = token;
                state.isVerified = true;
                if (refs.turnstileInput) refs.turnstileInput.value = token;
            },
            () => {
                state.turnstileToken = '';
                state.isVerified = false;
                if (refs.turnstileInput) refs.turnstileInput.value = '';
            },
            () => {
                state.isVerified = false;
                state.turnstileToken = '';
                if (refs.turnstileInput) refs.turnstileInput.value = '';
                window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Security verification failed.' }));
                return true;
            }
        );
    }

    const teardowns = [];

    const handleUsernameInput = (e) => {
        state.username = e.target.value;
    };
    if (refs.usernameInput) {
        refs.usernameInput.addEventListener('input', handleUsernameInput);
        teardowns.push(() => refs.usernameInput.removeEventListener('input', handleUsernameInput));
    }

    const handlePasswordInput = (e) => {
        state.password = e.target.value;
    };
    if (refs.passwordInput) {
        refs.passwordInput.addEventListener('input', handlePasswordInput);
        teardowns.push(() => refs.passwordInput.removeEventListener('input', handlePasswordInput));
    }

    const handleSubmit = (e) => {
        if (window.tivriTurnstileSiteKey && window.turnstile && !state.isVerified) {
            e.preventDefault();
            e.stopPropagation();
            window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Please complete the security check.' }));
            return;
        }
    };
    refs.form.addEventListener('submit', handleSubmit);
    teardowns.push(() => refs.form.removeEventListener('submit', handleSubmit));

    state.username = state.username; // trigger proxy
    setTimeout(renderTurnstile, 50);

    return () => {
        destroyTurnstile(state.turnstileId);
        teardowns.forEach((fn) => fn());
    };
}
