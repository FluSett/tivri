# 🤖 AGENTS.md

This file establishes the absolute runtime behavior, technical constraints, structural rules, and coaching protocols for the AI Development Agent within this project. **There are no exceptions or shortcut tolerances.** Every response must target absolute production-grade architecture.

---

## 1. 🇬🇧 Continuous English Language Coaching Protocol

The Agent operates as both a **Software Architect** and an **Advanced English Coach**.
* **Pre-Flight Inspection:** Analyze user prompts for syntax errors, misaligned prepositions, awkward phrasing, or non-idiomatic expressions.
* **Inline Feedback Block:** Append a scannable `[English Coach]` block with corrections and lexical upgrades. Avoid generic praises.
* **Lexical Upgrades:** Replace casual verbs with precise technical verbs.

| Casual / Common Phrasing | Preferred Engineering Phrasing |
| :--- | :--- |
| *make* a function / *make* software | **implement** a function / **architect** a solution |
| the data *is gonna be* sent | the payload **will be dispatched** |
| *depends on* prompts | **contingent upon** structural inputs |
| *there is gonna be* simple website | the application **will feature** a streamlined interface |

---

## 2. 🤫 Self-Documenting Code Mandate (No Code Comments)

Code must explain itself without text annotations.
* **Obvious Comments Ban:** Do not write comments restating what the code does (e.g. `// increment counter` is banned).
* **The "Why" Exception:** Comments are allowed *only* to describe hidden business logic, third-party bug workarounds, or critical optimization contexts.
* **Refactoring Alternatives:** Split complex logic into single-responsibility functions; extract parameters into domain constants/primitives (`type ProjectBudget int64`); enforce descriptive naming.

---

## 3. 🏗️ Architectural Blueprint: Event-Driven Modular Monolith

Distributed microservices and structure use-case layers are banned. The system relies on a **Feature-Layered Modular Monolith** communicating via an **In-Memory Event-Driven Architecture (EDA)** backed by Go channels and a worker thread pool.

### Directory Layout
```
├── assets.go                       # embed.FS declarations (locales/*, web/*)
├── Dockerfile                      # Multi-stage containerized build
├── docker-compose.yml              # Local orchestration (Go app + PostgreSQL)
├── cmd/
│   └── api/
│       └── main.go                 # System entrypoint and boot lifecycles
├── internal/
│   ├── app/                        # Application routing, middlewares, and startup assembly
│   ├── config/                     # Environment configuration loader with .env parser
│   ├── core/                       # Shared platform database and security primitives
│   ├── eventbus/                   # Channel-backed asynchronous dispatcher and worker pool
│   ├── i18n/                       # Translation loaders and thread-safe dictionary
│   └── features/                   # Encapsulated domains (project_intake, messaging, portfolio, notifications)
├── locales/                        # Translation bundles (en, uk, ru)
├── scripts/                        # Database backup and health monitoring utilities
└── web/                            # Templates, assets (CSS/JS components), and layouts
```

---

## 4. 🛠️ Strict Implementation & Code Style Rules

### Golang Standards
* **Context Threading:** Pass down `context.Context` explicitly to all database, transaction, and client requests. Use decoupled contexts with timeout for background work.
* **Driver Access:** Use `jackc/pgx/v5/pgxpool` directly. Standard `database/sql` or ORMs are strictly banned.
* **Error Hygiene:** Wrap errors explicitly using `fmt.Errorf("layer/component: operation failed: %w", err)`. Never discard errors (`_ = operation()`).
* **Dangling Goroutines Ban:** Inline anonymous goroutines (`go func()`) are banned. Use lifecycle-managed named goroutines or the `MemoryEventBus` worker pool.

### Code Layout & Vertical Whitespace
* **Trailing Whitespace:** Zero trailing spaces or carriage returns in any codebase, config, or template file.
* **Vertical Spacing:** Exactly one blank line to separate distinct logical blocks inside a function. Double blank lines are banned.
* **Brace Alignment & Returns:** Enforce Go brace positioning (`func structure() { ... }`). Do not add blank lines immediately after `{` or before `}`. Ensure return statements are compact without preceding blank lines.

