export function initNavigation() {
    function initNavObserver() {
        var sections = ['about', 'benefits', 'skills', 'portfolio', 'contact'];
        var navLinks = document.querySelectorAll('#main-nav .nav-link');

        function clearActive() {
            navLinks.forEach(function (link) {
                link.classList.remove('nav-active');
            });
        }

        function setActive(sectionId) {
            clearActive();
            navLinks.forEach(function (link) {
                if (link.getAttribute('href') === '/#' + sectionId) {
                    link.classList.add('nav-active');
                }
            });
        }

        var observerOptions = {
            root: null,
            rootMargin: '-20% 0px -60% 0px',
            threshold: 0
        };

        var observer = new IntersectionObserver(function (entries) {
            entries.forEach(function (entry) {
                if (entry.isIntersecting) {
                    setActive(entry.target.id);
                }
            });
        }, observerOptions);

        sections.forEach(function (id) {
            var el = document.getElementById(id);
            if (el) {
                observer.observe(el);
            }
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initNavObserver);
    } else {
        initNavObserver();
    }

    document.addEventListener('htmx:afterSwap', initNavObserver);

    document.addEventListener('htmx:responseError', function (evt) {
        var errorText = evt.detail.xhr.responseText || 'An error occurred during submission.';
        window.dispatchEvent(new CustomEvent('tivri-error', { detail: errorText }));
    });
}

document.addEventListener(
    'click',
    function (e) {
        const link = e.target.closest('a');
        if (!link) return;

        const href = link.getAttribute('href');
        if (!href) return;

        const isRootHash = href.startsWith('/#') && window.location.pathname === '/';
        const isPureHash = href.startsWith('#');

        if (isRootHash || isPureHash) {
            const hashStr = isRootHash ? href.substring(1) : href;
            if (hashStr && hashStr !== '#') {
                const target = document.querySelector(hashStr);
                if (target) {
                    e.preventDefault();
                    e.stopPropagation(); // Hide the click from HTMX
                    target.scrollIntoView({ behavior: 'smooth' });
                    if (history.pushState) history.pushState(null, null, hashStr);

                    if (link.closest('[data-action="closeMenu"]') || link.getAttribute('data-action') === 'closeMenu') {
                        window.dispatchEvent(new CustomEvent('tivri-close-menu'));
                    }
                }
            }
            return;
        }

        if (
            link.origin === window.location.origin &&
            link.pathname === window.location.pathname &&
            link.search === window.location.search &&
            !link.hash
        ) {
            e.preventDefault();
            e.stopPropagation();
            window.scrollTo({ top: 0, behavior: 'smooth' });

            if (link.closest('[data-action="closeMenu"]') || link.getAttribute('data-action') === 'closeMenu') {
                window.dispatchEvent(new CustomEvent('tivri-close-menu'));
            }
        }
    },
    true
); // use capture phase to run before HTMX body listener

initNavigation();
