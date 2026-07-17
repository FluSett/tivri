export function leadsTable() {
    return {
        clientFilter: Alpine.$persist('all').as('adm_l_clientFilter'),
        internalFilter: Alpine.$persist('all').as('adm_l_internalFilter'),
        sortBy: Alpine.$persist('date_desc').as('adm_l_sortBy'),
        filterOpen: false,
        leads: [],
        init() {
            const dataStr = this.$el.getAttribute('data-leads');
            if (dataStr) {
                this.leads = JSON.parse(dataStr);
            }
            window.tivriHandleLocaleChange(() => {
                this.clientFilter = 'all';
                this.internalFilter = 'all';
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
        setClientFilter(val) {
            this.clientFilter = val;
        },
        setInternalFilter(val) {
            this.internalFilter = val;
        },
        setSortBy(val) {
            this.sortBy = val;
        },
        getSortByText() {
            const map = {
                date_desc: 'Newest',
                date_asc: 'Oldest',
                budget_desc: 'Highest Budget',
                budget_asc: 'Lowest Budget',
                name_asc: 'Name A-Z',
                name_desc: 'Name Z-A'
            };
            return map[this.sortBy] || '';
        },
        getStatusClass(status) {
            if (status === 'pending') return 'text-blue-400 border-blue-500/20 bg-blue-500/5';
            if (status === 'active') return 'text-yellow-400 border-yellow-500/20 bg-yellow-500/5';
            if (status === 'done') return 'text-green-400 border-green-500/20 bg-green-500/5';
            if (status === 'canceled') return 'text-red-400 border-red-500/20 bg-red-500/5';
            return '';
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
    };
}