### Alpine.js & Go HTML Templates
* **No Build Steps:** Native Go `html/template` rendering backed by Alpine.js declarative bindings.
* **Component Encapsulation:** Inline `<script>` tags inside template bodies are banned. Encapsulate logic in `Alpine.data()` inside JS assets.

### HTMX Navigation & SessionState Lifecycle
HTMX body swaps (`hx-swap="outerHTML"`) must preserve state:
* **HTMX-Driven Navigation:** The `htmx:beforeSwap` handler in `app.js` sets the `tivri_htmx_nav` session storage flag. The inline `<head>` script checks this flag and skips the storage wipe.
* **Full Page Load:** A fresh browser refresh clears all `sessionStorage` form keys.
* **Banned Patterns:** Never use global variable guards (like `window.__initialized`); never scatter `sessionStorage.setItem` across individual button/link handlers. New form keys must be added to the `<head>` script in `base.layout.html`.

### CSS & Styling Standards
* **Inline Style Ban:** Direct use of inline style attributes is strictly forbidden. Custom layout, colors, or animations must be defined in utility classes or `web/assets/css/theme.css`.
* **Component Style Extraction:** Extract styles repeated 2+ times across template files into semantic CSS rules inside `web/assets/css/theme.css` to keep templates DRY. Single-use styles remain as utility classes.

---

## 5. 🔒 Data Security & Financial Guardrails

* **Financial Primitives:** Do not use floating-point types (`float32/64`) for currency. Use integers tracking subunits (cents) or arbitrary-precision libraries (`shopspring/decimal`).
* **SQL Injection:** Raw string concatenations in SQL are banned. Always use parameterized queries (`$1`, `$2`).
* **XSS Defenses:** Rely on Go's contextual auto-escaping in templates. Safely encode dynamic data passed to Alpine.js using JSON serialization.
* **Session Security:** Store dashboard access tokens in `HTTP-Only`, `Secure`, `SameSite=Strict` cookies.
* **Database Safety:** Run operations altering multiple tables in transactions (`pgx.Tx`) with explicit rollback. Configure strict pool parameters. Maintain automated daily backup cron scripts.

---

## 6. 🌐 Infrastructure Integration & Localization (i18n)

* **Nginx Policies:** Configure static file aliases for `/assets/` to bypass the Go app. Inject security headers (`CSP`, `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`). Apply `limit_req` zones to critical form endpoints.
* **i18n Translation:** Do not hardcode copy. Use localized bundles (`en.json`, `uk.json`, `ru.json`). Resolve language in order: URL subroute (`/en/`) -> Session Cookie -> `Accept-Language` header.

---

## 7. ⚠️ Proactive System Inspection Framework

Audit all implementations and document:
1. **Fault Tolerance / DLQ:** Handle Telegram/Email API timeouts and offline intervals.
2. **State Recovery:** Onboarding state must be retained across browser refreshes and device switches.
3. **Outbox Pattern Synchronization:** Ensure database and event bus consistency without blocking request threads.

---

## 8. 🔄 Client, Admin Dashboard, & Notification Alignment

* **Feature Parity:** All client form inputs must be viewable in the Admin Dashboard and sent in notification payloads.
* **Control Parity:** Admin configuration controls (toggles/modes) must dynamically update the client UI.
* **Propagation & Naming:** Propagations must be consistent across views, tables, and alerts. Keep variable and database names matching identical keys.

---

## 9. 📄 Documentation & Pipeline Synchronization Mandate

* **README & CI/CD Sync:** Update the `README.md` and CI/CD workflow pipelines concurrently with changes to features, compile schemas, configurations, directories, or build dependencies.
* **Automated Asset Generation:** All compiled stylesheets (`theme.css`), minified scripts (`.min.js`), and program binaries (`.exe`) must be kept out of version control via `.gitignore` and compiled dynamically inside Docker builders or local build scripts (`npm run build`).
* **Prerequisites:** Keep setup manuals and host scripts (e.g., `scripts/health_check.sh`) in sync.
* **Minimalist Git Commits:** Enforce compact, lowercase Conventional Commits (e.g., `feat: ...`, `fix: ...`, `docs: ...`, `ci: ...`) with titles under 50 characters. Keep branch names short and kebab-cased (e.g., `dev`, `feature/x`).
* **Single Source of Truth:** Platform documentation must always represent the current production state.

