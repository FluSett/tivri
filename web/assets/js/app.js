import './core/turnstile.js';
import './core/navigation.js';

import { initScroll } from './core/scroll.js';
import { router } from './core/router.js';
import { initGlobalError } from './core/error.js';
import { initGlobalSuccess } from './core/success.js';
import { initLayout } from './core/layout.js';
import { initCookieConsent } from './components/cookieConsent.js';
import { initStatePersistence } from './core/state.js';
import { initTechBanner } from './components/techBanner.js';

import { initPortfolioCards } from './components/portfolio_card.js';
import { initStepper } from './components/stepper.js';
import { initContact } from './components/contact.js';

document.addEventListener('htmx:configRequest', (event) => {
    window.htmx.config.globalViewTransitions = true;
});

// Suppress expected AbortErrors from skipped HTMX view transitions
window.addEventListener('unhandledrejection', (event) => {
    if (event.reason && event.reason.name === 'AbortError') {
        event.preventDefault();
    }
});

// Initialize global form persistence logic for HTMX navigations
initStatePersistence();

let loaderTimeout;
document.addEventListener('htmx:beforeRequest', (e) => {
    const loader = document.getElementById('global-loader');

    // Clean up any zombie dropdowns appended to body before swapping content
    document.querySelectorAll('.tivri-dropdown-zombie').forEach((el) => el.remove());

    if (!loader) return;

    if (
        e.detail.requestConfig.target &&
        e.detail.requestConfig.target.id !== 'app-body' &&
        !e.target.hasAttribute('hx-push-url')
    ) {
        return;
    }

    clearTimeout(loaderTimeout);
    loader.style.width = '0%';
    loader.classList.remove('opacity-0');
    loader.classList.add('opacity-100');

    setTimeout(() => {
        loader.style.width = '30%';
    }, 50);
    setTimeout(() => {
        loader.style.width = '60%';
    }, 300);
    setTimeout(() => {
        loader.style.width = '85%';
    }, 800);
});

document.addEventListener('htmx:afterSettle', () => {
    const loader = document.getElementById('global-loader');
    if (!loader) return;

    loader.style.width = '100%';
    loaderTimeout = setTimeout(() => {
        loader.classList.remove('opacity-100');
        loader.classList.add('opacity-0');
        setTimeout(() => {
            loader.style.width = '0%';
        }, 300); // Reset width after fade out
    }, 400);
});

router.on(/.*/, () => {
    const tdError = initGlobalError();
    const tdSuccess = initGlobalSuccess();
    const tdLayout = initLayout();
    const tdScroll = initScroll();
    const tdCookie = initCookieConsent();
    const tdTech = initTechBanner();

    return () => {
        if (tdError) tdError();
        if (tdSuccess) tdSuccess();
        if (tdLayout) tdLayout();
        if (tdScroll) tdScroll();
        if (tdCookie) tdCookie();
        if (tdTech) tdTech();
    };
});

router.on('/', () => {
    const tdPortfolio = initPortfolioCards();
    const tdContact = initContact();
    const tdStepper = initStepper();

    return () => {
        if (tdPortfolio) tdPortfolio();
        if (tdContact) tdContact();
        if (tdStepper) tdStepper();
    };
});
