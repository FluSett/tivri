import { createReactiveState, bindRefs, delegate } from '../core/state.js';
import { initTurnstile, resetTurnstile, destroyTurnstile } from '../core/turnstile.js';
import { isValidEmail } from '../core/validators.js';
import { toggleVisibility, toggleClasses, setButtonSubmittingState } from '../core/dom.js';

const MIN_TOPIC_LEN = 3;
const MIN_MSG_LEN = 10;

// Removed STORAGE_MAP

export function initContact() {
    const container = document.getElementById('contact-container');
    if (!container) return;

    const refs = bindRefs(container);

    const state = createReactiveState(
        'contact',
        {
            showForm: false,
            submitted: false,
            email: '',
            topic: '',
            message: '',
            emailTouched: false,
            topicTouched: false,
            messageTouched: false,
            submitStatus: 'idle',
            turnstileToken: '',
            turnstileId: null,
            isVerified: false
        },
        { ephemeralKeys: ['turnstileId', 'turnstileToken', 'isVerified'] },
        (newState) => {
            updateUI();
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

    function doResetTurnstile() {
        state.turnstileToken = '';
        state.isVerified = false;
        if (refs.turnstileInput) refs.turnstileInput.value = '';
        resetTurnstile(state.turnstileId);
    }

    function isFormValid() {
        const eValid = isValidEmail(state.email);
        const tValid = state.topic.trim().length >= MIN_TOPIC_LEN;
        const mValid = state.message.trim().length >= MIN_MSG_LEN;
        return eValid && tValid && mValid && (!window.tivriTurnstileSiteKey || state.isVerified);
    }

    function updateUI() {
        toggleClasses(refs.toggleIcon, state.showForm, ['rotate-180', 'text-white'], []);

        toggleClasses(
            refs.formContainer,
            state.showForm,
            ['grid-expand'],
            ['opacity-0', 'pointer-events-none', 'delay-300']
        );
        if (state.showForm && !state.submitted) setTimeout(renderTurnstile, 300);

        toggleClasses(
            refs.submittedView,
            state.submitted,
            ['opacity-100', 'scale-100', 'pointer-events-auto'],
            ['opacity-0', 'scale-90', 'pointer-events-none', 'hidden']
        );
        toggleClasses(
            refs.notSubmittedView,
            !state.submitted,
            ['opacity-100', 'scale-100', 'pointer-events-auto'],
            ['opacity-0', 'scale-90', 'pointer-events-none', 'hidden']
        );

        const eValid = isValidEmail(state.email);
        toggleClasses(refs.emailError, state.emailTouched && !eValid, ['opacity-100'], ['opacity-0', 'hidden']);

        const tValid = state.topic.trim().length >= MIN_TOPIC_LEN;
        toggleClasses(refs.topicError, state.topicTouched && !tValid, ['opacity-100'], ['opacity-0', 'hidden']);

        const mValid = state.message.trim().length >= MIN_MSG_LEN;
        toggleClasses(refs.messageError, state.messageTouched && !mValid, ['opacity-100'], ['opacity-0', 'hidden']);

        refs.messageCount.textContent = state.message.length + ' / 1000';

        setButtonSubmittingState(
            refs.submitBtn,
            refs.submitText,
            refs.submitSpinner,
            state.submitStatus !== 'idle',
            !isFormValid()
        );
    }

    const teardowns = [];

    teardowns.push(
        delegate(container, 'click', '#contact-toggle-btn', () => {
            state.showForm = !state.showForm;
        })
    );

    teardowns.push(
        delegate(container, 'click', '#contact-send-another-btn', () => {
            state.showForm = true;
            state.submitted = false;
            state.email = '';
            state.topic = '';
            state.message = '';
            state.emailTouched = false;
            state.topicTouched = false;
            state.messageTouched = false;
            state.submitStatus = 'idle';
            refs.email.value = '';
            refs.topic.value = '';
            refs.message.value = '';
            refs.message.style.height = 'auto';
            doResetTurnstile();
            setTimeout(renderTurnstile, 100);
        })
    );

    function debounce(func, wait) {
        let timeout;
        return function (...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }

    refs.email.value = state.email;
    const onEmailInput = debounce(() => {
        state.emailTouched = true;
    }, 500);
    teardowns.push(
        delegate(container, 'input', '#contact-email', (e) => {
            state.email = e.target.value;
            onEmailInput();
        })
    );
    teardowns.push(
        delegate(container, 'blur', '#contact-email', () => {
            state.emailTouched = true;
        })
    );

    refs.topic.value = state.topic;
    const onTopicInput = debounce(() => {
        state.topicTouched = true;
    }, 500);
    teardowns.push(
        delegate(container, 'input', '#contact-topic', (e) => {
            state.topic = e.target.value;
            onTopicInput();
        })
    );
    teardowns.push(
        delegate(container, 'blur', '#contact-topic', () => {
            state.topicTouched = true;
        })
    );

    refs.message.value = state.message;
    const onMessageInput = debounce(() => {
        state.messageTouched = true;
    }, 500);
    teardowns.push(
        delegate(container, 'input', '#contact-message', (e) => {
            state.message = e.target.value;
            e.target.style.height = 'auto';
            e.target.style.height = e.target.scrollHeight + 'px';
            onMessageInput();
        })
    );
    teardowns.push(
        delegate(container, 'blur', '#contact-message', () => {
            state.messageTouched = true;
        })
    );

    const formHandler = (e) => {
        if (window.tivriTurnstileSiteKey && window.turnstile && !state.isVerified) {
            e.preventDefault();
            e.stopPropagation();
            window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Please complete the security check.' }));
            return;
        }
        state.submitStatus = 'submitting';
    };
    refs.form.addEventListener('submit', formHandler);
    teardowns.push(() => refs.form.removeEventListener('submit', formHandler));

    const htmxAfterRequest = (e) => {
        if (e.detail.successful) {
            state.submitted = true;
            state.email = '';
            state.topic = '';
            state.message = '';
            refs.email.value = '';
            refs.topic.value = '';
            refs.message.value = '';
        }
        state.submitStatus = 'idle';
    };
    refs.form.addEventListener('htmx:afterRequest', htmxAfterRequest);
    teardowns.push(() => refs.form.removeEventListener('htmx:afterRequest', htmxAfterRequest));

    const htmxResponseError = () => {
        state.submitStatus = 'idle';
        doResetTurnstile();
    };
    document.addEventListener('htmx:responseError', htmxResponseError);
    teardowns.push(() => document.removeEventListener('htmx:responseError', htmxResponseError));

    updateUI();

    return () => {
        destroyTurnstile(state.turnstileId);
        teardowns.forEach((fn) => fn());
    };
}
