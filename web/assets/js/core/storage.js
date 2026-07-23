// Session Storage Wrappers
export function setSessionItem(key, val) {
    try {
        sessionStorage.setItem(key, val);
    } catch (e) {
        console.error('Failed to set session item:', e);
    }
}

export function getSessionItem(key) {
    try {
        return sessionStorage.getItem(key);
    } catch (e) {
        console.error('Failed to get session item:', e);
        return null;
    }
}

export function setSessionJSON(key, data) {
    try {
        sessionStorage.setItem(key, JSON.stringify(data));
    } catch (e) {
        console.error('Failed to set session JSON:', e);
    }
}

export function getSessionJSON(key) {
    try {
        const item = sessionStorage.getItem(key);
        return item ? JSON.parse(item) : null;
    } catch (e) {
        console.error('Failed to get session JSON:', e);
        return null;
    }
}

export function removeSessionKey(key) {
    try {
        sessionStorage.removeItem(key);
    } catch (e) {
        console.error('Failed to remove session key:', e);
    }
}

// Local Storage Wrappers
export function setLocalItem(key, val) {
    try {
        localStorage.setItem(key, val);
    } catch (e) {
        console.error('Failed to set local item:', e);
    }
}

export function getLocalItem(key) {
    try {
        return localStorage.getItem(key);
    } catch (e) {
        console.error('Failed to get local item:', e);
        return null;
    }
}

export function setLocalJSON(key, data) {
    try {
        localStorage.setItem(key, JSON.stringify(data));
    } catch (e) {
        console.error('Failed to set local JSON:', e);
    }
}

export function getLocalJSON(key) {
    try {
        const item = localStorage.getItem(key);
        return item ? JSON.parse(item) : null;
    } catch (e) {
        console.error('Failed to get local JSON:', e);
        return null;
    }
}

export function removeLocalKey(key) {
    try {
        localStorage.removeItem(key);
    } catch (e) {
        console.error('Failed to remove local key:', e);
    }
}

// Backward Compatibility Aliases
export const setStorageItem = setSessionItem;
export const getStorageItem = getSessionItem;
export const setStorageJSON = setSessionJSON;
export const getStorageJSON = getSessionJSON;
export const removeStorageKey = removeSessionKey;

export function persistState(newState, keysMap) {
    for (const [stateKey, storageKey] of Object.entries(keysMap)) {
        const val = newState[stateKey];
        if (typeof val === 'object' && val !== null) {
            setSessionJSON(storageKey, val);
        } else if (val !== undefined && val !== null) {
            setSessionItem(storageKey, String(val));
        }
    }
}

export function clearStorageKeys(keys) {
    for (const key of keys) {
        removeSessionKey(key);
    }
}
