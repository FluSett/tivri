if ('scrollRestoration' in history) {
    history.scrollRestoration = 'manual';
}

(function() {
    function updateScrollState() {
        var header = document.getElementById('site-header');
        if (header) {
            if (window.scrollY > 50) {
                header.classList.add('backdrop-blur-lg', 'bg-black/90', 'border-b', 'border-white/[0.08]', 'py-5', 'shadow-[0_4px_30px_rgba(0,0,0,0.8)]');
                header.classList.remove('bg-transparent', 'border-transparent', 'py-10');
            } else {
                header.classList.remove('backdrop-blur-lg', 'bg-black/90', 'border-b', 'border-white/[0.08]', 'py-5', 'shadow-[0_4px_30px_rgba(0,0,0,0.8)]');
                header.classList.add('bg-transparent', 'border-transparent', 'py-10');
            }
        }

        var footer = document.getElementById('site-footer');
        if (footer) {
            var isAtBottom = (window.innerHeight + window.pageYOffset >= document.documentElement.scrollHeight - 150);
            if (isAtBottom) {
                footer.classList.add('backdrop-blur-lg', 'bg-black/90', 'border-white/[0.08]', 'py-6', 'shadow-[0_-4px_30px_rgba(0,0,0,0.8)]');
                footer.classList.remove('bg-transparent', 'border-transparent', 'py-12');
            } else {
                footer.classList.remove('backdrop-blur-lg', 'bg-black/90', 'border-white/[0.08]', 'py-6', 'shadow-[0_-4px_30px_rgba(0,0,0,0.8)]');
                footer.classList.add('bg-transparent', 'border-transparent', 'py-12');
            }
        }
    }

    window.addEventListener('scroll', updateScrollState);

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', updateScrollState);
    } else {
        updateScrollState();
    }

    document.addEventListener('htmx:beforeSwap', function() {
        document.documentElement.classList.add('no-transition');
    });

    document.addEventListener('htmx:afterSwap', function() {
        updateScrollState();
        setTimeout(updateScrollState, 50);
        setTimeout(function() {
            document.documentElement.classList.remove('no-transition');
        }, 100);
    });
})();

(function() {
    function initNavObserver() {
        var sections = ['about', 'benefits', 'skills', 'portfolio', 'contact'];
        var navLinks = document.querySelectorAll('#main-nav .nav-link');

        function clearActive() {
            navLinks.forEach(function(link) {
                link.classList.remove('nav-active');
            });
        }

        function setActive(sectionId) {
            clearActive();
            navLinks.forEach(function(link) {
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

        var observer = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    setActive(entry.target.id);
                }
            });
        }, observerOptions);

        sections.forEach(function(id) {
            var el = document.getElementById(id);
            if (el) { observer.observe(el); }
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initNavObserver);
    } else {
        initNavObserver();
    }

    document.addEventListener('htmx:afterSwap', initNavObserver);
})();

