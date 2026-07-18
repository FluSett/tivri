export function persistState(newState, keysMap) {
    for (const [stateKey, storageKey] of Object.entries(keysMap)) {
        const val = newState[stateKey];
        if (typeof val === 'object' && val !== null) {
            sessionStorage.setItem(storageKey, JSON.stringify(val));
        } else {
            sessionStorage.setItem(storageKey, val);
        }
    }
}

export function clearStorageKeys(keys) {
    for (const key of keys) {
        sessionStorage.removeItem(key);
    }
}
