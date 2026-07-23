# TIVRI

High-performance client intake and portfolio management platform. Built with **Go**, **PostgreSQL**, **HTMX**, and **Vanilla JS** for a lightweight SPA experience without JavaScript framework bloat.

## 🚀 Key Features

- **Dynamic Intake Wizard**: Interactive 6-step client request wizard for micro-tasks (<$100), bug fixes, API integrations, code audits, and full custom software projects with async event processing.
- **Portfolio Showcase**: Filterable project showcase and dynamic media grids with server-side rendered tech stack badges (`splitTags`).
- **Real-Time Admin Dashboard**: Lead management, inquiry tracking, and instant system configuration toggles (**High Queue** alert banner & **Maintenance Mode**).
- **Multi-Locale (i18n)**: Seamless translation resolution across English (`en`), Ukrainian (`uk`), and Russian (`ru`) via query parameters, cookies, or `Accept-Language` headers.
- **Modern Status Pages**: Glassmorphic centered 404 & maintenance pages with animated pulse rings and smooth typography.

## 🛠️ Tech Stack

- **Backend**: Go 1.26+, `net/http`, `log/slog` structured logging, `pgxpool` PostgreSQL driver.
- **Frontend**: HTMX server-driven HTML partials, Vanilla JS ES modules, Tailwind CSS v4 (processed via Bun & esbuild).
- **Security**: PostgreSQL Row-Level Security (RLS), Nginx CSP headers, Gorilla CSRF (`X-CSRF-Token`), Cloudflare Turnstile bot protection.
- **Deployment**: Minimal ~22MB Docker image (`scratch` runtime).

## 🏗️ Architecture

- **Event-Driven Monolith**: Async in-memory event bus decoupling HTTP handlers from background notification workers.
- **Transactional Outbox**: Reliable event dispatching using PostgreSQL outbox pattern and `context.WithTimeout`.
- **PostgreSQL Row-Level Security (RLS)**: Enforced table-level isolation (`FORCE ROW LEVEL SECURITY`) with `(SELECT current_setting('app.current_role', true))` `InitPlan` query optimization for high-concurrency throughput.
- **Real-Time Database Settings**: Instant `system_settings` UPSERT propagation for system switches (`high_queue` and `maintenance_mode`) without requiring server restarts.
- **Resilient Media Pipeline**: Multi-file upload handling supporting up to 50MB form payloads, dual MIME/extension verification, and WebP encoding with automatic original stream copy fallback.
- **High-Load Trigram Search**: Sub-millisecond text search via PostgreSQL `pg_trgm` GIN indexes (`idx_intake_leads_company_trgm`).
- **Single-Pass Window Pagination**: High-performance lead and message listing using `COUNT(*) OVER()` window functions to execute pagination in a single DB roundtrip.
- **Transaction-Scoped Role Isolation**: Dynamic database role scoping using `SET LOCAL app.current_role` to prevent connection pool leakage.
- **Sub-Domain Handlers**: Modular handler package structure (`admin_leads.go`, `admin_messages.go`, `admin_portfolio.go`, `admin_settings.go`).
- **Immutable Caching**: Cache-busted static assets via SHA-256 content hashing (`?v=hash`).

## 💻 Local Development

### Run Container
```bash
docker compose up --build -d
```
- App: `http://localhost:8080`
- Admin: `http://localhost:8080/admin`

### Run Tests
```bash
go test -v ./...
```

## 📖 API Specification

The public API specification (OpenAPI 3.0) is maintained in [`docs/openapi.yaml`](file:///c:/core/main/tivri/tivri/docs/openapi.yaml). It documents all public endpoints (`/healthz`, `/api/intake`, `/api/contact`), form parameters, and response schemas.

### Viewing & Testing the API Spec
- **IDE Extension**: Use the **OpenAPI Preview** or **Swagger Viewer** extension in VS Code.
- **Online Editor**: Import [`docs/openapi.yaml`](file:///c:/core/main/tivri/tivri/docs/openapi.yaml) into [Swagger Editor](https://editor.swagger.io/).
- **Redocly CLI**: Run `npx @redocly/cli preview-docs docs/openapi.yaml` to launch an interactive docs server.
