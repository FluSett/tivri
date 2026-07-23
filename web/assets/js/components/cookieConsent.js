import { delegate } from '../core/state.js';
import { getLocalItem, setLocalItem } from '../core/storage.js';

export function initCookieConsent() {
    const banner = document.getElementById('cookie-banner');
    if (!banner) return;

    const accepted = getLocalItem('tivri_cookies_accepted') === 'true';

    const token = banner.getAttribute('data-cf-token');
    const nonce = banner.getAttribute('data-nonce');

    function loadAnalytics() {
        if (window.analyticsLoaded) return;
        window.analyticsLoaded = true;
        if (token && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
            const script = document.createElement('script');
            script.src = 'https://static.cloudflareinsights.com/beacon.min.js';
            script.defer = true;
            script.setAttribute('data-cf-beacon', '{"token": "' + token + '"}');
            if (nonce) {
                script.setAttribute('nonce', nonce);
            }
            document.body.appendChild(script);
        }
    }

    if (accepted) {
        loadAnalytics();
    } else {
        banner.classList.remove('hidden');
    }

    const cleanup = delegate(banner, 'click', '[data-action="acceptCookies"]', () => {
        setLocalItem('tivri_cookies_accepted', 'true');
        banner.classList.add('hidden');
        loadAnalytics();
    });

    return () => {
        cleanup();
    };
}
