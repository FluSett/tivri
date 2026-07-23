import { createReactiveState, bindRefs, delegate } from '../core/state.js';
import { toggleVisibility, toggleClasses } from '../core/dom.js';
import { MAX_FILE_SIZE_BYTES } from '../core/validators.js';

// Removed STORAGE_MAP

export function initPortfolioForm() {
    const form = document.getElementById('portfolio-form');
    if (!form) return;

    const refs = bindRefs(document);

    const tagTemplate = document.getElementById('pf-tag-template');
    const mediaRowTemplate = document.getElementById('pf-media-row-template');
    const previewTagTemplate = document.getElementById('pf-preview-tag-template');

    const state = createReactiveState(
        'portfolio_form',
        {
            title: '',
            description: '',
            techTags: [],
            mediaPreviews: [],
            activeIndex: 0
        },
        (newState) => {
            updateUI();
        }
    );

    let lastMediaFingerprint = '';
    let lastTagsFingerprint = '';

    function addTag(val) {
        let added = false;
        val.split(',').forEach((t) => {
            const trimmed = t.trim();
            if (trimmed && !state.techTags.includes(trimmed)) {
                state.techTags.push(trimmed);
                added = true;
            }
        });
        if (added) state.techTags = [...state.techTags];
    }

    function removeTag(index) {
        state.techTags.splice(index, 1);
        state.techTags = [...state.techTags];
    }

    function handleFiles(files) {
        for (let i = 0; i < files.length; i++) {
            const f = files[i];
            if (f.size > MAX_FILE_SIZE_BYTES) {
                window.dispatchEvent(
                    new CustomEvent('tivri-error', { detail: `File ${f.name} exceeds maximum size of 5MB.` })
                );
                continue;
            }
            state.mediaPreviews.push({
                id: Math.random().toString(36).substr(2, 9),
                file: f,
                name: f.name,
                size: (f.size / 1024 / 1024).toFixed(2) + ' MB',
                url: URL.createObjectURL(f),
                isVideo: f.type.startsWith('video/')
            });
        }
        refs.mediaInput.value = '';
        state.activeIndex = 0;
        state.mediaPreviews = [...state.mediaPreviews];
    }

    function moveItem(index, direction) {
        const targetIndex = index + direction;
        if (targetIndex < 0 || targetIndex >= state.mediaPreviews.length) return;
        const temp = state.mediaPreviews[index];
        state.mediaPreviews[index] = state.mediaPreviews[targetIndex];
        state.mediaPreviews[targetIndex] = temp;
        state.activeIndex = 0;
        state.mediaPreviews = [...state.mediaPreviews];
    }

    function removeItem(index) {
        URL.revokeObjectURL(state.mediaPreviews[index].url);
        state.mediaPreviews.splice(index, 1);
        if (state.activeIndex >= state.mediaPreviews.length) {
            state.activeIndex = Math.max(0, state.mediaPreviews.length - 1);
        }
        state.mediaPreviews = [...state.mediaPreviews];
    }

    function updateUI() {
        if (refs.title.value !== state.title) refs.title.value = state.title;
        if (refs.desc.value !== state.description) refs.desc.value = state.description;
        refs.descCount.textContent = state.description.length + ' / 500';

        refs.previewTitle.textContent = state.title || 'Project Title';
        refs.previewDesc.textContent =
            state.description ||
            'A brief description of your project will be displayed here, showcasing your work and achievements.';

        const currentTagsFingerprint = state.techTags.join(',');
        if (currentTagsFingerprint !== lastTagsFingerprint) {
            refs.tagsContainer.innerHTML = '';
            refs.previewTags.innerHTML = '';

            state.techTags.forEach((tag, idx) => {
                const tagClone = tagTemplate.content.cloneNode(true);
                const textSpan = tagClone.querySelector('[data-ref="text"]');
                textSpan.textContent = tag;
                const rmvBtn = tagClone.querySelector('[data-action="removeTag"]');
                rmvBtn.setAttribute('data-index', idx);
                refs.tagsContainer.appendChild(tagClone);

                const pTagClone = previewTagTemplate.content.cloneNode(true);
                pTagClone.querySelector('[data-ref="text"]').textContent = tag;
                refs.previewTags.appendChild(pTagClone);
            });
            lastTagsFingerprint = currentTagsFingerprint;
        }

        toggleClasses(refs.previewTagsDefault, state.techTags.length > 0, ['hidden'], []);

        toggleVisibility(refs.mediaListContainer, state.mediaPreviews.length > 0);
        if (state.mediaPreviews.length > 0) {
            refs.mediaList.innerHTML = '';
            state.mediaPreviews.forEach((item, idx) => {
                const clone = mediaRowTemplate.content.cloneNode(true);
                clone.querySelector('[data-ref="name"]').textContent = item.name;
                clone.querySelector('[data-ref="size"]').textContent = item.size;

                const preview = clone.querySelector('[data-ref="preview"]');
                preview.innerHTML = item.isVideo
                    ? `<div class="text-neutral-400"><svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h6a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2V6zM14.553 7.106A1 1 0 0014 8v4a1 1 0 00.553.894l2 1A1 1 0 0018 13V7a1 1 0 00-1.447-.894l-2 1z"></path></svg></div>`
                    : `<img src="${item.url}" class="w-full h-full object-cover">`;

                const btnUp = clone.querySelector('[data-action="moveUp"]');
                const btnDown = clone.querySelector('[data-action="moveDown"]');
                const btnRmv = clone.querySelector('[data-action="removeMedia"]');

                btnUp.setAttribute('data-index', idx);
                btnDown.setAttribute('data-index', idx);
                btnRmv.setAttribute('data-index', idx);

                if (idx === 0) btnUp.disabled = true;
                if (idx === state.mediaPreviews.length - 1) btnDown.disabled = true;

                refs.mediaList.appendChild(clone);
            });
        } else {
            refs.mediaList.innerHTML = '';
        }

        toggleVisibility(refs.previewCarousel, state.mediaPreviews.length > 0);

        const currentMediaFingerprint = state.mediaPreviews.map((m) => m.id).join(',');
        if (currentMediaFingerprint !== lastMediaFingerprint) {
            if (state.mediaPreviews.length > 0) {
                refs.previewCarouselItems.innerHTML = '';

                state.mediaPreviews.forEach((item, idx) => {
                    const itemDiv = document.createElement('div');
                    itemDiv.className = 'absolute inset-0 w-full h-full transition-transform duration-500 ease-in-out';
                    itemDiv.innerHTML = item.isVideo
                        ? `<video src="${item.url}" class="w-full h-full object-cover" controls muted playsinline></video>`
                        : `<img src="${item.url}" class="w-full h-full object-cover" alt="Portfolio media">`;
                    refs.previewCarouselItems.appendChild(itemDiv);
                });

                const hasMultiple = state.mediaPreviews.length > 1;
                toggleClasses(refs.previewCarouselDots, hasMultiple, ['flex'], ['hidden']);
                toggleVisibility(refs.previewCarouselControls, hasMultiple);

                if (hasMultiple) {
                    refs.previewCarouselDots.innerHTML = '';
                    state.mediaPreviews.forEach((item, idx) => {
                        const dotBtn = document.createElement('button');
                        dotBtn.type = 'button';
                        dotBtn.className = 'h-1.5 rounded-full transition-all duration-300 focus:outline-none';
                        dotBtn.addEventListener('click', () => {
                            state.activeIndex = idx;
                            updateUI();
                        });
                        refs.previewCarouselDots.appendChild(dotBtn);
                    });
                }
            } else {
                refs.previewCarouselItems.innerHTML = '';
            }
            lastMediaFingerprint = currentMediaFingerprint;
        }

        // Apply active transforms without recreating DOM
        if (state.mediaPreviews.length > 0) {
            const children = refs.previewCarouselItems.children;
            for (let i = 0; i < children.length; i++) {
                const offset = (i - state.activeIndex) * 100;
                children[i].style.transform = `translateX(${offset}%)`;
            }

            const dots = refs.previewCarouselDots.children;
            for (let i = 0; i < dots.length; i++) {
                dots[i].className =
                    `h-1.5 rounded-full transition-all duration-300 focus:outline-none ${i === state.activeIndex ? 'bg-primary w-4' : 'bg-white/40 w-1.5'}`;
            }
        }

        const valid = state.title.trim() && state.description.trim() && state.mediaPreviews.length > 0;
        refs.submitBtn.disabled = !valid;
        toggleClasses(
            refs.submitBtn,
            valid,
            ['hover:bg-primary/80', 'hover:shadow-[0_0_20px_rgba(255,51,102,0.35)]', 'cursor-pointer'],
            ['opacity-50', 'cursor-not-allowed']
        );

        if (refs.hiddenTags) refs.hiddenTags.value = state.techTags.join(', ');
        if (refs.hiddenMedia && window.DataTransfer) {
            const dt = new DataTransfer();
            state.mediaPreviews.forEach((item) => dt.items.add(item.file));
            refs.hiddenMedia.files = dt.files;
        }
    }

    const teardowns = [];

    const hTitle = (e) => (state.title = e.target.value);
    refs.title.addEventListener('input', hTitle);
    teardowns.push(() => refs.title.removeEventListener('input', hTitle));

    const hDesc = (e) => (state.description = e.target.value);
    refs.desc.addEventListener('input', hDesc);
    teardowns.push(() => refs.desc.removeEventListener('input', hDesc));

    const hMedia = (e) => handleFiles(e.target.files);
    refs.mediaInput.addEventListener('change', hMedia);
    teardowns.push(() => refs.mediaInput.removeEventListener('change', hMedia));

    const hTagInp = (e) => {
        if (e.target.value.includes(',')) {
            addTag(e.target.value);
            e.target.value = '';
        }
    };
    refs.tagInput.addEventListener('input', hTagInp);
    teardowns.push(() => refs.tagInput.removeEventListener('input', hTagInp));

    const hTagKey = (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            if (e.target.value.trim()) {
                addTag(e.target.value);
                e.target.value = '';
            }
        } else if (e.key === 'Backspace' && !e.target.value && state.techTags.length > 0) {
            removeTag(state.techTags.length - 1);
        }
    };
    refs.tagInput.addEventListener('keydown', hTagKey);
    teardowns.push(() => refs.tagInput.removeEventListener('keydown', hTagKey));

    const hTagBlur = (e) => {
        if (e.target.value.trim()) {
            addTag(e.target.value);
            e.target.value = '';
        }
    };
    refs.tagInput.addEventListener('blur', hTagBlur);
    teardowns.push(() => refs.tagInput.removeEventListener('blur', hTagBlur));

    teardowns.push(
        delegate(document.body, 'click', '[data-action]', (e, target) => {
            const action = target.dataset.action;
            const idx = parseInt(target.getAttribute('data-index'));

            switch (action) {
                case 'removeTag':
                    removeTag(idx);
                    break;
                case 'moveUp':
                    moveItem(idx, -1);
                    break;
                case 'moveDown':
                    moveItem(idx, 1);
                    break;
                case 'removeMedia':
                    removeItem(idx);
                    break;
                case 'prevImage':
                    if (state.mediaPreviews.length > 0) {
                        state.activeIndex =
                            (state.activeIndex - 1 + state.mediaPreviews.length) % state.mediaPreviews.length;
                    }
                    break;
                case 'nextImage':
                    if (state.mediaPreviews.length > 0) {
                        state.activeIndex = (state.activeIndex + 1) % state.mediaPreviews.length;
                    }
                    break;
            }
        })
    );

    const hFormSubmit = (e) => {
        if (!state.title.trim() || !state.description.trim() || state.techTags.length === 0) {
            e.preventDefault();
            window.dispatchEvent(
                new CustomEvent('tivri-error', {
                    detail: 'Please fill in all required fields (including tech stack tags).'
                })
            );
        }
    };
    refs.form.addEventListener('submit', hFormSubmit);
    teardowns.push(() => refs.form.removeEventListener('submit', hFormSubmit));

    const hAfterRequest = (e) => {
        if (e.detail.successful) {
            state.mediaPreviews.forEach((item) => URL.revokeObjectURL(item.url));
            state.title = '';
            state.description = '';
            state.techTags = [];
            state.mediaPreviews = [];
            state.activeIndex = 0;
            refs.form.reset();

            document.body.dispatchEvent(new CustomEvent('tivri:portfolio:added', { bubbles: true }));
            window.dispatchEvent(
                new CustomEvent('tivri-success', {
                    detail: { title: 'Portfolio Updated', message: 'Item added successfully!' }
                })
            );
            updateUI();
        } else {
            window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Upload failed' }));
        }
    };
    refs.form.addEventListener('htmx:afterRequest', hAfterRequest);
    teardowns.push(() => refs.form.removeEventListener('htmx:afterRequest', hAfterRequest));

    const hAfterSettle = (e) => {
        if (e.detail.successful) {
            import('../components/portfolio_card.js').then((module) => {
                if (module.initPortfolioCards) module.initPortfolioCards();
            });
        }
    };
    refs.form.addEventListener('htmx:afterSettle', hAfterSettle);
    teardowns.push(() => refs.form.removeEventListener('htmx:afterSettle', hAfterSettle));

    updateUI();

    return () => teardowns.forEach((fn) => fn());
}
