import './core/turnstile.js';
import './core/scroll.js';
import './core/navigation.js';
import './core/alpine.js';

document.addEventListener('htmx:configRequest', (event) => {
    window.htmx.config.globalViewTransitions = true;
});
