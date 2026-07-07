document.addEventListener('alpine:init', () => {
    Alpine.data('leadsTable', (initialLeads) => ({
        clientFilter: 'all',
        internalFilter: 'all',
        sortBy: 'date_desc',
        leads: initialLeads || [],
        get filteredLeads() {
            return this.leads
                .filter(l => {
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
        }
    }));

    Alpine.data('messagesTable', (initialMessages) => ({
        statusFilter: 'all',
        sortBy: 'date_desc',
        messages: initialMessages || [],
        get filteredMessages() {
            return this.messages
                .filter(m => {
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
        }
    }));

    Alpine.data('portfolioForm', () => ({
        title: sessionStorage.getItem('admin_portfolio_title') || '',
        description: sessionStorage.getItem('admin_portfolio_description') || '',
        techStack: sessionStorage.getItem('admin_portfolio_techStack') || '',
        mediaPreviews: [],
        init() {
            this.$watch('title', val => sessionStorage.setItem('admin_portfolio_title', val || ''));
            this.$watch('description', val => sessionStorage.setItem('admin_portfolio_description', val || ''));
            this.$watch('techStack', val => sessionStorage.setItem('admin_portfolio_techStack', val || ''));
        },
        handleFiles(el) {
            const files = el.files;
            for (let i = 0; i < files.length; i++) {
                const f = files[i];
                if (f.size > 5 * 1024 * 1024) {
                    window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'File ' + f.name + ' exceeds maximum size of 5MB.' }));
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
            this.mediaPreviews.forEach(item => URL.revokeObjectURL(item.url));
            this.title = '';
            this.description = '';
            this.techStack = '';
            this.mediaPreviews = [];
            sessionStorage.removeItem('admin_portfolio_title');
            sessionStorage.removeItem('admin_portfolio_description');
            sessionStorage.removeItem('admin_portfolio_techStack');
        },
        async submitForm(formEl) {
            if (!this.title.trim() || !this.description.trim() || !this.techStack.trim()) {
                window.dispatchEvent(new CustomEvent('tivri-error', { detail: 'Please fill in all required fields.' }));
                return;
            }
            
            const fd = new FormData();
            fd.append('title', this.title);
            fd.append('description', this.description);
            fd.append('tech_stack', this.techStack);
            
            this.mediaPreviews.forEach(item => {
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
                window.dispatchEvent(new CustomEvent('tivri-error', { detail: err.message || 'Network error occurred during portfolio upload.' }));
            }
        }
    }));
});
