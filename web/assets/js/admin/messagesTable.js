export function messagesTable() {
    return {
        statusFilter: Alpine.$persist('all').as('adm_m_statusFilter'),
        sortBy: Alpine.$persist('date_desc').as('adm_m_sortBy'),
        filterOpen: false,
        messages: [],
        init() {
            const dataStr = this.$el.getAttribute('data-messages');
            if (dataStr) {
                this.messages = JSON.parse(dataStr);
            }
            window.tivriHandleLocaleChange(() => {
                this.statusFilter = 'all';
                this.sortBy = 'date_desc';
            });
        },
        toggleFilter() {
            this.filterOpen = !this.filterOpen;
        },
        closeFilter() {
            this.filterOpen = false;
        },
        getFilterDropdownClass() {
            return this.filterOpen
                ? 'opacity-100 scale-100 pointer-events-auto'
                : 'opacity-0 scale-95 pointer-events-none delay-200';
        },
        setStatusFilter(val) {
            this.statusFilter = val;
        },
        setSortBy(val) {
            this.sortBy = val;
        },
        getSortByText() {
            const map = { date_desc: 'Newest', date_asc: 'Oldest', email_asc: 'Email A-Z', email_desc: 'Email Z-A' };
            return map[this.sortBy] || '';
        },
        getStatusClass(status) {
            if (status === 'new') return 'text-blue-400 border-blue-500/20 bg-blue-500/5';
            if (status === 'answered') return 'text-yellow-400 border-yellow-500/20 bg-yellow-500/5';
            if (status === 'done') return 'text-green-400 border-green-500/20 bg-green-500/5';
            return '';
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
    };
}
