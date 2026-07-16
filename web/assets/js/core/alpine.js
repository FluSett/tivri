export function initAlpine() {
    document.addEventListener('alpine:init', function () {
        Alpine.data('layout', function () {
            return {
                mobileMenuOpen: false,
                openStepper: this.$persist(false).as('openStepper').using(sessionStorage)
            };
        });

        Alpine.data('globalError', function () {
            return {
                errorMessage: '',
                errorTime: '',
                showError: false,
                lastErrorTime: null,
                timeoutId: null,
                init: function () {
                    var self = this;
                    window.addEventListener('tivri-error', function (e) {
                        var now = new Date();
                        
                        // Prevent duplicate error spam when already displaying
                        if (self.showError && self.errorMessage === e.detail) {
                            return;
                        }
                        
                        // Throttle error updates to prevent layout flash spam
                        if (self.lastErrorTime && (now - self.lastErrorTime < 1500)) {
                            return;
                        }
                        
                        self.lastErrorTime = now;
                        var pad = function (num) { return String(num).padStart(2, '0'); };
                        self.errorTime = pad(now.getHours()) + ':' + pad(now.getMinutes()) + ':' + pad(now.getSeconds());
                        self.errorMessage = e.detail;
                        self.showError = true;
                        
                        if (self.timeoutId) {
                            clearTimeout(self.timeoutId);
                        }
                        self.timeoutId = setTimeout(function () {
                            self.showError = false;
                        }, 6000);
                    });
                }
            };
        });

        Alpine.data('dropdown', function () {
            return {
                open: false
            };
        });
    });
}

// Execute immediately when imported
initAlpine();
