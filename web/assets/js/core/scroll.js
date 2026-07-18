import { bindRefs } from './state.js';

export function tivriHandleLocaleChange(onNormalLoad) {
    const isLocaleChange = sessionStorage.getItem('locale_change') === 'true';
    if (!isLocaleChange) {
        if (typeof onNormalLoad === 'function') onNormalLoad();
    } else {
        setTimeout(() => sessionStorage.removeItem('locale_change'), 500);
    }
}
window.tivriHandleLocaleChange = tivriHandleLocaleChange;

export function initScroll() {
    const teardowns = [];

    const localeChange = sessionStorage.getItem('locale_change');
    const tivriScroll = sessionStorage.getItem('tivri_scroll');
    const htmxNav = sessionStorage.getItem('tivri_htmx_nav');

    sessionStorage.clear();

    if (localeChange) sessionStorage.setItem('locale_change', localeChange);
    if (tivriScroll) sessionStorage.setItem('tivri_scroll', tivriScroll);
    if (htmxNav) sessionStorage.setItem('tivri_htmx_nav', htmxNav);

    if (!localeChange && !tivriScroll) {
        if ('scrollRestoration' in history) {
            history.scrollRestoration = 'manual';
        }
        window.scrollTo({ top: 0, behavior: 'instant' });
    }

    const clickPreserveHandler = function (e) {
        if (e.target.closest('[data-preserve-scroll]')) {
            sessionStorage.setItem('tivri_scroll', window.scrollY);
        }
    };
    document.addEventListener('click', clickPreserveHandler);
    teardowns.push(() => document.removeEventListener('click', clickPreserveHandler));

    const beforeSwapHandler = function (e) {
        document.documentElement.classList.add('no-transition');
        sessionStorage.setItem('tivri_htmx_nav', 'true');

        if (sessionStorage.getItem('tivri_scroll') !== null) {
            document.documentElement.style.minHeight = document.documentElement.scrollHeight + 'px';
        }

        if (e.detail.serverResponse) {
            try {
                var parser = new DOMParser();
                var doc = parser.parseFromString(e.detail.serverResponse, 'text/html');
                var oldHeader = document.querySelector('[data-ref="header"]');
                var newHeader = doc.querySelector('[data-ref="header"]');
                if (oldHeader && newHeader) {
                    newHeader.className = oldHeader.className;
                }

                var oldFooter = document.querySelector('[data-ref="footer"]');
                var newFooter = doc.querySelector('[data-ref="footer"]');
                if (oldFooter && newFooter) {
                    newFooter.className = oldFooter.className;
                }

                e.detail.serverResponse = doc.documentElement.outerHTML;
            } catch (err) {
                console.error('Failed to parse server response:', err);
            }
        }
    };
    document.addEventListener('htmx:beforeSwap', beforeSwapHandler);
    teardowns.push(() => document.removeEventListener('htmx:beforeSwap', beforeSwapHandler));

    const afterSettleHandler = function (e) {
        const s = sessionStorage.getItem('tivri_scroll');
        if (s !== null) {
            setTimeout(function () {
                const html = document.documentElement;
                const hadSmooth = html.classList.contains('scroll-smooth');
                if (hadSmooth) html.classList.remove('scroll-smooth');

                window.scrollTo({ top: parseInt(s), behavior: 'instant' });
                sessionStorage.removeItem('tivri_scroll');

                if (hadSmooth) {
                    setTimeout(function () {
                        html.classList.add('scroll-smooth');
                    }, 50);
                }
                html.style.minHeight = '';
            }, 50);
        }
    };
    document.addEventListener('htmx:afterSettle', afterSettleHandler);
    teardowns.push(() => document.removeEventListener('htmx:afterSettle', afterSettleHandler));

    let footerActive = false;

    function updateScrollState() {
        const refs = bindRefs(document.body);
        const header = refs.header;

        if (header) {
            if (window.scrollY > 50) {
                header.classList.add('header-scrolled');
                header.classList.remove('bg-transparent', 'border-transparent', 'py-10');
            } else {
                header.classList.remove('header-scrolled');
                header.classList.add('bg-transparent', 'border-transparent', 'py-10');
            }
        }

        const footer = refs.footer;
        if (footer) {
            var scrollY = window.pageYOffset || window.scrollY;
            var maxScroll = document.documentElement.scrollHeight - window.innerHeight;
            var threshold = footerActive ? 155 : 135;
            var isAtBottom = maxScroll - scrollY <= threshold;

            if (isAtBottom && !footerActive) {
                footerActive = true;
                footer.classList.add('footer-scrolled');
                footer.classList.remove('bg-transparent', 'border-transparent');
            } else if (!isAtBottom && footerActive) {
                footerActive = false;
                footer.classList.remove('footer-scrolled');
                footer.classList.add('bg-transparent', 'border-transparent');
            }
        }
    }

    window.addEventListener('scroll', updateScrollState);
    teardowns.push(() => window.removeEventListener('scroll', updateScrollState));

    const afterSwapHandler = function () {
        updateScrollState();
        setTimeout(updateScrollState, 50);
        setTimeout(function () {
            document.documentElement.classList.remove('no-transition');
        }, 100);
    };
    document.addEventListener('htmx:afterSwap', afterSwapHandler);
    teardowns.push(() => document.removeEventListener('htmx:afterSwap', afterSwapHandler));

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', updateScrollState);
    } else {
        updateScrollState();
    }

    return () => teardowns.forEach((fn) => fn());
}
