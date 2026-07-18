export function toggleVisibility(element, isVisible) {
    if (!element) return;
    if (isVisible) {
        element.classList.remove('hidden');
    } else {
        element.classList.add('hidden');
    }
}

export function toggleClasses(element, condition, classesToAdd, classesToRemove) {
    if (!element) return;
    if (condition) {
        if (classesToAdd) element.classList.add(...classesToAdd);
        if (classesToRemove) element.classList.remove(...classesToRemove);
    } else {
        if (classesToAdd) element.classList.remove(...classesToAdd);
        if (classesToRemove) element.classList.add(...classesToRemove);
    }
}

export function setButtonSubmittingState(btn, textElement, spinnerElement, isSubmitting, isDisabled = false) {
    if (!btn) return;
    if (isSubmitting) {
        btn.classList.add('submitting');
        if (textElement) toggleClasses(textElement, true, ['opacity-0', 'hidden'], ['opacity-100']);
        if (spinnerElement)
            toggleClasses(
                spinnerElement,
                true,
                ['opacity-100', 'flex', 'items-center', 'space-x-2'],
                ['opacity-0', 'hidden']
            );
        btn.disabled = true;
    } else {
        btn.classList.remove('submitting');
        if (spinnerElement)
            toggleClasses(
                spinnerElement,
                true,
                ['opacity-0', 'hidden'],
                ['opacity-100', 'flex', 'items-center', 'space-x-2']
            );
        if (textElement) toggleClasses(textElement, true, ['opacity-100'], ['opacity-0', 'hidden']);
        btn.disabled = isDisabled;
    }
}
