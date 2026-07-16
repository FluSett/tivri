// Global utility to handle Alpine state reset only on hard reloads (not locale swaps)
export function tivriHandleLocaleChange(onNormalLoad) {
    const isLocaleChange = sessionStorage.getItem('locale_change') === 'true';
    if (!isLocaleChange) {
        if (typeof onNormalLoad === 'function') onNormalLoad();
    } else {
        setTimeout(() => sessionStorage.removeItem('locale_change'), 500);
    }
}
window.tivriHandleLocaleChange = tivriHandleLocaleChange;

// Scroll preservation and height freezing on htmx body swaps
(function () {
    sessionStorage.clear();
    if ('scrollRestoration' in history) {
        history.scrollRestoration = 'manual';
    }
    window.scrollTo({ top: 0, behavior: 'instant' });

    document.addEventListener('click', function (e) {
        if (e.target.closest('[data-preserve-scroll]')) {
            sessionStorage.setItem('tivri_scroll', window.scrollY);
        }
    });

    document.addEventListener('htmx:beforeSwap', function (e) {
        if (sessionStorage.getItem('tivri_scroll') !== null) {
            document.documentElement.style.minHeight = document.documentElement.scrollHeight + 'px';
        }
    });

    document.addEventListener('htmx:afterSettle', function (e) {
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
    });
})();

(function () {
    var footerActive = false;

    function updateScrollState() {
        var header = document.getElementById('site-header');
        if (header) {
            if (window.scrollY > 50) {
                header.classList.add(
                    'backdrop-blur-lg',
                    'bg-black/90',
                    'border-b',
                    'border-white/[0.08]',
                    'py-5',
                    'shadow-[0_4px_30px_rgba(0,0,0,0.8)]'
                );
                header.classList.remove('bg-transparent', 'border-transparent', 'py-10');
            } else {
                header.classList.remove(
                    'backdrop-blur-lg',
                    'bg-black/90',
                    'border-b',
                    'border-white/[0.08]',
                    'py-5',
                    'shadow-[0_4px_30px_rgba(0,0,0,0.8)]'
                );
                header.classList.add('bg-transparent', 'border-transparent', 'py-10');
            }
        }

        var footer = document.getElementById('site-footer');
        if (footer) {
            var scrollY = window.pageYOffset || window.scrollY;
            var maxScroll = document.documentElement.scrollHeight - window.innerHeight;
            var threshold = footerActive ? 155 : 135;
            var isAtBottom = maxScroll - scrollY <= threshold;

            if (isAtBottom && !footerActive) {
                footerActive = true;
                footer.classList.add(
                    'backdrop-blur-lg',
                    'bg-black/90',
                    'border-white/[0.08]',
                    'shadow-[0_-4px_30px_rgba(0,0,0,0.8)]'
                );
                footer.classList.remove('bg-transparent', 'border-transparent');
            } else if (!isAtBottom && footerActive) {
                footerActive = false;
                footer.classList.remove(
                    'backdrop-blur-lg',
                    'bg-black/90',
                    'border-white/[0.08]',
                    'shadow-[0_-4px_30px_rgba(0,0,0,0.8)]'
                );
                footer.classList.add('bg-transparent', 'border-transparent');
            }
        }
    }

    window.addEventListener('scroll', updateScrollState);

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', updateScrollState);
    } else {
        updateScrollState();
    }

    document.addEventListener('htmx:beforeSwap', function (evt) {
        document.documentElement.classList.add('no-transition');
        sessionStorage.setItem('tivri_htmx_nav', 'true');

        if (evt.detail.serverResponse) {
            try {
                var parser = new DOMParser();
                var doc = parser.parseFromString(evt.detail.serverResponse, 'text/html');
                var oldHeader = document.getElementById('site-header');
                var newHeader = doc.getElementById('site-header');
                if (oldHeader && newHeader) {
                    newHeader.className = oldHeader.className;
                }

                var oldFooter = document.getElementById('site-footer');
                var newFooter = doc.getElementById('site-footer');
                if (oldFooter && newFooter) {
                    newFooter.className = oldFooter.className;
                }

                evt.detail.serverResponse = doc.documentElement.outerHTML;
            } catch (e) {
                console.error('Failed to parse server response:', e);
            }
        }
    });

    document.addEventListener('htmx:afterSwap', function () {
        updateScrollState();
        setTimeout(updateScrollState, 50);
        setTimeout(function () {
            document.documentElement.classList.remove('no-transition');
        }, 100);
    });
})();
