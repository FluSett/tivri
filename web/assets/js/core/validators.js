export const MIN_NAME_LEN = 2;
export const MIN_SCOPE_LEN = 20;
export const MIN_DEADLINE_LEN = 2;
export const MIN_BUDGET_USD = 5;
export const MIN_EMAIL_LEN = 5;
export const MAX_EMAIL_LEN = 254;
export const MAX_FILE_SIZE_MB = 5;
export const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;

const EMAIL_REGEX = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

export function isValidName(text) {
    return text.trim().length >= MIN_NAME_LEN;
}

export function isValidServiceType(type) {
    return type.trim() !== '';
}

export function isValidScope(text) {
    return text.trim().length >= MIN_SCOPE_LEN;
}

export function isValidDeadline(text) {
    return text.trim().length >= MIN_DEADLINE_LEN;
}

export function isValidEmail(email) {
    if (!email) return false;
    const e = email.trim();
    return e.length >= MIN_EMAIL_LEN && e.length <= MAX_EMAIL_LEN && EMAIL_REGEX.test(e);
}

export function isValidBudget(budgetStr) {
    const s = budgetStr ? budgetStr.trim() : '';
    return s !== '' && !isNaN(s) && parseInt(s, 10) >= MIN_BUDGET_USD;
}

export function isValidFile(file) {
    return file.size <= MAX_FILE_SIZE_BYTES;
}
