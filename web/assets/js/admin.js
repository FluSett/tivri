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
        title: '',
        description: '',
        techStack: '',
        mediaPreviews: [],
        handleFiles(el) {
            this.mediaPreviews = [];
            const files = el.files;
            for (let i = 0; i < files.length; i++) {
                if (files[i].size > 5 * 1024 * 1024) {
                    alert('File ' + files[i].name + ' exceeds maximum size of 5MB.');
                    el.value = '';
                    this.mediaPreviews = [];
                    return;
                }
                this.mediaPreviews.push({
                    url: URL.createObjectURL(files[i]),
                    isVideo: files[i].type.startsWith('video/')
                });
            }
        }
    }));
});
