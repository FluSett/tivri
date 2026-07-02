# TIVRI — Strict Coding Rules & Quality Constraints

All rules below are strictly enforced. Any modification to this codebase, regardless of scope, must fully comply with every rule before being committed.

---

## 1. Zero Comments Policy

No comments of any kind are permitted in any source file, template, configuration, Dockerfile, or environment file.

Prohibited formats: `//` (except compiler directives like `//go:embed` and Go toolchain auto-generated markers like `// indirect` in `go.mod`), `/* */`, ``, `#` (outside shell scripts), `--`.

All code must be entirely self-documenting through meaningful naming and logical structure.

---

## 2. No Magic Values

Every environmental variable, credential, port binding, directory path, timeout, and runtime constant must be declared in `internal/config/config.go` and sourced from environment variables.

No hardcoded literal strings, numbers, or paths may appear inside logic blocks, handlers, or SQL queries.

---

## 3. Integer-Only Monetary Values

All monetary amounts (budgets, prices, estimates) are stored and processed as integer cents (e.g. `$50.00` = `5000`). No floating-point arithmetic is permitted for monetary fields anywhere in the codebase.

---

## 4. Exhaustive Error Handling

Every error return value must be checked, handled, and either returned or logged. The blank identifier `_` is never permitted for discarding errors. Database rows, template execution, form parsing, file I/O — all must be audited.

---

## 5. Logical Code Spacing

Go functions and method bodies must include blank lines between distinct logical steps. Tightly packed, run-on blocks of code are not acceptable. Group variable sanitization, input validation blocks, entity allocation, and database mutations into clearly separated blocks.

---

## 6. SOLID Architecture & Domain-Driven Package Boundaries

Strict separation is enforced between delivery layers (`services/web/`, `services/tg-bot/`) and core business logic (`internal/domain/`).

### Package Boundaries:
- `services/web/`: Web application entrypoint, HTTP handlers, templates (`ui/`), and local web assets.
- `services/tg-bot/`: Telegram bot runner, bot command handlers, and localization.
- `internal/app/`: Framework-agnostic bootstrap layer (config loader, logger initialization, DB connection pool setups). Never imports domain layers or handlers.
- `internal/domain/{feature}/`:
  - `model.go`: Domain entities, structs, and dependency-inversion `Repository` interfaces.
  - `service.go`: Transport-agnostic domain business logic. Depends only on domain models/interfaces. Zero awareness of SQL, HTTP, or Telegram.
  - `postgres/`: Concrete repository adapters mapping SQL statements to models.

### Strict Dependency Direction & Anti-Pattern Boundaries:
1. **No Cross-Domain Imports:** Code inside `internal/domain/A` must never import `internal/domain/B`. Cross-feature communication must be handled exclusively by delivery layer handlers passing primitive data or explicit DTOs.
2. **Downward Dependency Flow Only:** Handlers can invoke Services. Services can invoke Repositories. Infrastructure/Repositories must never import Services or Handlers. 
3. **Transport-Agnostic Core:** `internal/domain/` packages must remain 100% free of web primitives (e.g., `http.Request`, status codes) and Telegram event contexts.

---

## 7. i18n Key Conventions

All user-facing strings must have a corresponding key in all three locale files: `en.json`, `uk.json`, `ru.json`.

Key names use PascalCase and follow this pattern: `{Context}{Type}` (e.g. `CompanyNameLabel`, `PlaceholderScope`, `ContactStepTitle`).

No hardcoded English strings may appear in templates — all text goes through `{{.T.KeyName}}`.

---

## 8. Frontend Design System

**Palette:**
- Canvas: `#0A0A0A`
- Primary accent: `#FF3366`
- Text primary: `#F3F4F6`
- Text muted: `text-neutral-400` / `text-neutral-500`

**Glassmorphism elevations:** `border border-white/[0.08] backdrop-blur-md bg-white/[0.02]`

**Typography:** Inter font. Always use `tracking-tight` or `tracking-tighter` on headings. Labels: `text-xs font-bold uppercase tracking-wider`.

**Buttons:**
- Primary CTA: `bg-[#FF3366] text-white hover:bg-[#FF3366]/80`
- Secondary: `border border-[#FF3366] hover:bg-[#FF3366] hover:text-white`
- Ghost: `border border-white/20 hover:bg-white/5`

**Forms:** Single-pixel borders `border-white/[0.08]`, focus state `focus:border-[#FF3366]`, background `bg-[#0A0A0A]`, no outline on focus (`focus:outline-none`), `rounded-lg`, `py-3`.

**Animations:** Use CSS `@keyframes` for decorative animations. Use Tailwind transitions `transition-all duration-300` for interactive state changes. No JavaScript-driven animation for simple hover/focus states.

**Active nav indicator:** The `nav-active` class + `.nav-active-dot` injected via `IntersectionObserver`. No Alpine state required for scroll-based nav logic.

**Layout grid lines:** Fixed vertical `w-px h-full bg-white/[0.03]` lines in a `pointer-events-none` overlay — always present.

**Responsiveness:** Desktop grid, collapsible side-drawer menu on mobile. No horizontal overflow. `whitespace-nowrap` on nav links.

