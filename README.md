# TIVRI - High-Performance Software Platform

TIVRI is a premium Go, HTMX, and Alpine.js onboarding and portfolio system designed for modern web systems.

## 🏗️ Architecture: Event-Driven Monolith
* **Feature-Layered Monolith**: Encapsulated feature domains in `internal/features/`.
* **In-Memory Event Bus**: Asynchronous communication between domains via channels.

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
Deploys to DigitalOcean Droplets on every commit push to `main` branch.
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
