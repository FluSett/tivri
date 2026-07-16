# TIVRI — High-Performance Software Platform

TIVRI is a high-performance client intake and portfolio management system designed to showcase modern, lightweight web architectures. Built using **Go**, **PostgreSQL**, **HTMX**, and **Alpine.js**, it demonstrates how to build responsive, interactive interfaces without the overhead of heavy Single Page Application (SPA) frameworks.

---

## ✨ Features

- **Multi-Step Client Intake Form**: A multi-step stepper wizard with real-time input validation, responsive budget options, and asynchronous event triggers.
- **Interactive Portfolio**: Showcase agency projects and case studies with dynamic content filtering and media grids.
- **Secure Admin Panel**: Administrative dashboard to view client leads, manage incoming contact queries, update client statuses, and configure system maintenance and queue modes.
- **Universal Multi-Locale Support**: Context-aware localization using query parameters (`?lang=`), cookies, and Accept-Language header fallbacks.

---

## 🛠️ Technology Decisions

- **Go 1.22+**: Used for backend API execution, routing, and template compiling, providing extreme runtime performance and a small memory footprint.
- **PostgreSQL (`pgxpool`)**: Eliminates ORM query translation layers, allowing absolute control over execution plans, indexes, and database transactions.
- **HTMX**: Handles partial HTML replacements over the wire, providing a smooth, single-page application user experience without client-side bundle compilation.
- **Alpine.js**: Manages lightweight, localized client-side states (such as form steps, menus, and dropdowns) natively.
- **Tailwind CSS v4**: Utility styling compiled dynamically using the new CLI engine. Custom styling is modularized (base, components, utilities) and compiled into a single `theme.css`.
- **ES Modules**: JavaScript is broken into feature-specific components (`core/`) and bundled via `esbuild` for optimal client execution without heavy monolithic files.

---

## 🏗️ Architectural Highlights

- **Event-Driven Monolith**: Encapsulated modules communicate asynchronously via an in-memory event bus, decoupling critical HTTP pipelines from background jobs.
- **Transactional Outbox**: Writes event states (e.g. notifications) directly to the database in the same transaction as state updates, resolving distributed transaction sync issues.
- **Self-Documenting Code**: Built around explicit errors handling, descriptive naming, and separation of concerns rather than verbose markup comments.
- **Timing Attack Mitigation**: Verifies admin panel login attempts in constant-time using cryptographic SHA-256 hashes and standard constant-time comparisons.
- **IP Lockout Limits**: Dynamic session tracking that issues temporary bans to client addresses following consecutive failed authentication attempts.

---

## 🚀 Running Locally

### Start using Docker
```bash
docker compose up --build -d
```
- **Platform URL**: `http://localhost:8080`
- **Admin Login**: `http://localhost:8080/admin` (Default: `admin` / `password`)

### Run Tests
```bash
# Run unit tests
go test -v ./...

# Run integration tests (requires setting DB_DSN)
$env:DB_DSN="postgres://postgres:postgres@localhost:5432/tivri?sslmode=disable"
go test -v ./...
```