**Template Structure:**
- Public pages are placed under `services/web/ui/html/pages/public/` (e.g. `home.html`, `404.html`).
- Administrative and auth pages are placed under `services/web/ui/html/pages/admin/` (e.g. `dashboard.html`, `login.html`, `unauthorized.html`).
- Reusable page sections and partials are placed under `services/web/ui/html/partials/`.
- Static assets (images, logos, backgrounds) are organized cleanly by type under `services/web/ui/static/` (e.g., `static/favicons/`, `static/img/branding/`, `static/img/backgrounds/`).

---

## 9. Intake Form Standards

The intake stepper must always include these steps in order:
1. Identity (name / company)
2. Project scope (description with char counter)
3. Budget (tile selection + optional custom amount)
4. Contact info (email required, phone optional)

Each step must have: server-side validation mirroring client-side constraints, minlength/maxlength attributes on inputs, and char counters where applicable.

Budget values are converted to and evaluated as integer cents in `intake_leads.budget`. The "other" budget path reads `custom_budget` from the form, not `budget`.

---

## 10. Docker & Deployment Rules

**Database:**
- Development (non-Docker): SQLite via `glebarez/go-sqlite`. Driver selected by `APP_ENV != production`.
- Production (Docker): PostgreSQL 16 via `jackc/pgx/v5`. `APP_ENV=production` activates pgx driver.
- The `db` service must declare a `healthcheck` and `web` must declare `depends_on: db: condition: service_healthy`.

**Dockerfile:**
- Use `alpine:3.19` (not `scratch`) as the final stage — required for DNS resolution and CA certificates with pgx.
- Install `ca-certificates` and `tzdata` in the final stage.
- Use `CGO_ENABLED=0` and `-ldflags="-s -w"` for the build stage.
- Copy only `main`, `ui/`, and `locales/` into the final image — never copy `.env` or development artifacts.

**Volumes:**
- Only named volumes for persistent data (e.g. `tivri_pgdata`) are permitted.
- No host-path bind mounts in the `web` service (no `./ui:/ui`, no `./locales:/locales`). All assets are baked into the image at build time.
- No path mounts from developer workstation directories (e.g. no `C:\Users\...:/artifact`).

**Environment:**
- The `.env` file is for local development defaults only.
- Docker Compose interpolates `${VAR}` from `.env` — never hardcode credentials in `docker-compose.yml`.
- `.env` must be listed in `.gitignore`.

---

## 11. No Dead Code or Placeholder Files

Delete files that serve no production purpose: `ui/static/placeholder.txt`, unused imports, commented-out code blocks, startup hacks that read from non-reproducible file paths.

Every file in the repository must be directly traceable to a production function.

---

## 12. Static Embedding & Asset Resolution

All UI templates and local JSON assets for the web service are baked directly into the web executable via Go's native `//go:embed` directive. 
- Real-time page template modifications during development use template parsing, while production compiles assets statically inside the single executable binary.
- Embedded resources must resolve using sub-file-systems via `fs.Sub(embedFS, "target")`.

---

## 13. Modularity & Clean Design Principles (SOLID, DRY, KISS, YAGNI)

To maintain a highly maintainable, cohesive, and decoupled codebase, all components must adhere to strict software design principles:
- **SOLID**: Every module, package, and file must have a single responsibility (SRP). Interface boundaries must be kept small and client-focused (ISP).
- **DRY & KISS**: Avoid duplicate logic, but do not over-engineer. The simplest solution that works is preferred.
- **YAGNI**: Do not write code or create abstractions for future needs. No placeholder folders, unused struct fields, or inactive code routes are permitted.
- **Direct Domain Flow (No Usecases Layer)**: Business logic flows directly from delivery handlers (`services/web/` or `services/tg-bot/`) into transport-agnostic domain services (`internal/domain/{feature}/service.go`). The introduction of an intermediate `usecase` or `application` layer (and packages/files named `usecase` or `usecase.go`) is strictly prohibited.
- **Decomposition**: Source files and templates must not become monolithic. Complex forms, page sections, or multi-step flows must be decomposed into cohesive partials and referenced explicitly.

---

## 14. Safety, Security, & Resource Management Guidelines

To ensure code safety, reliability, and robust performance, developers and assistants must adhere to the following:
- **Secure Sessions**: Session tokens must be cryptographically secure random identifiers (e.g. generated via `crypto/rand` or strong UUIDs). Deterministic or predictable session identifiers (e.g., static hashing of static credentials) are strictly prohibited.
- **Hardened Cookies**: All HTTP cookies must specify appropriate security flags (`HttpOnly`, `Secure` in production environments, and `SameSite` flags such as `Lax` or `Strict` to mitigate CSRF).
- **Input Validation & Sanitization**: Rely on context-aware escaping templates or proven HTML sanitization libraries rather than naive blocklists or simple string checks (like `strings.Contains`) for security validation.
- **Resource Management**: Configure explicit limits and timeouts on all network connection pools, file descriptors, and database connections (e.g., `SetMaxOpenConns`, `SetMaxIdleConns`, and lifetime limits).
- **Memory Leak Mitigation**: Any in-memory tracking structures (e.g., brute-force lockout maps) must implement a cleanup mechanism (such as background eviction, TTL, or an LRU cache) to prevent unbound memory consumption.

---

## 15. Collaborative Pre-Implementation Design

Before implementing any feature, refactoring, or bug fix, the developer/assistant must:
1. Discuss the requirements and implementation alternatives with the user to align on intent, constraints, and architecture.
2. Outline proposed changes and secure mutual agreement on the design before modifying code or running actions.
3. Validate that the proposed changes are secure, robust, and align with all code safety and structural rules defined in this document.