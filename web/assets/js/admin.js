document.addEventListener('alpine:init', () => {
    Alpine.data('loginForm', () => ({
        username: Alpine.$persist('').as('adm_log_user'), 
        password: Alpine.$persist('').as('adm_log_pass'),
        init() {
            window.tivriHandleLocaleChange(() => {
                this.username = '';
                this.password = '';
            });
        }
    }));

    Alpine.data('leadsTable', (initialLeads) => ({
        clientFilter: Alpine.$persist('all').as('adm_l_clientFilter'),
        internalFilter: Alpine.$persist('all').as('adm_l_internalFilter'),
        sortBy: Alpine.$persist('date_desc').as('adm_l_sortBy'),
        leads: initialLeads || [],
        init() {
            window.tivriHandleLocaleChange(() => {
                this.clientFilter = 'all';
                this.internalFilter = 'all';
                this.sortBy = 'date_desc';
            });
        },
        get filteredLeads() {
            return this.leads
                .filter((l) => {
                    if (this.clientFilter !== 'all' && l.clientStatus !== this.clientFilter) return false;
                    if (this.internalFilter !== 'all' && l.internalStatus !== this.internalFilter) return false;
                    return true;
                })
                .sort((a, b) => {
                    if (this.sortBy === 'date_asc') return a.createdAt - b.createdAt;
                    if (this.sortBy === 'date_desc') return b.createdAt - a.createdAt;
                    if (this.sortBy === 'budget_asc') return a.budget - b.budget;
                    if (this.sortBy === 'budget_desc') return b.budget - a.budget;
                    if (this.sortBy === 'name_asc') return a.companyName.localeCompare(b.companyName);
                    if (this.sortBy === 'name_desc') return b.companyName.localeCompare(a.companyName);
                    return 0;
                });
        },
        updateStatus(lead, type, status) {
            fetch('/admin/leads/status', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: new URLSearchParams({ id: lead.id, type: type, status: status })
            });
            if (type === 'client') {
                lead.clientStatus = status;
            } else {
                lead.internalStatus = status;
            }
            const now = new Date();
            lead.updatedAt = Math.floor(now.getTime() / 1000);
            const pad = (n) => String(n).padStart(2, '0');
            lead.updatedAtStr = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}`;
        },
        formatBudgetTier(cents, isCustom) {
            if (isCustom) {
                return 'Custom ($' + (cents / 100).toLocaleString() + ')';
            }
            switch (cents) {
                case 250000:
                    return 'Small ($1k–$5k)';
                case 750000:
                    return 'Starter ($5k–$10k)';
                case 1750000:
                    return 'Growth ($10k–$25k)';
                case 5000000:
                    return 'Scale ($25k–$75k)';
                case 11250000:
                    return 'Enterprise ($75k–$150k)';
                case 20000000:
                    return 'Premium ($150k+)';
                default:
                    return 'Custom ($' + (cents / 100).toLocaleString() + ')';
            }
        }
    }));

    Alpine.data('messagesTable', (initialMessages) => ({
        statusFilter: Alpine.$persist('all').as('adm_m_statusFilter'),
        sortBy: Alpine.$persist('date_desc').as('adm_m_sortBy'),
        messages: initialMessages || [],
        init() {
            window.tivriHandleLocaleChange(() => {
                this.statusFilter = 'all';
                this.sortBy = 'date_desc';
            });
        },
        get filteredMessages() {
            return this.messages
                .filter((m) => {
                    if (this.statusFilter !== 'all' && m.status !== this.statusFilter) return false;
                    return true;
                })
                .sort((a, b) => {
                    if (this.sortBy === 'date_asc') return a.createdAt - b.createdAt;
                    if (this.sortBy === 'date_desc') return b.createdAt - a.createdAt;
                    if (this.sortBy === 'email_asc') return a.email.localeCompare(b.email);
                    if (this.sortBy === 'email_desc') return b.email.localeCompare(a.email);
                    return 0;
                });
        },
        updateStatus(msg, status) {
            fetch('/admin/messages/status', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: new URLSearchParams({ id: msg.id, status: status })
            });
            msg.status = status;
            const now = new Date();
            msg.updatedAt = Math.floor(now.getTime() / 1000);
            const pad = (n) => String(n).padStart(2, '0');
            msg.updatedAtStr = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}`;
        }
    }));

    Alpine.data('portfolioForm', () => ({
        title: Alpine.$persist('').as('adm_pf_title'),
        description: Alpine.$persist('').as('adm_pf_description'),
        techTags: Alpine.$persist([]).as('adm_pf_techtags'),
        mediaPreviews: [],
        init() {
            window.tivriHandleLocaleChange(() => {
                this.title = '';
                this.description = '';
                this.techTags = [];
            });
        },
        addTag(val) {
            val.split(',').forEach(t => {
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
    }));
});
