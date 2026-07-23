import { bindRefs, delegate } from '../core/state.js';
import { getLocalItem, setLocalItem, getSessionItem, setSessionItem, removeSessionKey } from '../core/storage.js';

export function initTechBanner() {
    const banner = document.getElementById('tech-banner');
    if (!banner) return;
    
    if (getLocalItem('techBannerDismissed') === 'true') {
        banner.remove();
        return;
    }

    const refs = bindRefs(banner);

    if (getSessionItem('techBannerExpanded') === 'true' && refs.techInfo) {
        refs.techInfo.classList.remove('hidden');
    }

    // Delay entrance slightly
    setTimeout(() => {
        if (!banner) return;
        banner.classList.remove('translate-y-8', 'opacity-0', 'pointer-events-none');
        banner.classList.add('translate-y-0', 'opacity-100');
    }, 1000);
    
    const cleanup = delegate(banner, 'click', '[data-action]', (e, target) => {
        const action = target.dataset.action;
        
        if (action === 'closeTechBanner') {
            banner.classList.remove('translate-y-0', 'opacity-100');
            banner.classList.add('translate-y-8', 'opacity-0', 'pointer-events-none');
            setLocalItem('techBannerDismissed', 'true');
            removeSessionKey('techBannerExpanded');
            setTimeout(() => banner.remove(), 500);
        } else if (action === 'toggleTechInfo') {
            if (refs.techInfo) {
                const isHidden = refs.techInfo.classList.toggle('hidden');
                setSessionItem('techBannerExpanded', isHidden ? 'false' : 'true');
            }
        }
    });

    return () => {
        cleanup();
    };
}
