import { loginForm } from './admin/loginForm.js';
import { leadsTable } from './admin/leadsTable.js';
import { messagesTable } from './admin/messagesTable.js';
import { portfolioForm } from './admin/portfolioForm.js';

document.addEventListener('alpine:init', () => {
    Alpine.data('loginForm', loginForm);
    Alpine.data('leadsTable', leadsTable);
    Alpine.data('messagesTable', messagesTable);
    Alpine.data('portfolioForm', portfolioForm);
});
