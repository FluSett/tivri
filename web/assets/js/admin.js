import { router } from './core/router.js';

import { initLoginForm } from './admin/loginForm.js';
import { initLeadsTable } from './admin/leadsTable.js';
import { initMessagesTable } from './admin/messagesTable.js';
import { initPortfolioForm } from './admin/portfolioForm.js';
import { initPortfolioCards } from './components/portfolio_card.js';

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
