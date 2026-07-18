import { createReactiveState, bindRefs, delegate } from '../core/state.js';
import { initTurnstile, resetTurnstile, destroyTurnstile } from '../core/turnstile.js';
import { isValidName, isValidScope, isValidDeadline, isValidEmail, isValidBudget } from '../core/validators.js';
import { toggleVisibility, toggleClasses, setButtonSubmittingState } from '../core/dom.js';

// Removed STORAGE_MAP

export function initStepper() {
    const container = document.getElementById('stepper-container');
    if (!container) return;

    const highQueueActive = container.getAttribute('data-high-queue') === 'true';
    const refs = bindRefs(container);
    refs.budgetBtns = Array.from(container.querySelectorAll('.stepper-budget-btn'));

    let turnstileTimeout;

    const state = createReactiveState(
        'stepper',
        {
            openStepper: false,
            step: 1,
            budget: '',
            customBudget: '',
            scopeText: '',
            nameText: '',
            deadlineNeeded: false,
            deadlineSpec: '',
            contactEmail: '',
            contactInfo: '',
            submitted: false,
            nameTouched: false,
            scopeTouched: false,
            budgetTouched: false,
            emailTouched: false,
            deadlineTouched: false,
            submitStatus: 'idle',
            turnstileToken: '',
            turnstileId: null,
            isVerified: false
        },
        { ephemeralKeys: ['turnstileId', 'turnstileToken', 'isVerified'] },
        (newState) => {
            if (newState.step === 5 && !newState.submitted && newState.openStepper) {
                clearTimeout(turnstileTimeout);
                turnstileTimeout = setTimeout(renderTurnstile, 300);
            }

            updateUI();
        }
    );

    if (highQueueActive) {
        state.deadlineNeeded = false;
        state.deadlineSpec = '';
    }

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

    function canGoNext(currentStep) {
        if (currentStep === 1) return isValidName(state.nameText);
        if (currentStep === 2) return isValidScope(state.scopeText);
        if (currentStep === 3) return state.deadlineNeeded ? isValidDeadline(state.deadlineSpec) : true;
        if (currentStep === 4) return isValidBudget(state.budget, state.customBudget);
        return true;
    }

    function updateUI() {
        toggleClasses(
            refs.intro,
            !state.openStepper,
            ['grid-expand'],
            ['opacity-0', 'pointer-events-none', 'delay-300']
        );
        toggleClasses(
            refs.formContainer,
            state.openStepper,
            ['grid-expand'],
            ['opacity-0', 'pointer-events-none', 'delay-300']
        );

        toggleClasses(
            refs.submittedView,
            state.submitted,
            ['opacity-100', 'scale-100', 'pointer-events-auto'],
            ['opacity-0', 'scale-95', 'pointer-events-none', 'hidden']
        );
        toggleClasses(
            refs.notSubmittedView,
            !state.submitted,
            ['opacity-100', 'scale-100', 'pointer-events-auto'],
            ['opacity-0', 'scale-95', 'pointer-events-none', 'hidden']
        );

        if (refs.progressTextStep) refs.progressTextStep.textContent = state.step;
        if (refs.progressBar) refs.progressBar.style.width = (state.step / 5) * 100 + '%';

        [refs.step1, refs.step2, refs.step3, refs.step4, refs.step5].forEach((el, idx) => {
            toggleVisibility(el, state.step === idx + 1);
        });

        refs.nameCount.textContent = state.nameText.length + ' / 150';
        toggleVisibility(refs.nameError, state.nameTouched && !isValidName(state.nameText));
        refs.nameNext.disabled = !canGoNext(1);

        refs.scopeCount.textContent = state.scopeText.length + ' / 2000';
        toggleVisibility(refs.scopeError, state.scopeTouched && !isValidScope(state.scopeText));
        refs.scopeNext.disabled = !canGoNext(2);

        if (highQueueActive) {
            refs.deadlineStrict.disabled = true;
            refs.deadlineStandard.disabled = true;
        } else {
            toggleClasses(refs.deadlineStrict, state.deadlineNeeded, ['btn-choice-active'], []);
            toggleClasses(refs.deadlineStandard, !state.deadlineNeeded, ['btn-choice-active'], []);
            toggleVisibility(refs.deadlineInputContainer, state.deadlineNeeded);
        }
        refs.deadlineCount.textContent = state.deadlineSpec.length + ' / 300';
        toggleVisibility(
            refs.deadlineError,
            state.deadlineTouched && state.deadlineNeeded && !isValidDeadline(state.deadlineSpec)
        );
        refs.deadlineNext.disabled = !canGoNext(3);

        refs.budgetHidden.value = state.budget;
        refs.customBudgetHidden.value = state.budget === 'other' ? state.customBudget : '';
        refs.deadlineNeededHidden.value = state.deadlineNeeded ? 'true' : 'false';
        refs.deadlineSpecHidden.value = state.deadlineNeeded ? state.deadlineSpec : '';

        if (refs.budgetBtns) {
            refs.budgetBtns.forEach((btn) => {
                toggleClasses(btn, btn.getAttribute('data-val') === state.budget, ['btn-choice-active'], []);
            });
        }
        toggleVisibility(refs.customBudgetContainer, state.budget === 'other');

        const cBudgetValid = state.customBudget.trim() !== '' && parseInt(state.customBudget, 10) >= 100;
        toggleClasses(
            refs.customBudgetError,
            state.budgetTouched && !cBudgetValid && state.budget === 'other',
            ['opacity-100', 'flex'],
            ['opacity-0', 'hidden']
        );
        refs.budgetNext.disabled = !canGoNext(4);

        const eValid = isValidEmail(state.contactEmail);
        toggleClasses(
            refs.contactEmailError,
            state.emailTouched && !eValid,
            ['opacity-100', 'flex'],
            ['opacity-0', 'hidden']
        );

        if (refs.contactInfoCount) {
            refs.contactInfoCount.textContent = state.contactInfo.length + ' / 100';
        }

        setButtonSubmittingState(
            refs.submitBtn,
            refs.submitText,
            refs.submitSpinner,
            state.submitStatus !== 'idle',
            !eValid || (window.tivriTurnstileSiteKey && !state.isVerified)
        );
    }

    refs.name.value = state.nameText;
    refs.scope.value = state.scopeText;
    refs.deadlineSpec.value = state.deadlineSpec;
    refs.customBudget.value = state.customBudget;
    refs.contactEmail.value = state.contactEmail;
    refs.contactInfo.value = state.contactInfo;

    const teardowns = [];

    teardowns.push(
        delegate(container, 'click', '[data-action]', (e, target) => {
            switch (target.dataset.action) {
                case 'open':
                    state.openStepper = true;
                    break;
                case 'close':
                    state.step = 1;
                    state.budget = '';
                    state.customBudget = '';
                    state.scopeText = '';
                    state.nameText = '';
                    state.deadlineNeeded = false;
                    state.deadlineSpec = '';
                    state.contactEmail = '';
                    state.contactInfo = '';
                    state.submitted = false;
                    state.nameTouched = false;
                    state.scopeTouched = false;
                    state.budgetTouched = false;
                    state.emailTouched = false;
                    state.deadlineTouched = false;
                    state.submitStatus = 'idle';
                    state.openStepper = false;

                    refs.name.value = '';
                    refs.scope.value = '';
                    refs.deadlineSpec.value = '';
                    refs.customBudget.value = '';
                    refs.contactEmail.value = '';
                    refs.contactInfo.value = '';
                    refs.form.reset();
                    doResetTurnstile();
                    break;
                case 'next':
                    if (state.step < 5) state.step++;
                    break;
                case 'prev':
                    if (state.step > 1) state.step--;
                    break;
            }
        })
    );

    if (refs.budgetBtns) {
        refs.budgetBtns.forEach((btn) => {
            const h = () => {
                state.budget = btn.getAttribute('data-val');
                if (state.budget !== 'other') state.customBudget = '';
            };
            btn.addEventListener('click', h);
            teardowns.push(() => btn.removeEventListener('click', h));
        });
    }

    const hStrict = () => {
        if (!highQueueActive) state.deadlineNeeded = true;
    };
    refs.deadlineStrict.addEventListener('click', hStrict);
    teardowns.push(() => refs.deadlineStrict.removeEventListener('click', hStrict));

    const hStandard = () => {
        if (!highQueueActive) {
            state.deadlineNeeded = false;
            state.deadlineSpec = '';
            refs.deadlineSpec.value = '';
        }
    };
    refs.deadlineStandard.addEventListener('click', hStandard);
    teardowns.push(() => refs.deadlineStandard.removeEventListener('click', hStandard));

    function debounce(func, wait) {
        let timeout;
        return function (...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }

    const onInputDebounced = debounce((field) => {
        state[`${field}Touched`] = true;
    }, 500);

    teardowns.push(
        delegate(container, 'input', 'input, textarea', (e, target) => {
            const ref = target.getAttribute('data-ref') || target.id;
            if (!ref) return;

            if (ref === 'name') {
                state.nameText = target.value;
                onInputDebounced('name');
            } else if (ref === 'scope') {
                state.scopeText = target.value;
                onInputDebounced('scope');
            } else if (ref === 'deadlineSpec') {
                state.deadlineSpec = target.value;
                onInputDebounced('deadline');
            } else if (ref === 'customBudget') {
                target.value = target.value.replace(/[^0-9]/g, '');
                state.customBudget = target.value;
                onInputDebounced('budget');
            } else if (ref === 'contactEmail') {
                state.contactEmail = target.value;
                onInputDebounced('email');
            } else if (ref === 'contactInfo') {
                state.contactInfo = target.value;
            }
        })
    );

    teardowns.push(
        delegate(container, 'focusout', 'input, textarea', (e, target) => {
            const ref = target.getAttribute('data-ref') || target.id;
            if (!ref) return;

            if (ref === 'name') state.nameTouched = true;
            else if (ref === 'scope') state.scopeTouched = true;
            else if (ref === 'deadlineSpec') state.deadlineTouched = true;
            else if (ref === 'customBudget') state.budgetTouched = true;
            else if (ref === 'contactEmail') state.emailTouched = true;
        })
    );

    const hFormSubmit = (e) => {
        if (window.tivriTurnstileSiteKey && window.turnstile && !state.isVerified) {
            e.preventDefault();
            e.stopPropagation();
            window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Please complete the security check.' }));
            return;
        }
        state.submitStatus = 'submitting';
    };
    refs.form.addEventListener('submit', hFormSubmit);
    teardowns.push(() => refs.form.removeEventListener('submit', hFormSubmit));

    const hFormAfterReq = (e) => {
        if (e.detail.successful) {
            state.submitted = true;
            state.openStepper = true;
            container.dispatchEvent(new CustomEvent('tivri:stepper:completed', { bubbles: true }));
        }
        state.submitStatus = 'idle';
    };
    refs.form.addEventListener('htmx:afterRequest', hFormAfterReq);
    teardowns.push(() => refs.form.removeEventListener('htmx:afterRequest', hFormAfterReq));

    const hResErr = () => {
        state.submitStatus = 'idle';
        doResetTurnstile();
    };
    document.addEventListener('htmx:responseError', hResErr);
    teardowns.push(() => document.removeEventListener('htmx:responseError', hResErr));

    updateUI();

    return () => {
        clearTimeout(turnstileTimeout);
        destroyTurnstile(state.turnstileId);
        teardowns.forEach((fn) => fn());
    };
}
