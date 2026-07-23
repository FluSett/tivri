import { bindRefs } from './state.js';
import { setSessionItem, getSessionItem, setSessionJSON, getSessionJSON, removeSessionKey } from './storage.js';

export function tivriHandleLocaleChange(onNormalLoad) {
    const isLocaleChange = getSessionItem('locale_change') === 'true';
    if (!isLocaleChange) {
        if (typeof onNormalLoad === 'function') onNormalLoad();
    } else {
        setTimeout(() => removeSessionKey('locale_change'), 500);
    }
}
window.tivriHandleLocaleChange = tivriHandleLocaleChange;

function savePreservedScroll(isLocaleChange = false) {
    const sections = Array.from(document.querySelectorAll('section[id], main[id], .admin-panel-box[id], #stepper-container, #messages-container, #leads-container')).filter((el) => {
        const rect = el.getBoundingClientRect();
        return rect.bottom > 0 && rect.top < window.innerHeight;
    });

    let bestEl = null;
    let bestDist = Infinity;
    sections.forEach((el) => {
        const rect = el.getBoundingClientRect();
        const dist = Math.abs(rect.top - 80);
        if (dist < bestDist) {
            bestDist = dist;
            bestEl = el;
        }
    });

    const data = {
        id: bestEl ? bestEl.id : null,
        offsetTop: bestEl ? bestEl.getBoundingClientRect().top : 0,
        scrollY: window.scrollY
    };

    setSessionJSON('tivri_preserved_scroll', data);
    if (isLocaleChange) {
        setSessionItem('locale_change', 'true');
    }
}

function restorePreservedScroll() {
    const data = getSessionJSON('tivri_preserved_scroll');
    if (!data) return false;
    removeSessionKey('tivri_preserved_scroll');

    try {
        const html = document.documentElement;
        html.style.scrollBehavior = 'auto';
        html.classList.remove('scroll-smooth');

        const doScroll = () => {
            let restored = false;
            if (data.id) {
                const el = document.getElementById(data.id);
                if (el) {
                    const rect = el.getBoundingClientRect();
                    const targetY = window.scrollY + rect.top - data.offsetTop;
                    window.scrollTo(0, Math.max(0, targetY));
                    restored = true;
                }
            }
            if (!restored && typeof data.scrollY === 'number') {
                window.scrollTo(0, data.scrollY);
            }
        };

        doScroll();
        requestAnimationFrame(() => {
            doScroll();
            setTimeout(() => {
                doScroll();
                html.style.scrollBehavior = '';
                html.classList.add('scroll-smooth');
            }, 100);
        });
        return true;
    } catch (e) {
        console.error('Failed to restore scroll:', e);
        return false;
    }
}

if ('scrollRestoration' in history) {
    history.scrollRestoration = 'manual';
}

export function initScroll() {
    const teardowns = [];

    const isReload = (performance.getEntriesByType && performance.getEntriesByType('navigation')[0] && performance.getEntriesByType('navigation')[0].type === 'reload') || performance.navigation?.type === 1;
    const localeChange = getSessionItem('locale_change');
    const preservedScroll = getSessionItem('tivri_preserved_scroll');
    const isRealRefresh = isReload && localeChange !== 'true';

    const hasHash = Boolean(window.location.hash && window.location.hash !== '#');

    if (!hasHash) {
        if (isRealRefresh || (!preservedScroll && localeChange !== 'true')) {
            removeSessionKey('tivri_preserved_scroll');
            removeSessionKey('tivri_htmx_nav');
            removeSessionKey('locale_change');

            const html = document.documentElement;
            html.style.scrollBehavior = 'auto';
            html.classList.remove('scroll-smooth');

            window.scrollTo(0, 0);
            requestAnimationFrame(() => {
                window.scrollTo(0, 0);
                setTimeout(() => {
                    window.scrollTo(0, 0);
                    html.style.scrollBehavior = '';
                    html.classList.add('scroll-smooth');
                }, 50);
            });
        } else {
            restorePreservedScroll();
        }
    }

    const clickPreserveHandler = function (e) {
        const preserveEl = e.target.closest('[data-preserve-scroll]') || e.target.closest('.lang-switch-btn') || e.target.closest('.lang-switch-desktop');
        if (preserveEl) {
            savePreservedScroll(true);
        }
    };
    document.addEventListener('click', clickPreserveHandler);
    teardowns.push(() => document.removeEventListener('click', clickPreserveHandler));

    const beforeRequestHandler = function (e) {
        const target = e.target;
        if (target && (target.closest('[data-preserve-scroll]') || target.getAttribute('hx-swap') === 'none' || target.getAttribute('hx-target') === 'this')) {
            savePreservedScroll(false);
        }
    };
    document.addEventListener('htmx:beforeRequest', beforeRequestHandler);
    teardowns.push(() => document.removeEventListener('htmx:beforeRequest', beforeRequestHandler));

    const beforeSwapHandler = function (e) {
        document.documentElement.classList.add('no-transition');
        setSessionItem('tivri_htmx_nav', 'true');

        if (getSessionItem('tivri_preserved_scroll') !== null) {
            document.documentElement.style.minHeight = document.documentElement.scrollHeight + 'px';
        }

        if (e.detail.serverResponse) {
            try {
                const parser = new DOMParser();
                const doc = parser.parseFromString(e.detail.serverResponse, 'text/html');
                const oldHeader = document.querySelector('[data-ref="header"]');
                const newHeader = doc.querySelector('[data-ref="header"]');
                if (oldHeader && newHeader) {
                    newHeader.className = oldHeader.className;
                }

                const oldFooter = document.querySelector('[data-ref="footer"]');
                const newFooter = doc.querySelector('[data-ref="footer"]');
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

    const handleHtmxRestore = function () {
        restorePreservedScroll();
    };
    document.addEventListener('htmx:afterSettle', handleHtmxRestore);
    teardowns.push(() => document.removeEventListener('htmx:afterSettle', handleHtmxRestore));

    document.addEventListener('htmx:afterRequest', handleHtmxRestore);
    teardowns.push(() => document.removeEventListener('htmx:afterRequest', handleHtmxRestore));

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
            const scrollY = window.pageYOffset || window.scrollY;
            const maxScroll = document.documentElement.scrollHeight - window.innerHeight;
            const threshold = footerActive ? 155 : 135;
            const isAtBottom = maxScroll - scrollY <= threshold;

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
