import { createReactiveState, bindRefs, delegate } from '../core/state.js';
import { makeDraggable } from '../core/utils.js';

export function initLeadsTable() {
    const container = document.getElementById('leads-container');
    if (!container) return;

    const refs = bindRefs(container);
    const rows = Array.from(container.querySelectorAll('tbody tr'));

    const overflowWrap = container.querySelector('.overflow-x-auto');
    if (overflowWrap) makeDraggable(overflowWrap);

    const state = createReactiveState(
        'leads_table',
        {
            activeDropdown: null
        },
        (newState) => {
            renderUI();
        }
    );

    function getStatusClass(status) {
        if (status === 'pending') return 'text-blue-400 border-blue-500/20 bg-blue-500/5';
        if (status === 'active') return 'text-yellow-400 border-yellow-500/20 bg-yellow-500/5';
        if (status === 'done') return 'text-green-400 border-green-500/20 bg-green-500/5';
        if (status === 'canceled') return 'text-red-400 border-red-500/20 bg-red-500/5';
        return '';
    }

    function toggleDropdownVisibility(dropdownElement, isOpen) {
        if (!dropdownElement) return;
        const btn =
            dropdownElement.previousElementSibling ||
            document.querySelector(`[data-ref="${dropdownElement.getAttribute('data-ref').replace('Menu_', 'Btn_')}"]`);

        if (isOpen) {
            if (btn && dropdownElement.parentElement !== document.body) {
                dropdownElement._originalParent = dropdownElement.parentElement;
                document.body.appendChild(dropdownElement);
                dropdownElement.classList.add('tivri-dropdown-zombie');
            }

            dropdownElement.classList.remove('bottom-full', 'mb-1.5', 'mt-1.5');

            if (btn) {
                dropdownElement.classList.remove('hidden');
                dropdownElement.style.position = 'fixed';
                const dropHeight = dropdownElement.offsetHeight || 150;

                requestAnimationFrame(() => {
                    dropdownElement.classList.remove('opacity-0', 'translate-y-2', 'pointer-events-none', 'delay-200');
                    dropdownElement.classList.add('opacity-100', 'translate-y-0', 'pointer-events-auto');
                });

                function updatePos() {
                    if (dropdownElement.classList.contains('hidden')) return;
                    const r = btn.getBoundingClientRect();
                    dropdownElement.style.left = r.left + 'px';
                    dropdownElement.style.width = Math.max(r.width, 150) + 'px';

                    if (r.bottom + dropHeight + 10 > window.innerHeight) {
                        dropdownElement.style.top = r.top - dropHeight - 6 + 'px';
                    } else {
                        dropdownElement.style.top = r.bottom + 6 + 'px';
                    }
                    dropdownElement._frame = requestAnimationFrame(updatePos);
                }
                updatePos();
            } else {
                dropdownElement.classList.remove('hidden');
                requestAnimationFrame(() => {
                    dropdownElement.classList.remove('opacity-0', 'translate-y-2', 'pointer-events-none', 'delay-200');
                    dropdownElement.classList.add('opacity-100', 'translate-y-0', 'pointer-events-auto');
                });
            }
        } else {
            if (!dropdownElement.classList.contains('opacity-100')) return;
            if (dropdownElement._frame) cancelAnimationFrame(dropdownElement._frame);
            dropdownElement.classList.add('opacity-0', 'translate-y-2', 'pointer-events-none', 'delay-200');
            dropdownElement.classList.remove('opacity-100', 'translate-y-0', 'pointer-events-auto');
            setTimeout(() => {
                if (!dropdownElement.classList.contains('opacity-100')) {
                    if (dropdownElement._originalParent && dropdownElement.parentElement === document.body) {
                        dropdownElement._originalParent.appendChild(dropdownElement);
                    }
                    dropdownElement.classList.remove('tivri-dropdown-zombie');
                    dropdownElement.style.position = 'absolute';
                    dropdownElement.style.top = '';
                    dropdownElement.style.left = '';
                    dropdownElement.style.width = '100%';
                    dropdownElement.classList.add('mt-1.5', 'hidden');
                }
            }, 300);
        }
    }

    function renderUI() {
        toggleDropdownVisibility(refs.serviceFilterDropdown, state.activeDropdown === 'serviceFilter');
        toggleDropdownVisibility(refs.clientFilterDropdown, state.activeDropdown === 'clientFilter');
        toggleDropdownVisibility(refs.internalFilterDropdown, state.activeDropdown === 'internalFilter');
        toggleDropdownVisibility(refs.sortDropdown, state.activeDropdown === 'sort');

        rows.forEach((row) => {
            const id = row.getAttribute('data-id');
            const clientMenu = document.querySelector(`[data-ref="clientMenu_${id}"]`);
            const internalMenu = document.querySelector(`[data-ref="internalMenu_${id}"]`);
            toggleDropdownVisibility(clientMenu, state.activeDropdown === `clientMenu_${id}`);
            toggleDropdownVisibility(internalMenu, state.activeDropdown === `internalMenu_${id}`);
        });
    }

    const teardowns = [];

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleServiceFilter"]', (e) => {
            e.stopPropagation();
            state.activeDropdown = state.activeDropdown === 'serviceFilter' ? null : 'serviceFilter';
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleClientFilter"]', (e) => {
            e.stopPropagation();
            state.activeDropdown = state.activeDropdown === 'clientFilter' ? null : 'clientFilter';
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleInternalFilter"]', (e) => {
            e.stopPropagation();
            state.activeDropdown = state.activeDropdown === 'internalFilter' ? null : 'internalFilter';
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleSort"]', (e) => {
            e.stopPropagation();
            state.activeDropdown = state.activeDropdown === 'sort' ? null : 'sort';
        })
    );

    teardowns.push(
        delegate(document.body, 'click', '.dropdown-option', (e, target) => {
            e.stopPropagation();
            const action = target.getAttribute('data-action');
            const val = target.getAttribute('data-val');
            if (action === 'setServiceFilter') {
                if (refs.serviceTypeInput) refs.serviceTypeInput.value = val;
                if (refs.serviceFilterText) refs.serviceFilterText.textContent = target.textContent.trim();
            } else if (action === 'setClientFilter') {
                if (refs.clientStatusInput) refs.clientStatusInput.value = val;
                if (refs.clientFilterText) refs.clientFilterText.textContent = target.textContent.trim();
            } else if (action === 'setInternalFilter') {
                if (refs.internalStatusInput) refs.internalStatusInput.value = val;
                if (refs.internalFilterText) refs.internalFilterText.textContent = target.textContent.trim();
            } else if (action === 'setSort') {
                if (refs.sortInput) refs.sortInput.value = val;
                if (refs.sortText) refs.sortText.textContent = target.textContent.trim();
            }

            if (['setServiceFilter', 'setClientFilter', 'setInternalFilter', 'setSort'].includes(action)) {
                if (refs.pageInput) refs.pageInput.value = '1';
                state.activeDropdown = null;
                if (window.htmx) {
                    htmx.trigger(container, 'change');
                } else {
                    container.dispatchEvent(new Event('change', { bubbles: true }));
                }
            } else if (action === 'setClientStatus') {
                const dropdown = target.closest('.panel-animate');
                const ref = dropdown ? dropdown.getAttribute('data-ref') : '';
                const id = ref ? ref.split('_')[1] : null;
                if (id) updateStatus(id, 'client', val);
            } else if (action === 'setInternalStatus') {
                const dropdown = target.closest('.panel-animate');
                const ref = dropdown ? dropdown.getAttribute('data-ref') : '';
                const id = ref ? ref.split('_')[1] : null;
                if (id) updateStatus(id, 'internal', val);
            }
        })
    );

    function updateStatus(leadId, type, status) {
        const headers = { 'Content-Type': 'application/x-www-form-urlencoded' };
        const csrfInput = document.querySelector('input[name="gorilla.csrf.Token"]');
        if (csrfInput) {
            headers['X-CSRF-Token'] = csrfInput.value;
        }

        fetch('/admin/leads/status', {
            method: 'POST',
            headers: headers,
            body: new URLSearchParams({ id: leadId, type: type, status: status })
        });

        const row = rows.find((r) => r.getAttribute('data-id') === String(leadId));
        if (row) {
            if (type === 'client') {
                row.setAttribute('data-client-status', status);
                const btn = row.querySelector('[data-action="toggleClientStatus"]');
                const span = btn.querySelector('[data-ref="clientText"]');
                btn.className = `badge-status w-full justify-between status-${status} ${getStatusClass(status)}`;
                span.textContent = status.toUpperCase();
            } else {
                row.setAttribute('data-internal-status', status);
                const btn = row.querySelector('[data-action="toggleInternalStatus"]');
                const span = btn.querySelector('[data-ref="internalText"]');
                btn.className = `badge-status w-full justify-between status-${status} ${getStatusClass(status)}`;
                span.textContent = status.toUpperCase();
            }
        }
        state.activeDropdown = null;
    }

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleClientStatus"]', (e, target) => {
            e.stopPropagation();
            const id = target.getAttribute('data-id');
            state.activeDropdown = state.activeDropdown === `clientMenu_${id}` ? null : `clientMenu_${id}`;
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleInternalStatus"]', (e, target) => {
            e.stopPropagation();
            const id = target.getAttribute('data-id');
            state.activeDropdown = state.activeDropdown === `internalMenu_${id}` ? null : `internalMenu_${id}`;
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleExpand"]', (e, target) => {
            e.stopPropagation();
            const collapsed = target.querySelector('.collapsed');
            const expanded = target.querySelector('.expanded');
            if (collapsed && expanded) {
                collapsed.classList.toggle('hidden');
                expanded.classList.toggle('hidden');
            } else {
                target.classList.toggle('line-clamp-2');
            }
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="setPage"]', (e, target) => {
            e.stopPropagation();
            const page = target.getAttribute('data-page');
            if (refs.pageInput && page) {
                refs.pageInput.value = page;
                if (window.htmx) {
                    htmx.trigger(container, 'change');
                } else {
                    refs.pageInput.dispatchEvent(new Event('change', { bubbles: true }));
                }
            }
        })
    );

    const hClickOutside = () => {
        if (state.activeDropdown !== null) {
            state.activeDropdown = null;
        }
    };
    document.addEventListener('click', hClickOutside);
    teardowns.push(() => document.removeEventListener('click', hClickOutside));

    renderUI();

    return () => teardowns.forEach((fn) => fn());
}
