# TIVRI

A high-performance client intake and portfolio management platform demonstrating modern, lightweight web architecture. Built with **Go**, **PostgreSQL**, **HTMX**, and **Alpine.js**, TIVRI delivers a responsive SPA experience without the heavy JavaScript framework bloat.

## 🚀 Core Features

- **Dynamic Intake Wizard**: Multi-step client onboarding with real-time validation and asynchronous event triggers.
- **Portfolio Management**: Interactive project showcase featuring dynamic filtering and media grids.
- **Secure Admin Dashboard**: Centralized management for leads, queries, and system configuration.
- **Context-Aware Localization**: Seamless multi-locale support via query parameters, cookies, and headers.

## 🛠️ Tech Stack & Rationale

We bypassed complex SPA frameworks (React/Vue/Solid.js) in favor of a lean, server-driven approach. While tools like Solid.js offer excellent client-side performance and compiled DOM updates, they force you to build a separate JSON API and maintain state in two places. Our HTMX + Alpine stack allows us to render HTML directly from Go, keeping a **single source of truth** on the server. This dramatically reduces complexity, eliminates the need for an external Node.js SSR server, and still delivers a highly interactive SPA feel.

- **Go (1.26+)**: Blazing-fast backend, native concurrency, and a single compiled binary footprint.
- **PostgreSQL (`pgxpool`)**: Direct query execution without ORM overhead for maximum transaction control.
- **HTMX**: Wire-delivered HTML partials. We send HTML over the wire instead of JSON, eliminating heavy client-side rendering logic.
- **Alpine.js (Strict CSP)**: Sprinkles lightweight interactivity directly onto our Go templates without a massive Virtual DOM. It is configured to run under strict Content Security Policies, proving you don't need a heavy compiled framework like Solid.js just to build a secure, XSS-resistant frontend.
- **Tailwind CSS v4 & ESBuild**: Dynamic utility styling and modular JS bundled into minimal, highly-optimized assets.

## 🚢 Infrastructure & Deployment

Designed for cost-efficiency and atomic, reproducible deployments.

- **Automated CI/CD**: Pushes to production automatically compile Go binaries, bundle assets, and deploy a lean Docker container.
- **DigitalOcean Infrastructure**: Optimized to comfortably serve high-traffic loads on entry-level Droplets.
- **Cloudflare DNS & Security**: Cloudflare manages our DNS, proxies traffic to obscure origin IPs, and integrates Cloudflare Turnstile to block automated bot submissions. We also utilize privacy-first Cloudflare Web Analytics with a dynamic, user-consented cookie banner for compliance.
- **Custom Domain & Automated SSL**: Nginx acts as our reverse proxy, terminating TLS connections. We use Certbot (Let's Encrypt) to automatically provision and renew SSL certificates for our custom domain.
- **Custom Email Domain**: Platform notifications and client communications are securely routed using SMTP configured for our custom agency domain, ensuring high deliverability.
- **Dockerized Environment**: The application and PostgreSQL database run in isolated, easily reproducible containers. Our optimized production Docker image is incredibly minimalist—weighing in at just **~22MB**. It contains only the standalone Go binary and static assets, completely bypassing heavy OS base images or Node.js runtimes.

## 🏗️ Architectural Highlights

- **Event-Driven Monolith**: Asynchronous in-memory event bus decoupling HTTP pipelines from background tasks.
- **Transactional Outbox**: Guaranteed event delivery by writing state updates alongside events in single database transactions.
- **Security-First**: Constant-time cryptographic verification (SHA-256) and dynamic IP-based rate limiting to prevent timing attacks and brute forcing.
- **Self-Documenting Code**: Clean, explicit error handling and logical separation of concerns.

## 💻 Local Development

### Quick Start (Docker)
```bash
docker compose up --build -d
```
- **App**: `http://localhost:8080`
- **Admin**: `http://localhost:8080/admin` (Credentials set via `.env` file)

### Testing
```bash
# Unit Tests
go test -v ./...

# Integration Tests
$env:DB_DSN="postgres://postgres:postgres@localhost:5432/tivri?sslmode=disable"
go test -v ./...
```
