export function setStorageItem(key, val) {
    try {
        sessionStorage.setItem(key, val);
    } catch (e) {
        console.error('Failed to set storage item:', e);
    }
}

export function getStorageItem(key) {
    try {
        return sessionStorage.getItem(key);
    } catch (e) {
        console.error('Failed to get storage item:', e);
        return null;
    }
}

export function setStorageJSON(key, data) {
    try {
        sessionStorage.setItem(key, JSON.stringify(data));
    } catch (e) {
        console.error('Failed to set storage JSON:', e);
    }
}

export function getStorageJSON(key) {
    try {
        const item = sessionStorage.getItem(key);
        return item ? JSON.parse(item) : null;
    } catch (e) {
        console.error('Failed to get storage JSON:', e);
        return null;
    }
}

export function removeStorageKey(key) {
    try {
        sessionStorage.removeItem(key);
    } catch (e) {
        console.error('Failed to remove storage key:', e);
    }
}

export function persistState(newState, keysMap) {
    for (const [stateKey, storageKey] of Object.entries(keysMap)) {
        const val = newState[stateKey];
        if (typeof val === 'object' && val !== null) {
            setStorageJSON(storageKey, val);
        } else if (val !== undefined && val !== null) {
            setStorageItem(storageKey, String(val));
        }
    }
}

export function clearStorageKeys(keys) {
    for (const key of keys) {
        removeStorageKey(key);
    }
}
