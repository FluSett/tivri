import { createReactiveState, bindRefs, delegate } from './state.js';
import { setSessionItem } from './storage.js';

export function initLayout() {
    const refs = bindRefs(document.body);
    const teardowns = [];

    // Focus management handles the instant hover states natively

    const state = createReactiveState(
        'layout',
        {
            mobileMenuOpen: false,
            langDropdownOpen: false
        },
        (newState) => {
            if (newState.mobileMenuOpen) {
                let rect;
                if (refs.openMenuBtn) {
                    rect = refs.openMenuBtn.getBoundingClientRect();
                }
                document.body.classList.add('overflow-hidden');
                if (rect && refs.closeMenuBtn) {
                    refs.closeMenuBtn.style.top = rect.top + 'px';
                    refs.closeMenuBtn.style.left = rect.left + 'px';
                    setTimeout(() => {
                        refs.closeMenuBtn.focus({ preventScroll: true });
                        document.body.style.cursor = 'pointer';
                        window.addEventListener('mousemove', () => (document.body.style.cursor = ''), { once: true });
                    }, 10);
                }
                if (refs.header) refs.header.classList.add('header-hidden');
                if (refs.mobileMenuOverlay) {
                    refs.mobileMenuOverlay.classList.remove('hidden');
                    setTimeout(() => {
                        if (refs.mobileMenuOverlay) {
                            refs.mobileMenuOverlay.classList.remove('opacity-0', 'pointer-events-none', 'delay-300');
                            refs.mobileMenuOverlay.classList.add('opacity-100', 'pointer-events-auto');
                        }
                    }, 10);
                }
                if (refs.mobileMenuPanel) {
                    refs.mobileMenuPanel.classList.remove('translate-x-full');
                    refs.mobileMenuPanel.classList.add('translate-x-0');
                }
            } else {
                document.body.classList.remove('overflow-hidden');
                if (refs.header) {
                    refs.header.classList.remove('header-hidden');
                    setTimeout(() => {
                        if (refs.openMenuBtn) {
                            refs.openMenuBtn.focus({ preventScroll: true });
                            document.body.style.cursor = 'pointer';
                            window.addEventListener('mousemove', () => (document.body.style.cursor = ''), {
                                once: true
                            });
                        }
                    }, 10);
                }
                if (refs.mobileMenuOverlay) {
                    refs.mobileMenuOverlay.classList.remove('opacity-100', 'pointer-events-auto');
                    refs.mobileMenuOverlay.classList.add('opacity-0', 'pointer-events-none', 'delay-300');
                    setTimeout(() => {
                        if (!state.mobileMenuOpen && refs.mobileMenuOverlay) {
                            refs.mobileMenuOverlay.classList.add('hidden');
                        }
                    }, 300);
                }
                if (refs.mobileMenuPanel) {
                    refs.mobileMenuPanel.classList.remove('translate-x-0');
                    refs.mobileMenuPanel.classList.add('translate-x-full');
                }
            }

            if (newState.langDropdownOpen) {
                if (refs.langDropdownMenu) {
                    refs.langDropdownMenu.classList.remove('hidden');
                    setTimeout(() => {
                        if (refs.langDropdownMenu) {
                            refs.langDropdownMenu.classList.remove(
                                'opacity-0',
                                'scale-95',
                                'pointer-events-none',
                                'delay-200'
                            );
                            refs.langDropdownMenu.classList.add('opacity-100', 'scale-100', 'pointer-events-auto');
                        }
                    }, 10);
                }
                if (refs.langDropdownIcon) refs.langDropdownIcon.classList.add('rotate-180');
            } else {
                if (refs.langDropdownMenu) {
                    refs.langDropdownMenu.classList.remove('opacity-100', 'scale-100', 'pointer-events-auto');
                    refs.langDropdownMenu.classList.add('opacity-0', 'scale-95', 'pointer-events-none', 'delay-200');
                    setTimeout(() => {
                        if (!state.langDropdownOpen && refs.langDropdownMenu) {
                            refs.langDropdownMenu.classList.add('hidden');
                        }
                    }, 200);
                }
                if (refs.langDropdownIcon) refs.langDropdownIcon.classList.remove('rotate-180');
            }
        }
    );

    teardowns.push(
        delegate(document.body, 'click', '[data-action="openMenu"]', () => {
            state.mobileMenuOpen = true;
        })
    );

    teardowns.push(
        delegate(document.body, 'click', '[data-action="closeMenu"]', () => {
            state.mobileMenuOpen = false;
        })
    );

    const closeHandler = () => {
        state.mobileMenuOpen = false;
    };
    window.addEventListener('tivri-close-menu', closeHandler);
    teardowns.push(() => window.removeEventListener('tivri-close-menu', closeHandler));

    const handleResize = () => {
        if (window.innerWidth >= 768 && state.mobileMenuOpen) {
            state.mobileMenuOpen = false;
        }
    };
    window.addEventListener('resize', handleResize);
    teardowns.push(() => window.removeEventListener('resize', handleResize));

    teardowns.push(
        delegate(document.body, 'click', '[data-action="toggleLang"]', (e) => {
            e.stopPropagation();
            state.langDropdownOpen = !state.langDropdownOpen;
        })
    );

    const handleClickOutside = (e) => {
        if (
            state.langDropdownOpen &&
            refs.langDropdownMenu &&
            !refs.langDropdownMenu.contains(e.target) &&
            (!refs.langDropdownBtn || !refs.langDropdownBtn.contains(e.target))
        ) {
            state.langDropdownOpen = false;
        }
    };
    document.addEventListener('click', handleClickOutside);
    teardowns.push(() => document.removeEventListener('click', handleClickOutside));

    teardowns.push(
        delegate(document.body, 'click', '.lang-switch-btn', (e, btn) => {
            const lang = btn.getAttribute('data-lang');
            setSessionItem('locale_change', 'true');
            state.mobileMenuOpen = false;
            setTimeout(() => {
                window.htmx.ajax('GET', '/api/lang?lang=' + lang, {
                    target: 'body',
                    swap: 'outerHTML transition:true show:none'
                });
            }, 350);
        })
    );

    teardowns.push(
        delegate(document.body, 'click', '.lang-switch-desktop', () => {
            setSessionItem('locale_change', 'true');
            state.langDropdownOpen = false;
        })
    );

    teardowns.push(
        delegate(document.body, 'click', '.expandable-text', (e, expander) => {
            const collapsed = expander.querySelector('.collapsed');
            const expanded = expander.querySelector('.expanded');
            if (collapsed && expanded) {
                collapsed.classList.toggle('hidden');
                collapsed.classList.toggle('block');
                expanded.classList.toggle('hidden');
                expanded.classList.toggle('block');
            }
        })
    );

    return () => {
        teardowns.forEach((fn) => fn());
    };
}
