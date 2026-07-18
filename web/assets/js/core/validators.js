export const MIN_NAME_LEN = 2;
export const MIN_SCOPE_LEN = 20;
export const MIN_DEADLINE_LEN = 2;
export const MIN_BUDGET_USD = 100;
export const MIN_EMAIL_LEN = 5;
export const MAX_FILE_SIZE_MB = 5;
export const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;

export function isValidName(text) {
    return text.trim().length >= MIN_NAME_LEN;
}

export function isValidScope(text) {
    return text.trim().length >= MIN_SCOPE_LEN;
}

export function isValidDeadline(text) {
    return text.trim().length >= MIN_DEADLINE_LEN;
}

export function isValidEmail(email) {
    const e = email.trim();
    return e.length >= MIN_EMAIL_LEN && e.includes('@') && e.includes('.');
}

export function isValidBudget(budgetType, customBudgetStr) {
    if (budgetType === '') return false;
    if (budgetType === 'other') {
        return (
            customBudgetStr.trim() !== '' && !isNaN(customBudgetStr) && parseInt(customBudgetStr, 10) >= MIN_BUDGET_USD
        );
    }
    return true;
}

export function isValidFile(file) {
    return file.size <= MAX_FILE_SIZE_BYTES;
}
