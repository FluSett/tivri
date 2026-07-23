export function initNavigation() {
    const sections = ['about', 'benefits', 'skills', 'portfolio', 'contact'];
    let isManualScrolling = false;

    function getNavLinks() {
        return document.querySelectorAll('#main-nav .nav-link');
    }

    function clearActive() {
        getNavLinks().forEach((link) => link.classList.remove('nav-active'));
    }

    function setActive(sectionId) {
        clearActive();
        if (!sectionId) return;
        getNavLinks().forEach((link) => {
            const href = link.getAttribute('href');
            if (href === '/#' + sectionId || href === '#' + sectionId) {
                link.classList.add('nav-active');
            }
        });
    }

    function scrollToSection(targetEl, hashStr) {
        if (!targetEl) return;
        isManualScrolling = true;
        targetEl.scrollIntoView({ behavior: 'smooth' });

        const cleanHash = hashStr.startsWith('/#') ? hashStr : '/#' + hashStr.replace(/^#\/?/, '');
        if (history.pushState) {
            history.pushState(null, null, cleanHash);
        }

        setActive(cleanHash.replace('/#', ''));

        setTimeout(() => {
            isManualScrolling = false;
        }, 1000);
    }

    function initNavObserver() {
        const observerOptions = {
            root: null,
            rootMargin: '-20% 0px -50% 0px',
            threshold: 0
        };

        const observer = new IntersectionObserver((entries) => {
            if (isManualScrolling) return;

            const aboutEl = document.getElementById('about');
            const aboutTop = aboutEl ? aboutEl.getBoundingClientRect().top + window.scrollY - 150 : 300;

            if (window.scrollY < aboutTop) {
                clearActive();
                if (history.replaceState && window.location.hash) {
                    history.replaceState(null, null, window.location.pathname + window.location.search);
                }
                return;
            }

            entries.forEach((entry) => {
                if (entry.isIntersecting) {
                    const id = entry.target.id;
                    setActive(id);
                    const targetHash = '/#' + id;
                    if (history.replaceState && window.location.hash !== targetHash && window.location.hash !== '#' + id) {
                        history.replaceState(null, null, targetHash);
                    }
                }
            });
        }, observerOptions);

        sections.forEach((id) => {
            const el = document.getElementById(id);
            if (el) observer.observe(el);
        });
    }

    function handleInitialHashScroll() {
        if (window.location.hash && window.location.hash !== '#') {
            isManualScrolling = true;
            const rawHash = window.location.hash.replace('/#', '#');
            const target = document.querySelector(rawHash);
            if (target) {
                setTimeout(() => {
                    target.scrollIntoView({ behavior: 'smooth' });
                    setActive(rawHash.replace('#', ''));
                    setTimeout(() => {
                        isManualScrolling = false;
                    }, 1000);
                }, 100);
            } else {
                isManualScrolling = false;
            }
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            initNavObserver();
            handleInitialHashScroll();
        });
    } else {
        initNavObserver();
        handleInitialHashScroll();
    }

    document.addEventListener('htmx:afterSwap', initNavObserver);

    document.addEventListener('htmx:responseError', (evt) => {
        const errorText = evt.detail.xhr?.responseText || 'An error occurred during submission.';
        window.dispatchEvent(new CustomEvent('tivri-error', { detail: errorText }));
    });

    document.addEventListener(
        'click',
        (e) => {
            const link = e.target.closest('a');
            if (!link) return;

            const href = link.getAttribute('href');
            if (!href) return;

            const isRootHash = href.startsWith('/#') && window.location.pathname === '/';
            const isPureHash = href.startsWith('#');

            if (isRootHash || isPureHash) {
                const rawHash = isRootHash ? href.substring(1) : href;
                if (rawHash && rawHash !== '#') {
                    const target = document.querySelector(rawHash);
                    if (target) {
                        e.preventDefault();
                        e.stopPropagation();
                        scrollToSection(target, href);

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

                isManualScrolling = true;
                window.scrollTo({ top: 0, behavior: 'smooth' });

                if (history.pushState) {
                    history.pushState(null, null, window.location.pathname + window.location.search);
                }

                clearActive();

                if (link.closest('[data-action="closeMenu"]') || link.getAttribute('data-action') === 'closeMenu') {
                    window.dispatchEvent(new CustomEvent('tivri-close-menu'));
                }

                setTimeout(() => {
                    isManualScrolling = false;
                }, 1000);
            }
        },
        true
    );
}

initNavigation();
