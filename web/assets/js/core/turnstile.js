window.tivriTurnstileSiteKey = document.body ? document.body.dataset.turnstileSitekey : '';
window.tivriTurnstileOnError = function (code) {
    console.warn('Turnstile challenge error:', code);
    return true;
};

export function initTurnstile(container, onVerify, onExpire, onError) {
    if (!window.tivriTurnstileSiteKey || !window.turnstile || !container) return null;
    try {
        return window.turnstile.render(container, {
            sitekey: window.tivriTurnstileSiteKey,
            theme: 'dark',
            size: 'normal',
            language: document.documentElement.lang || 'en',
            callback: onVerify,
            'expired-callback': onExpire,
            'error-callback':
                onError ||
                function () {
                    window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Security verification failed.' }));
                    return true;
                }
        });
    } catch (e) {
        console.error('Turnstile render failed:', e);
        return null;
    }
}

export function resetTurnstile(id) {
    if (window.turnstile && id !== null && id !== undefined) {
        window.turnstile.reset(id);
    }
}

export function destroyTurnstile(id) {
    if (window.turnstile && id !== null && id !== undefined) {
        try {
            window.turnstile.remove(id);
        } catch (e) {}
    }
}
