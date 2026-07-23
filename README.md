# TIVRI

High-performance client intake and portfolio management platform. Built with **Go**, **PostgreSQL**, **HTMX**, and **Vanilla JS** for a lightweight SPA experience without JavaScript framework bloat.

## 🚀 Key Features

- **Dynamic Intake Wizard**: Interactive 6-step client request wizard for micro-tasks (<$100), bug fixes, API integrations, code audits, and full custom software projects with async event processing.
- **Portfolio Showcase**: Filterable project showcase and dynamic media grids.
- **Admin Dashboard**: Lead management, inquiry tracking, and system configuration.
- **Multi-Locale (i18n)**: Seamless translation resolution via query, cookie, or `Accept-Language`.

## 🛠️ Tech Stack

- **Backend**: Go 1.26+, `net/http`, `log/slog` structured logging, `pgxpool` PostgreSQL driver.
- **Frontend**: HTMX server-driven HTML partials, Vanilla JS ES modules, Tailwind CSS v4 (processed via Bun & esbuild).
- **Security**: Nginx CSP headers, Gorilla CSRF (`X-CSRF-Token`), Cloudflare Turnstile bot protection.
- **Deployment**: Minimal ~22MB Docker image (`scratch` runtime).

## 🏗️ Architecture

- **Event-Driven Monolith**: Async in-memory event bus decoupling HTTP handlers from background tasks.
- **Transactional Outbox**: Reliable event dispatching using PostgreSQL outbox pattern and `context.WithTimeout`.
- **Sub-Domain Handlers**: Modular handler package structure (`admin_leads.go`, `admin_messages.go`, `admin_portfolio.go`).
- **Composite Indexes**: Optimized PostgreSQL composite indexes for sub-millisecond query execution.
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
