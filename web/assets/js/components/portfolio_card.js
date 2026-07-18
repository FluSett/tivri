import { delegate } from '../core/state.js';

let eventsBound = false;

export function initPortfolioCards() {
    const tagsContainers = document.querySelectorAll('.portfolio-card-js .tags-container:not([data-processed="true"])');
    tagsContainers.forEach((container) => {
        const techStack = container.getAttribute('data-tags');
        if (techStack) {
            const tags = techStack
                .split(',')
                .map((t) => t.trim())
                .filter((t) => t);
            container.innerHTML = tags
                .map(
                    (tag) =>
                        `<span class="bg-white/4 text-neutral-300 text-xs px-2.5 py-1 rounded-lg border border-white/6 tracking-tight font-medium">${tag}</span>`
                )
                .join('');
        }
        container.setAttribute('data-processed', 'true');
    });

    // Initialize initial transforms
    document.querySelectorAll('.vanilla-carousel:not([data-carousel-processed="true"])').forEach((carousel) => {
        carousel.setAttribute('data-carousel-processed', 'true');
        const container = carousel.querySelector('.carousel-items-container');
        if (!container) return;
        const items = carousel.querySelectorAll('.carousel-item');
        const activeIdx = parseInt(container.getAttribute('data-active-index') || '0', 10);
        items.forEach((item, idx) => {
            const offset = (idx - activeIdx) * 100;
            item.style.transform = `translateX(${offset}%)`;
        });
    });

    const teardowns = [];

    function updateCarousel(carousel, newIndex) {
        const items = carousel.querySelectorAll('.carousel-item');
        const dots = carousel.querySelectorAll('.carousel-dot');

        const container = carousel.querySelector('.carousel-items-container');
        if (container) {
            container.setAttribute('data-active-index', newIndex);
        }

        items.forEach((item, idx) => {
            const offset = (idx - newIndex) * 100;
            item.style.transform = `translateX(${offset}%)`;
        });

        dots.forEach((dot, idx) => {
            if (idx === newIndex) {
                dot.classList.replace('bg-white/40', 'bg-primary');
                dot.classList.replace('w-1.5', 'w-4');
            } else {
                dot.classList.replace('bg-primary', 'bg-white/40');
                dot.classList.replace('w-4', 'w-1.5');
            }
        });
    }

    if (!eventsBound) {
        eventsBound = true;

        delegate(document.body, 'click', '.carousel-prev', (e, target) => {
            const carousel = target.closest('.vanilla-carousel');
            if (!carousel) return;
            const items = Array.from(carousel.querySelectorAll('.carousel-item'));
            const container = carousel.querySelector('.carousel-items-container');
            if (!container) return;
            const activeIdx = parseInt(container.getAttribute('data-active-index') || '0', 10);
            const newIndex = (activeIdx - 1 + items.length) % items.length;
            updateCarousel(carousel, newIndex);
        });

        delegate(document.body, 'click', '.carousel-next', (e, target) => {
            const carousel = target.closest('.vanilla-carousel');
            if (!carousel) return;
            const items = Array.from(carousel.querySelectorAll('.carousel-item'));
            const container = carousel.querySelector('.carousel-items-container');
            if (!container) return;
            const activeIdx = parseInt(container.getAttribute('data-active-index') || '0', 10);
            const newIndex = (activeIdx + 1) % items.length;
            updateCarousel(carousel, newIndex);
        });

        delegate(document.body, 'click', '.carousel-dot', (e, target) => {
            const carousel = target.closest('.vanilla-carousel');
            if (!carousel) return;
            const dots = Array.from(carousel.querySelectorAll('.carousel-dot'));
            const idx = dots.indexOf(target);
            if (idx !== -1) updateCarousel(carousel, idx);
        });
    }

    // Since events are bound to document.body globally once, we don't need to tear them down
    // when a single page component unmounts in HTMX. HTMX replaces body children, body stays.
    return () => {};
}
