class Router {
    constructor() {
        this.routes = [];
        this.activeTeardowns = [];
        this.currentPath = null;
        this.initialized = false;

        this.routeTimeout = null;
        this.majorSwapPending = false;

        document.addEventListener('htmx:beforeSwap', (e) => {
            const targetEl = e.detail.target || e.target;
            const targetId = targetEl ? targetEl.id : null;
            
            const isTargetedId = targetId && ['app-body', 'messages-container', 'leads-container'].includes(targetId);
            const isBody = targetEl === document.body;
            
            if (isTargetedId || isBody) {
                this.majorSwapPending = true;
                this.teardown();
            }
        });

        document.addEventListener('htmx:load', (e) => {
            if (this.majorSwapPending) {
                this.majorSwapPending = false;
                clearTimeout(this.routeTimeout);
                this.routeTimeout = setTimeout(() => this.handleRoute(), 50);
            }
        });

        document.addEventListener('htmx:historyRestore', () => {
            clearTimeout(this.routeTimeout);
            this.routeTimeout = setTimeout(() => this.handleRoute(), 50);
        });

        document.addEventListener('DOMContentLoaded', () => {
            if (!this.initialized) {
                this.handleRoute();
            }
        });
    }

    /**
     * Register a route handler.
     * @param {string|RegExp} path - The path to match.
     * @param {Function} handler - The function to run when the route matches. Can optionally return a teardown function.
     */
    on(path, handler) {
        this.routes.push({ path, handler });
        return this; // For chaining
    }

    matchPath(routePath, currentPath) {
        if (typeof routePath === 'string') {
            return routePath === currentPath || (routePath === '/' && currentPath === '');
        }
        if (routePath instanceof RegExp) {
            return routePath.test(currentPath);
        }
        return false;
    }

    handleRoute() {
        this.initialized = true;
        const newPath = window.location.pathname;

        this.teardown();
        this.currentPath = newPath;

        for (const route of this.routes) {
            if (this.matchPath(route.path, this.currentPath)) {
                try {
                    const teardown = route.handler();
                    if (typeof teardown === 'function') {
                        this.activeTeardowns.push(teardown);
                    }
                } catch (e) {
                    console.error(`Router error on path ${this.currentPath}:`, e);
                }
            }
        }
    }

    teardown() {
        while (this.activeTeardowns.length > 0) {
            const teardown = this.activeTeardowns.pop();
            try {
                teardown();
            } catch (e) {
                console.error('Router teardown error:', e);
            }
        }
    }
}

export const router = new Router();
