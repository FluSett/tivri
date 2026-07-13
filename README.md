# TIVRI - High-Performance Software Platform

TIVRI is a premium Go, HTMX, and Alpine.js onboarding and portfolio system designed for modern web systems.

## 🏗️ Architecture: Event-Driven Monolith
* **Feature-Layered Monolith**: Encapsulated feature domains in `internal/features/`.
* **In-Memory Event Bus**: Asynchronous communication between domains via channels and a worker thread pool.
* **Banned Patterns**: Distributed microservices and structured use-case layers are strictly banned.

## 📁 Repository Layout
* `cmd/api/main.go` — Entrypoint.
* `internal/app/` — HTTP server configuration and routing assembly.
* `internal/config/` — Configurations loaded from environment variables/fallback `.env`.
* `internal/core/` — Primitives for database (`pgxpool`) and session/lockout security.
* `internal/eventbus/` — Asynchronous broker and thread workers.
* `internal/i18n/` — Translation handlers.
* `internal/features/` — Encapsulated domains (`project_intake`, `messaging`, `portfolio`, `notifications`).
* `locales/` — Localization tables (`en`, `uk`, `ru`).
* `nginx/tivri.conf` — Proxy cache rules, rate-limits, and CSP policies.
* `scripts/` — Automated pg_dump backups and cron server health alerts.
* `web/assets/` — CSS design tokens (`input.css`), compiled theme (`theme.css`), and components.
* `web/templates/` — Renders and layout files.

## 🛠️ Tech Stack
* **Backend**: Go (Golang), `pgxpool` direct query access (no ORMs).
* **Frontend**: HTML templates, Alpine.js (UI state wizard), HTMX (outerHTML locale swaps).
* **Styling**: Tailwind CSS v4.

## ⚙️ Technical & Coding Standards

### Golang & Database Standards
* **Context Threading**: Pass down `context.Context` explicitly to all database, transaction, and client requests. Use decoupled contexts with timeout for background work.
* **Direct Access**: Use `jackc/pgx/v5/pgxpool` directly. Standard `database/sql` or ORMs are banned.
* **Error Hygiene**: Wrap errors explicitly (`fmt.Errorf("layer/component: operation failed: %w", err)`). Never discard errors.
* **Concurrency**: Dangling goroutines (`go func()`) are banned. Use lifecycle-managed named goroutines or the `MemoryEventBus`.

### Alpine.js, HTMX & Styling
* **No Build Steps**: Native Go `html/template` rendering backed by Alpine.js declarative bindings.
* **No Inline Scripts/Styles**: Inline `<script>` tags inside templates and inline style attributes are banned. Use `Alpine.data()` and CSS classes.
* **State Preservation**: Preserve HTMX body swaps state via the `tivri_htmx_nav` sessionStorage flag handled by `app.js`.
* **CSS Dryness**: Extract repeated styles (2+ times) into semantic rules in `web/assets/css/theme.css`.

### Security Guardrails
* **Financial Primitives**: Do not use floats for currency. Use integer subunits (cents) or arbitrary-precision libraries.
* **SQL Injection & XSS**: Always use parameterized queries (`$1`, `$2`). Encoded dynamic data passed to Alpine.js using JSON serialization.
* **Sessions & Transactions**: Access tokens stored in `HTTP-Only`, `Secure`, `SameSite=Strict` cookies. Multi-table modifications must run in transaction blocks (`pgx.Tx`) with explicit rollback.

### Infrastructure & i18n
* **Nginx Policies**: Static assets bypass Go app via `/assets/` aliases. Strict security headers (CSP, Frame options) and rate-limiting zones on form endpoints.
* **i18n**: Resolve language in order: URL subroute (`/en/`) -> Session Cookie -> `Accept-Language` header.

### Parity & Inspection
* **Outbox & State Recovery**: Ensure database and event bus consistency asynchronously. Retain onboarding state across refreshes.
* **Feature Parity**: Admin configs must dynamically update client views. Keep variable and database names matching identical keys.

## 🌐 Cloud Infrastructure & Integrations
* **Cloudflare (DNS, SSL, & Security)**:
  * DNS proxying masks the hosting server's origin IP to prevent direct DDoS attacks.
  * SSL/TLS encryption is terminated using Cloudflare Origin Certificates (staged inside the container).
  * Cloudflare Turnstile blocks automated bot spam on client intake and contact forms.
* **DigitalOcean (Hosting & Environment)**:
  * Application runs on a DigitalOcean Droplet managed via Docker Compose.
  * Nginx acts as a reverse proxy, enforcing rate-limiting zones, caching assets, and injecting strict security headers (CSP, XSS protection).
* **Telegram Notifications (Asynchronous Alerting)**:
  * Uses Telegram Bot API to notify administrators of incoming leads, messages, and login attempts.
  * A host-level cron job (`scripts/health_check.sh`) pings the site every 5 minutes, raising critical alerts to Telegram if the site fails to respond or returns 5xx/503 errors.

## ⚙️ Environment Variables (.env)
```env
APP_ENV=development             # development / production
PORT=8080
DB_DSN=tivri.db
LOCALES_DIR=locales
POSTGRES_USER=tivri
POSTGRES_PASSWORD=secret
POSTGRES_DB=tivri
ADMIN_USERNAME=admin
ADMIN_PASSWORD=secureadminpass
TURNSTILE_SITE_KEY=xxxxxx       # Cloudflare TurnstileSiteKey
TURNSTILE_SECRET_KEY=xxxxxx     # Cloudflare TurnstileSecretKey
TELEGRAM_BOT_TOKEN=xxxxxx
TELEGRAM_CHAT_ID=xxxxxx
```

## 🚀 Local Running
1. Run Postgres database and web app:
   ```bash
   docker compose up --build
   ```
2. Compile Assets locally (only needed if running the Go app directly without Docker):
   ```bash
   npm install
   npm run build
   ```
3. Endpoints: Public: `http://localhost:8080` | Console: `http://localhost:8080/admin`

## ✈️ Automated Deploy & Uptime Monitoring (CI/CD)
Deploys to DigitalOcean Droplets.
* **Automated Asset Generation**: Assets (Tailwind CSS compilation & JS minification) are built dynamically in Node.js builder stages inside Docker, keeping Git clean.
* **Filter copying**: The `scripts/` folder is conditionally uploaded only if files are changed.
* **Cron Auto-Config**: The pipeline automatically configures and schedules host crontab tasks utilizing `.env` secrets:
  ```bash
  */5 * * * * /var/www/tivri/scripts/health_check.sh "https://tivri.cc" "YOUR_TELEGRAM_BOT_TOKEN" "YOUR_TELEGRAM_CHAT_ID" >> /var/log/tivri_health.log 2>&1
  ```

### Diagnostics
If server-side Telegram alerts fail:
1. Verify outbound connectivity:
   ```bash
   curl -I https://api.telegram.org
   ```
2. Check Docker web container logs:
   ```bash
   docker compose logs web | grep notifications/telegram
   ```
