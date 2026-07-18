import { createReactiveState, bindRefs, delegate } from '../core/state.js';
import { makeDraggable } from '../core/utils.js';

export function initMessagesTable() {
    const container = document.getElementById('messages-container');
    if (!container) return;

    const refs = bindRefs(container);
    const rows = Array.from(container.querySelectorAll('tbody tr'));

    const overflowWrap = container.querySelector('.overflow-x-auto');
    if (overflowWrap) makeDraggable(overflowWrap);

    const state = createReactiveState(
        'messages_table',
        {
            activeDropdown: null
        },
        (newState) => {
            renderUI();
        }
    );

    function getStatusClass(status) {
        if (status === 'new') return 'text-blue-400 border-blue-500/20 bg-blue-500/5';
        if (status === 'answered') return 'text-yellow-400 border-yellow-500/20 bg-yellow-500/5';
        if (status === 'done') return 'text-green-400 border-green-500/20 bg-green-500/5';
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
        toggleDropdownVisibility(refs.statusFilterDropdown, state.activeDropdown === 'statusFilter');
        toggleDropdownVisibility(refs.sortDropdown, state.activeDropdown === 'sort');

        rows.forEach((row) => {
            const id = row.getAttribute('data-id');
            const statusMenu = document.querySelector(`[data-ref="statusMenu_${id}"]`);
            toggleDropdownVisibility(statusMenu, state.activeDropdown === `statusMenu_${id}`);
        });
    }

    const teardowns = [];

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleStatusFilter"]', (e) => {
            e.stopPropagation();
            state.activeDropdown = state.activeDropdown === 'statusFilter' ? null : 'statusFilter';
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleSort"]', (e) => {
            e.stopPropagation();
            state.activeDropdown = state.activeDropdown === 'sort' ? null : 'sort';
        })
    );

    teardowns.push(
        delegate(container, 'click', '[data-action="toggleRowStatus"]', (e, target) => {
            e.stopPropagation();
            const tr = target.closest('tr');
            const id = tr ? tr.getAttribute('data-id') : null;
            if (id) state.activeDropdown = state.activeDropdown === 'statusMenu_' + id ? null : 'statusMenu_' + id;
        })
    );

    teardowns.push(
        delegate(document.body, 'click', '.dropdown-option', (e, target) => {
            e.stopPropagation();
            const action = target.getAttribute('data-action');
            const val = target.getAttribute('data-val');

            if (action === 'setStatusFilter') {
                if (refs.statusFilterInput) refs.statusFilterInput.value = val;
                if (refs.statusFilterText) refs.statusFilterText.textContent = target.textContent.trim();
                if (refs.pageInput) refs.pageInput.value = '1';
                state.activeDropdown = null;
                if (window.htmx) htmx.trigger(container, 'change');
            } else if (action === 'setSort') {
                if (refs.sortInput) refs.sortInput.value = val;
                if (refs.sortText) refs.sortText.textContent = target.textContent.trim();
                if (refs.pageInput) refs.pageInput.value = '1';
                state.activeDropdown = null;
                if (window.htmx) htmx.trigger(container, 'change');
            } else if (action === 'updateRowStatus') {
                const dropdown = target.closest('.panel-animate');
                const ref = dropdown ? dropdown.getAttribute('data-ref') : '';
                const id = ref ? ref.split('_')[1] : null;
                if (id) updateStatus(id, val);
            }
        })
    );

    function updateStatus(msgId, status) {
        const headers = { 'Content-Type': 'application/x-www-form-urlencoded' };
        const csrfInput = document.querySelector('input[name="gorilla.csrf.Token"]');
        if (csrfInput) {
            headers['X-CSRF-Token'] = csrfInput.value;
        }

        fetch('/admin/messages/status', {
            method: 'POST',
            headers: headers,
            body: new URLSearchParams({ id: msgId, status: status })
        });

        const row = rows.find((r) => r.getAttribute('data-id') === String(msgId));
        if (row) {
            const btn = row.querySelector('[data-action="toggleRowStatus"]');
            const span = btn.querySelector('[data-ref="statusText"]');
            btn.className = `badge-status w-full justify-between status-${status}`;
            span.textContent = status.toUpperCase();
        }
        state.activeDropdown = null;
    }

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
                if (window.htmx) htmx.trigger(container, 'change');
            }
        })
    );

    const handleClickOutside = () => {
        if (state.activeDropdown !== null) {
            state.activeDropdown = null;
        }
    };
    document.addEventListener('click', handleClickOutside);
    teardowns.push(() => document.removeEventListener('click', handleClickOutside));

    renderUI();

    return () => teardowns.forEach((fn) => fn());
}
