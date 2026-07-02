# PROJECT.md - TIVRI Software Agency Monorepo Blueprint

TIVRI is a premium, high-performance multilingual software engineering agency. The application operates as a local-first, containerized monorepo supporting both a public-facing Web Service (HTTP/HTMX/Alpine.js) and an administrative Telegram Bot microservice.

---

## 🏛️ Architecture & Clean Domain Separation

The codebase follows a Domain-Driven Clean Architecture, isolating delivery mechanisms under `services/` from transport-agnostic business logic under `internal/`.

### 1. Delivery Layers (`services/`)
- **Web App (`services/web/`)**:
  - Exposes the agency home page, multi-step lead intake stepper form, and the administrative dashboard.
  - Dynamically switches locales based on HTTP parameters/headers.
  - Uses Go's native `embed` capabilities (`//go:embed`) to build a single static executable binary containing all assets and templates.
- **Telegram Bot (`services/tg-bot/`)**:
  - Serves as an administrative alerting service and telemetry cockpit.
  - Intercepts webhook triggers or polling streams.
  - Consumes the exact same domain services as the web server, ensuring DRY business integrity.

### 2. Core Business Logic (`internal/domain/`)
- Divided into three distinct domain folders representing the primary agency entities:
  - **`contact`**: Manages direct messages, feedback topics, and validation limits.
  - **`lead`**: Controls the project intake stepper, budgets, scopes, and customer contact pipelines.
  - **`portfolio`**: Serves visual tags, project URLs, multilingual title tracks, and asset layouts.
- Dependency Inversion (DIP) is strictly maintained. The core logic inside `service.go` knows only of interfaces defined in `model.go`. The persistent SQL adapter is abstracted in `postgres/` as a concrete driver layer.

---

## 🌐 Dynamic Trilingual Localization Engine

The agency operates globally, supporting:
- **English (`en`)**
- **Ukrainian (`uk`)**
- **Russian (`ru`)**

All user-facing strings are resolved dynamically from thread-safe memory localization caches loaded during bootstrap.

---

## 🎨 Premium Glassmorphic Design Specifications
- Background Canvas: Deep pitch-black primary canvas (`#0A0A0A`).
- elevations: Translucent card perimeters (`border border-white/[0.08]`) with high-end ambient backdrop blending (`backdrop-blur-md bg-white/[0.02]`).
- Visual Focal Points: Indigo backgrounds (`#4F46E5`) and primary accents (`#FF3366`).
- Font: Inter.