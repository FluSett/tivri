# TIVRI - High-Performance Software Platform

Premium Go, HTMX, and Alpine.js onboarding and portfolio system.

## 🏗️ Architecture & Layout
- **Event-Driven Monolith**: Encapsulated domains (`internal/features/`) communicating via in-memory event bus. Microservices are banned.
- `cmd/api/main.go` — Entrypoint.
- `internal/app/` — Routing & HTTP assembly.
- `internal/features/` — Domains (`project_intake`, `messaging`, `portfolio`).
- `web/` — UI templates and compiled assets (Tailwind v4).

## 🛠️ Tech Stack & Standards
- **Backend**: Go 1.25+, `pgxpool` direct query (no ORMs).
- **Frontend**: Go `html/template` (utilizing `dict` helper for reusable components), Alpine.js (utilizing `$persist` for native state), HTMX. No inline scripts/styles.
- **Styling**: Tailwind CSS v4, DRY semantic classes in `theme.css`.

## ⚙️ Coding Guidelines
- **Context & Errors**: Pass `context.Context` everywhere. Wrap errors explicitly; never discard.
- **Concurrency**: Use lifecycle-managed workers or EventBus; dangling `go func()` is banned.
- **Security**: Parameterized queries only. Multi-table mutations must use `pgx.Tx` with rollback. Integer subunits (cents) for currency.
- **i18n**: Subroute (`/en/`) -> Cookie -> `Accept-Language`.

## 🌐 Infrastructure & CI/CD
- **Cloudflare & DigitalOcean**: DNS proxy, Turnstile anti-bot, Docker Compose + Nginx reverse proxy.
- **Deployments**: Assets (CSS/JS) compile dynamically inside Docker builder stages to keep Git clean.
- **Environment**: Global configuration like `APP_URL` (canonical domain) and `CONTACT_EMAIL` ensure zero hardcoded values.

## 🚀 Local Running
```bash
docker compose up --build
```
- Public: `http://localhost:8080`
- Admin: `http://localhost:8080/admin`
