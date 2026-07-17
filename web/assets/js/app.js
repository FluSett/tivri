import './core/turnstile.js';
import './core/scroll.js';
import './core/navigation.js';
import './core/alpine.js';

// Consolidate all modules into a single bundle
import './admin.js';
import './components/stepper.js';
import './components/contact.js';

document.addEventListener('htmx:configRequest', (event) => {
    window.htmx.config.globalViewTransitions = true;
});
