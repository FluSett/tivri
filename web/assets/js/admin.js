import { router } from './core/router.js';

import { initLoginForm } from './admin/loginForm.js';
import { initLeadsTable } from './admin/leadsTable.js';
import { initMessagesTable } from './admin/messagesTable.js';
import { initPortfolioForm } from './admin/portfolioForm.js';
import { initPortfolioCards } from './components/portfolio_card.js';

let savedSearchState = null;

document.addEventListener('htmx:configRequest', (event) => {
    const token = document.querySelector('meta[name="csrf-token"]')?.content;
    if (token) {
        event.detail.headers['X-CSRF-Token'] = token;
    }
});

document.body.addEventListener('htmx:beforeRequest', (e) => {
    const btn = e.detail.elt.querySelector('button[type="submit"]') || (e.detail.elt.tagName === 'BUTTON' ? e.detail.elt : null);
    if (btn) {
        btn.disabled = true;
        btn.classList.add('opacity-50', 'pointer-events-none');
    }
});

document.body.addEventListener('htmx:afterRequest', (e) => {
    const btn = e.detail.elt.querySelector('button[type="submit"]') || (e.detail.elt.tagName === 'BUTTON' ? e.detail.elt : null);
    if (btn) {
        btn.disabled = false;
        btn.classList.remove('opacity-50', 'pointer-events-none');
    }
});

document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        const modals = document.querySelectorAll('[role="dialog"]:not(.hidden), .modal:not(.hidden)');
        modals.forEach((m) => m.classList.add('hidden'));
    }
});

document.body.addEventListener('htmx:beforeSwap', (e) => {
    const activeEl = document.activeElement;
    if (activeEl && (activeEl.name === 'search_query' || (activeEl.id && activeEl.id.includes('search-input')))) {
        savedSearchState = {
            id: activeEl.id,
            name: activeEl.name,
            selectionStart: activeEl.selectionStart,
            selectionEnd: activeEl.selectionEnd
        };
    } else {
        savedSearchState = null;
    }
});

document.body.addEventListener('htmx:afterSwap', (e) => {
    if (savedSearchState) {
        let input = null;
        if (savedSearchState.id) {
            input = document.getElementById(savedSearchState.id);
        }
        if (!input && savedSearchState.name && e.detail.target) {
            input = e.detail.target.querySelector(`input[name="${savedSearchState.name}"]`);
        }
        if (input) {
            input.focus();
            if (typeof savedSearchState.selectionStart === 'number' && typeof savedSearchState.selectionEnd === 'number') {
                try {
                    input.setSelectionRange(savedSearchState.selectionStart, savedSearchState.selectionEnd);
                } catch (err) {
                    const len = input.value.length;
                    input.setSelectionRange(len, len);
                }
            }
        }
        savedSearchState = null;
    }
});

router.on('/admin/login', () => {
    return initLoginForm();
});

router.on('/admin/leads', () => {
    let cleanup = initLeadsTable();
    const handleSwap = (e) => {
        if (e.detail.target.id === 'leads-container') {
            if (cleanup) cleanup();
            cleanup = initLeadsTable();
        }
    };
    document.body.addEventListener('htmx:afterSwap', handleSwap);
    return () => {
        if (cleanup) cleanup();
        document.body.removeEventListener('htmx:afterSwap', handleSwap);
    };
});

router.on('/admin/messages', () => {
    let cleanup = initMessagesTable();
    const handleSwap = (e) => {
        if (e.detail.target.id === 'messages-container') {
            if (cleanup) cleanup();
            cleanup = initMessagesTable();
        }
    };
    document.body.addEventListener('htmx:afterSwap', handleSwap);
    return () => {
        if (cleanup) cleanup();
        document.body.removeEventListener('htmx:afterSwap', handleSwap);
    };
});

router.on(/^\/admin(\/portfolio)?$/, () => {
    const tdForm = initPortfolioForm();
    const tdCards = initPortfolioCards();
    return () => {
        if (tdForm) tdForm();
        if (tdCards) tdCards();
    };
});
