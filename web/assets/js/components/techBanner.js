import { bindRefs, delegate } from '../core/state.js';

export function initTechBanner() {
    const banner = document.getElementById('tech-banner');
    if (!banner) return;
    
    if (localStorage.getItem('techBannerDismissed') === 'true') {
        banner.remove();
        return;
    }

    // Delay entrance slightly
    setTimeout(() => {
        if (!banner) return;
        banner.classList.remove('translate-y-8', 'opacity-0', 'pointer-events-none');
        banner.classList.add('translate-y-0', 'opacity-100');
    }, 1000);

    const refs = bindRefs(banner);
    
    const cleanup = delegate(banner, 'click', '[data-action]', (e, target) => {
        const action = target.dataset.action;
        
        if (action === 'closeTechBanner') {
            banner.classList.remove('translate-y-0', 'opacity-100');
            banner.classList.add('translate-y-8', 'opacity-0', 'pointer-events-none');
            localStorage.setItem('techBannerDismissed', 'true');
            setTimeout(() => banner.remove(), 500);
        } else if (action === 'toggleTechInfo') {
            if (refs.techInfo) {
                refs.techInfo.classList.toggle('hidden');
            }
        }
    });

    return () => {
        cleanup();
    };
}
