export function makeDraggable(el) {
    if (!el) return;
    let isDown = false;
    let startX;
    let scrollLeft;

    el.classList.add('cursor-grab');

    el.addEventListener('mousedown', (e) => {
        if (e.button !== 0) return; // Only left click
        // Allow text selection if holding shift, or clicking inside interactive elements
        if (
            e.shiftKey ||
            e.target.closest(
                'button, a, input, select, textarea, [data-action="toggleExpand"], [data-action="toggleClientStatus"], [data-action="toggleInternalStatus"]'
            )
        )
            return;

        isDown = true;
        el.classList.remove('cursor-grab');
        el.classList.add('cursor-grabbing', 'select-none');
        document.body.classList.add('select-none');
        window.getSelection().removeAllRanges();

        startX = e.pageX - el.offsetLeft;
        scrollLeft = el.scrollLeft;
    });

    const stopDragging = () => {
        if (!isDown) return;
        isDown = false;
        el.classList.remove('cursor-grabbing', 'select-none');
        el.classList.add('cursor-grab');
        document.body.classList.remove('select-none');
    };

    el.addEventListener('mouseleave', stopDragging);
    el.addEventListener('mouseup', stopDragging);

    el.addEventListener('mousemove', (e) => {
        if (!isDown) return;
        e.preventDefault();
        const x = e.pageX - el.offsetLeft;
        const walk = (x - startX) * 1.5; // multiplier
        el.scrollLeft = scrollLeft - walk;
    });
}
