export function portfolioForm() {
    return {
        title: Alpine.$persist('').as('adm_pf_title'),
        description: Alpine.$persist('').as('adm_pf_description'),
        techTags: Alpine.$persist([]).as('adm_pf_techtags'),
        mediaPreviews: [],
        tagInput: '',
        activeIndex: 0,
        init() {
            window.tivriHandleLocaleChange(() => {
                this.title = '';
                this.description = '';
                this.techTags = [];
            });
        },
        get hasMediaPreviews() {
            return this.mediaPreviews.length > 0;
        },
        get hasMultipleMediaPreviews() {
            return this.mediaPreviews.length > 1;
        },
        get hasNoTechTags() {
            return this.techTags.length === 0;
        },
        get activePreviewIndex() {
            return this.activeIndex;
        },
        setActiveIndex(index) {
            this.activeIndex = index;
        },
        nextPreview() {
            if (this.mediaPreviews.length > 0) {
                this.activeIndex = (this.activeIndex + 1) % this.mediaPreviews.length;
            }
        },
        prevPreview() {
            if (this.mediaPreviews.length > 0) {
                this.activeIndex = (this.activeIndex - 1 + this.mediaPreviews.length) % this.mediaPreviews.length;
            }
        },
        handleTagInputInput(event) {
            if (this.tagInput.includes(',')) {
                this.addTag(this.tagInput);
                this.tagInput = '';
            }
        },
        handleTagInputEnter(event) {
            if (this.tagInput.trim()) {
                this.addTag(this.tagInput);
                this.tagInput = '';
            }
        },
        handleTagInputBackspace(event) {
            if (!this.tagInput && this.techTags.length > 0) {
                this.removeTag(this.techTags.length - 1);
            }
        },
        handleTagInputBlur(event) {
            if (this.tagInput.trim()) {
                this.addTag(this.tagInput);
                this.tagInput = '';
            }
        },
        getPreviewClass(index) {
            return this.activeIndex === index
                ? 'opacity-100 pointer-events-auto'
                : 'opacity-0 pointer-events-none delay-300';
        },
        addTag(val) {
            val.split(',').forEach((t) => {
                const trimmed = t.trim();
                if (trimmed && !this.techTags.includes(trimmed)) {
                    this.techTags.push(trimmed);
                }
            });
        },
        removeTag(index) {
            this.techTags.splice(index, 1);
        },
        handleFiles(el) {
            const files = el.files;
            for (let i = 0; i < files.length; i++) {
                const f = files[i];
                if (f.size > 5 * 1024 * 1024) {
                    window.dispatchEvent(
                        new CustomEvent('tivri-error', { detail: 'File ' + f.name + ' exceeds maximum size of 5MB.' })
                    );
                    continue;
                }
                this.mediaPreviews.push({
                    id: Math.random().toString(36).substr(2, 9),
                    file: f,
                    name: f.name,
                    url: URL.createObjectURL(f),
                    isVideo: f.type.startsWith('video/')
                });
            }
            el.value = '';
        },
        moveItem(index, direction) {
            const targetIndex = index + direction;
            if (targetIndex < 0 || targetIndex >= this.mediaPreviews.length) return;
            const temp = this.mediaPreviews[index];
            this.mediaPreviews[index] = this.mediaPreviews[targetIndex];
            this.mediaPreviews[targetIndex] = temp;
        },
        removeItem(index) {
            URL.revokeObjectURL(this.mediaPreviews[index].url);
            this.mediaPreviews.splice(index, 1);
        },
        clearForm() {
            this.mediaPreviews.forEach((item) => URL.revokeObjectURL(item.url));
            this.title = '';
            this.description = '';
            this.techTags = [];
            this.mediaPreviews = [];
        },
        async submitForm(formEl) {
            if (!this.title.trim() || !this.description.trim() || this.techTags.length === 0) {
                window.dispatchEvent(
                    new CustomEvent('tivri-error', {
                        detail: 'Please fill in all required fields (including tech stack tags).'
                    })
                );
                return;
            }

            const fd = new FormData();
            fd.append('title', this.title);
            fd.append('description', this.description);
            fd.append('tech_stack', this.techTags.join(', '));
            this.mediaPreviews.forEach((item) => {
                fd.append('media', item.file);
            });

            try {
                const response = await fetch('/admin/portfolio', {
                    method: 'POST',
                    headers: {
                        'HX-Request': 'true',
                        'HX-Trigger': 'portfolio-upload'
                    },
                    body: fd
                });
                if (!response.ok) {
                    const errMsg = await response.text();
                    window.dispatchEvent(new CustomEvent('tivri-error', { detail: errMsg || 'Upload failed' }));
                    return;
                }

                const html = await response.text();
                const grid = document.getElementById('portfolio-grid');
                if (grid) {
                    grid.insertAdjacentHTML('beforeend', html);
                }
                this.clearForm();
                formEl.reset();
            } catch (err) {
                window.dispatchEvent(
                    new CustomEvent('tivri-error', {
                        detail: err.message || 'Network error occurred during portfolio upload.'
                    })
                );
            }
        }
    };
}
