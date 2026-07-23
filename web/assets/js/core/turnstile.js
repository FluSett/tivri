window.tivriTurnstileSiteKey = document.body ? document.body.dataset.turnstileSitekey : '';

const SUCCESS_DISPLAY_DELAY_MS = 450;
const COLLAPSE_ANIM_DURATION_MS = 400;

const widgetContainers = new Map();

function getTargetElement(container) {
    if (!container) return null;
    return container.closest('[data-ref="turnstileWrapper"]') || container.parentElement || container;
}

function prepareTarget(container) {
    const target = getTargetElement(container);
    if (!target) return null;
    target.style.overflow = 'hidden';
    target.style.transition =
        'max-height 0.4s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.35s ease, transform 0.35s ease, padding 0.4s ease, margin 0.4s ease';
    return target;
}

function applyCollapsedStyles(target) {
    target.style.opacity = '0';
    target.style.transform = 'translateY(-8px) scale(0.95)';
    target.style.maxHeight = '0px';
    target.style.paddingTop = '0px';
    target.style.paddingBottom = '0px';
    target.style.marginTop = '0px';
    target.style.marginBottom = '0px';
}

export function animateShow(container) {
    const target = prepareTarget(container);
    if (!target) return;

    target.style.display = 'flex';
    void target.offsetHeight;

    target.style.maxHeight = '140px';
    target.style.opacity = '1';
    target.style.transform = 'translateY(0) scale(1)';
    target.style.paddingTop = '0.5rem';
    target.style.paddingBottom = '0.5rem';
}

export function animateHide(container) {
    const target = prepareTarget(container);
    if (!target) return;

    applyCollapsedStyles(target);

    setTimeout(() => {
        if (target.style.opacity === '0' || target.style.maxHeight === '0px') {
            target.style.display = 'none';
        }
    }, COLLAPSE_ANIM_DURATION_MS);
}

export function initTurnstile(container, onVerify, onExpire, onError) {
    if (!window.tivriTurnstileSiteKey || !window.turnstile || !container) return null;

    const target = prepareTarget(container);
    if (target && !target.dataset.turnstileInitialized) {
        target.style.display = 'flex';
        applyCollapsedStyles(target);
        target.dataset.turnstileInitialized = 'true';
    }

    if (window.MutationObserver && !container.dataset.observerActive) {
        const observer = new MutationObserver(() => {
            if (container.childElementCount > 0) {
                const iframe = container.querySelector('iframe');
                if (iframe && iframe.offsetHeight > 0) {
                    animateShow(container);
                }
            }
        });
        observer.observe(container, { childList: true, subtree: true });
        container.dataset.observerActive = 'true';
    }

    try {
        const widgetId = window.turnstile.render(container, {
            sitekey: window.tivriTurnstileSiteKey,
            theme: 'dark',
            size: 'normal',
            appearance: 'interaction-only',
            language: document.documentElement.lang || 'en',
            callback: function (token) {
                setTimeout(() => {
                    animateHide(container);
                }, SUCCESS_DISPLAY_DELAY_MS);
                if (onVerify) onVerify(token);
            },
            'expired-callback': function () {
                animateShow(container);
                if (onExpire) onExpire();
            },
            'error-callback': function (err) {
                animateShow(container);
                if (onError) {
                    onError(err);
                } else {
                    window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Security verification failed.' }));
                    return true;
                }
            },
            'before-interactive-callback': function () {
                animateShow(container);
            }
        });

        if (widgetId !== null && widgetId !== undefined) {
            widgetContainers.set(widgetId, container);
        }
        return widgetId;
    } catch (e) {
        console.error('Turnstile render failed:', e);
        return null;
    }
}

export function resetTurnstile(id, container) {
    const targetContainer = container || widgetContainers.get(id);
    if (targetContainer) animateShow(targetContainer);
    if (window.turnstile && id !== null && id !== undefined) {
        window.turnstile.reset(id);
    }
}

export function destroyTurnstile(id) {
    const targetContainer = widgetContainers.get(id);
    if (targetContainer) animateHide(targetContainer);
    widgetContainers.delete(id);
    if (window.turnstile && id !== null && id !== undefined) {
        try {
            window.turnstile.remove(id);
        } catch (e) {}
    }
}
