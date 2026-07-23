/**
 * Creates a reactive state proxy that automatically calls onUpdate when modified.
 * @param {string|Object} idOrState - A string ID for persistence, or the initialState object.
 * @param {Object|Function} stateOrUpdate - If ID provided, this is initialState. Otherwise, it's onUpdate.
 * @param {Function} [onUpdateCallback] - If ID provided, this is onUpdate.
 * @returns {Proxy} A reactive proxy wrapping the state
 */
export function createReactiveState(idOrState, stateOrUpdate, onUpdateCallback) {
    let id = null;
    let initialState;
    let options = {};
    let onUpdate;

    if (typeof idOrState === 'string') {
        id = idOrState;
        initialState = stateOrUpdate;

        if (typeof onUpdateCallback === 'function') {
            onUpdate = onUpdateCallback;
        } else if (typeof onUpdateCallback === 'object') {
            options = onUpdateCallback;
            onUpdate = arguments[3];
        }
    } else {
        initialState = idOrState;
        onUpdate = stateOrUpdate;
    }

    if (id) {
        if (!window.__tivriPersistedStates) window.__tivriPersistedStates = {};
        if (window.__tivriPersistedStates[id]) {
            // Restore from persistence, merge so we don't lose new keys
            const persisted = { ...window.__tivriPersistedStates[id] };
            if (options.ephemeralKeys) {
                options.ephemeralKeys.forEach((key) => delete persisted[key]);
            }
            initialState = { ...initialState, ...persisted };
        } else {
            // Initialize persistence
            const toPersist = { ...initialState };
            if (options.ephemeralKeys) {
                options.ephemeralKeys.forEach((key) => delete toPersist[key]);
            }
            window.__tivriPersistedStates[id] = toPersist;
        }
    }

    let updateScheduled = false;

    return new Proxy(initialState, {
        set(target, property, value) {
            if (target[property] === value) return true;
            target[property] = value;

            if (id && window.__tivriPersistedStates) {
                if (!options.ephemeralKeys || !options.ephemeralKeys.includes(property)) {
                    window.__tivriPersistedStates[id][property] = value;
                }
            }

            if (!updateScheduled) {
                updateScheduled = true;
                Promise.resolve().then(() => {
                    updateScheduled = false;
                    onUpdate(target);
                });
            }
            return true;
        }
    });
}

/**
 * Scans a container for elements with data-ref="..." and returns a mapped object.
 * Also finds elements with data-ref-list="..." and maps them to arrays of elements.
 * @param {HTMLElement} container - The DOM element to search within
 * @returns {Object} An object mapping ref names to DOM elements
 */
export function bindRefs(container) {
    const refs = {};

    container.querySelectorAll('[data-ref]').forEach((el) => {
        refs[el.dataset.ref] = el;
    });

    container.querySelectorAll('[data-ref-list]').forEach((el) => {
        const listName = el.dataset.refList;
        if (!refs[listName]) refs[listName] = [];
        refs[listName].push(el);
    });

    return refs;
}

/**
 * Attaches a delegated event listener to a container.
 * @param {HTMLElement} container - The container to attach the listener to
 * @param {string} eventName - The event to listen for (e.g. 'click')
 * @param {string} selector - The selector to match targets against (e.g. '[data-action]')
 * @param {Function} handler - The callback function receiving (event, matchedTarget)
 */
export function delegate(container, eventName, selector, handler) {
    const fn = (e) => {
        const target = e.target.closest(selector);
        if (target && container.contains(target)) {
            handler(e, target);
        }
    };
    container.addEventListener(eventName, fn);
    return () => container.removeEventListener(eventName, fn);
}

import { getSessionItem } from './storage.js';

/**
 * Initializes global state persistence logic.
 * Clears in-memory persisted states on standard HTMX navigations (not locale changes).
 */
export function initStatePersistence() {
    document.addEventListener('htmx:beforeSwap', (e) => {
        const isMainNavigation = e.detail.target && e.detail.target.id === 'app-body';
        if (isMainNavigation && getSessionItem('locale_change') !== 'true') {
            window.__tivriPersistedStates = {};
        }
    });
}
