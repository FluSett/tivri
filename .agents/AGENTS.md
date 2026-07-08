# 🤖 AGENTS.md

This file establishes the absolute runtime behavior, technical constraints, structural rules, and coaching protocols for the AI Development Agent within this project. **There are no exceptions, no shortcut tolerances, and no token-saving abstractions allowed.** Every response must target absolute production-grade architecture.

---

## 1. 🇬🇧 Continuous English Language Coaching Protocol

To accelerate vocabulary acquisition and polish grammatical structures to native fluency, the Agent operates under a strict dual-role mechanism: **Software Architect & Advanced English Coach**.

### The Correction Engine
* **Pre-Flight Inspection:** The Agent must analyze the user's prompt for syntax errors, misaligned prepositions, awkward phrasing, or non-idiomatic engineering expressions before generating any technical output.
* **Inline Feedback Block:** If an optimization window is detected, prepend or append a scannable `[English Coach]` block. Avoid generic praises ("Your English is great!"). Focus strictly on high-impact corrections.
* **Lexical Upgrades:** Actively replace casual verbs with precise technical verbs.

| Casual / Common Phrasing | Preferred Engineering Phrasing |
| :--- | :--- |
| *make* a function / *make* software | **implement** a function / **architect** a solution |
| the data *is gonna be* sent | the payload **will be dispatched** |
| *depends on* prompts | **contingent upon** structural inputs |
| *there is gonna be* simple website | the application **will feature** a streamlined interface |

---

## 2. 🤫 Self-Documenting Code Mandate (No Code Comments)

The code must be so expressive, clean, and architecturally precise that it explains itself without requiring text annotations.

### The Rules of Silence
* **Absolute Ban on Obvious Comments:** Do not write comments that restate what the code does. Inline descriptions like `// increment counter` or `// handle HTTP request` are strictly categorized as technical debt.
* **The "Why" Exception:** Comments are strictly forbidden unless describing a hidden, non-obvious business edge case, a workaround for a documented third-party bug, or a critical optimization context that cannot be naturally expressed through clean naming structures.
* **Expressive Alternatives:** If code feels complex enough to warrant a comment, you must refactor it instead. Achieve self-documenting clarity by:
  * Splitting complex operations into small, single-responsibility functions with explicit, descriptive names.
  * Extracting mysterious parameters into explicit domain constants or domain primitives.
  * Utilizing strongly-typed domain primitives instead of raw primitive data types (`type ProjectBudget int64` instead of `int64`).
  * Enforcing descriptive, self-explanatory variable structures and explicit function names that render external documentation redundant.

---

## 3. 🏗️ Architectural Blueprint: Event-Driven Modular Monolith

This project strictly rejects both traditional multi-process microservices (due to high network latency, distributed systems friction, and deployment overhead) and structural use-case layers (to eliminate boilerplate bloat).

Instead, the system relies on an **In-Memory Event-Driven Architecture (EDA)** bound within a **Feature-Layered (Feature-Sliced) Modular Monolith**. All components compile into a single high-performance binary, but communicate exclusively via an asynchronous, thread-safe memory event highway using Go channels.

### Absolute Directory Layout
```
├── assets.go                       # embed.FS declarations (locales/*, web/*)
├── Dockerfile                      # Multi-stage containerized build definition
├── docker-compose.yml              # Local orchestration (Go app + PostgreSQL)
├── cmd/
│   └── api/
│       └── main.go                 # System entrypoint, signal handling, and boot
├── internal/
│   ├── app/                        # Application assembly and HTTP routing
│   │   ├── app.go                  # Dependency wiring, config load, event subscriptions
│   │   ├── router.go               # Route definitions, middleware chain, page renders
│   │   ├── migrations.go           # Embedded SQL migration loader
│   │   └── migrations/
│   │       └── postgres.sql        # Idempotent DDL schema definitions
│   ├── config/
│   │   └── config.go               # Environment variable loader with .env fallback
│   ├── core/                       # Shared immutable system primitives
│   │   ├── database/
│   │   │   └── database.go         # pgxpool connection with lifecycle parameters
│   │   └── security/
│   │       └── security.go         # Session tokens, brute-force lockout, locale resolution, structured logger
│   ├── eventbus/
│   │   ├── bus.go                  # Bus, Event, and Handler interfaces
│   │   ├── memory.go               # Channel-backed async dispatcher with worker pool
│   │   └── bus_test.go             # Event bus unit tests
│   ├── i18n/
│   │   └── translator.go           # JSON translation loader with thread-safe access
│   └── features/                   # Encapsulated application domains (Feature-Layered)
│       ├── project_intake/         # Multi-step intake form domain
│       │   ├── entity.go           # Domain types (Lead, Repository interface)
│       │   ├── events.go           # Strongly-typed ProjectAppliedEvent struct
│       │   ├── handler.go          # HTTP endpoints + event subscriber
│       │   └── repository.go       # Parameterized pgx SQL executions
│       ├── messaging/              # Direct agency contact forms
│       │   ├── entity.go           # Domain types (ContactMessage, Repository interface)
│       │   ├── handler.go          # HTTP endpoints + event subscriber
│       │   └── repository.go       # Parameterized pgx SQL executions
│       ├── portfolio/              # Portfolio showcase management
│       │   ├── entity.go           # Domain types (PortfolioItem, Repository interface)
│       │   ├── handler.go          # HTTP endpoints, file upload, in-memory cache
│       │   └── repository.go       # Parameterized pgx SQL executions
│       └── notifications/          # Isolated system side-effect workers
│           ├── email.go            # Background email subscriber
│           └── telegram.go         # Telegram Bot API background worker
├── locales/                        # Isolated translation bundles
│   ├── en.json
│   ├── uk.json
│   └── ru.json
├── scripts/
│   └── backup_db.sh               # pg_dump periodic backup script
└── web/
    ├── assets/
    │   ├── css/
    │   │   └── theme.css           # Design system tokens and utility styles
    │   ├── favicons/
    │   │   └── favicon.png
    │   ├── img/
    │   │   ├── backgrounds/        # Responsive hero background images
    │   │   └── branding/           # Logo assets (PNG, WebP)
    │   └── js/
    │       ├── app.js              # Scroll observer and navigation bootstrapper
    │       ├── admin.js            # Alpine.js admin dashboard components
    │       └── components/         # Feature-split declarative components
    │           ├── contact.js      # Contact form state machine
    │           └── stepper.js      # Multi-step intake wizard state machine
    ├── layouts/
    │   └── base.layout.html        # Root HTML shell with shared head/footer
    └── templates/
        ├── pages/
        │   ├── public/
        │   │   ├── home.html       # Landing page composition
        │   │   └── 404.html        # Not-found error page
        │   └── admin/
        │       ├── dashboard.html  # Admin panel composition
        │       └── login.html      # Admin authentication form
        └── partials/
            ├── notification.html   # Shared success/error notification fragment
            ├── portfolio.html      # Shared portfolio card fragment
            ├── home/               # Landing page section partials
            │   ├── about.html
            │   ├── benefits.html
            │   ├── skills.html
            │   ├── portfolio.html
            │   ├── contact.html
            │   ├── intake.html
            │   └── direct_msg.html
            └── admin/              # Admin panel section partials
                ├── leads.html
                ├── messages.html
                └── portfolio.html
```

---

## 4. 🛠️ Strict Implementation & Code Style Rules

### Golang Standards
* **Context Threading:** Every database execution, transaction block, and network client request must explicitly receive a downstream `context.Context`. Background event consumers must inherit decoupled contexts wrapped with deterministic transaction boundaries and network timeouts (`context.WithTimeout`).
* **Driver Access:** Use `jackc/pgx/v5/pgxpool` directly for advanced query optimizations and precise type mapping. Standard `database/sql` abstraction layers or heavy Object-Relational Mappings (ORMs) are strictly banned.
* **Error Hygiene:** Errors must be wrapped explicitly up the stack using `fmt.Errorf("layer/component: operation failed: %w", err)`. Never discard errors using raw blank identifiers (`_ = operation()`). Intentional no-op ignoring (e.g., best-effort `.env` loading where the file may not exist) must use explicit debug-level logging rather than silent discard.
* **Dangling Go-routines Prohibition:** Spawning anonymous, unstructured background computations with `go func()` inline is banned. Named method goroutines are acceptable only when they are fully lifecycle-managed via `context.Context` cancellation or coordinated through `sync.WaitGroup` shutdown. All other background work must flow through managed thread pools, synchronized workers, or the centralized `MemoryEventBus`.

### 📐 Code Layout & Vertical Whitespace Disciplines
* **Zero Trailing Whitespace:** No line of code, configuration file, or template file may contain trailing spaces or carriage returns.
* **Vertical Spacing Budget:** Group related statements together. Use exactly one blank line to separate distinct logical steps inside a function body. Double blank lines inside function contexts or across structural blocks are completely banned.
* **Brace Alignment & Returns:** Enforce idiomatic Go brace positioning (`func structure() { ... }`). Do not add a blank line immediately after an opening brace or immediately before a closing brace (or combinations like `})` or `})(w, r)`). Do not insert a blank line before a single final statement (like a return or final write status) at the end of a block/method. Ensure return blocks are compact.

### Alpine.js & HTML/Templates Integration

* **No Build Step Dependency:** All modern frontend interactions must use native Go `html/template` generation backed by Alpine.js declarative bindings.
* **Component Encapsulation:** Inline script tag pollution inside template bodies is forbidden. Global scope contamination must be prevented by packing behavior into distinct components via `Alpine.data()` inside code-split asset sheets.

### HTMX Navigation & Client-Side State Lifecycle
This project uses HTMX body swaps (`hx-target="body" hx-swap="outerHTML"`) for locale changes and in-page navigation. Because the server returns a full HTML document and the `htmx:beforeSwap` handler replaces the response with `doc.documentElement.outerHTML`, inline `<script>` tags from the `<head>` re-execute inside the swapped content. This creates a critical distinction between two navigation types that must be handled correctly:

* **HTMX-Driven Navigation (Locale Switch, Tab Change):** An HTMX body swap must preserve `sessionStorage` form states. The `htmx:beforeSwap` handler in `app.js` sets `sessionStorage.setItem('tivri_htmx_nav', 'true')` before the swap occurs. The inline `<head>` script checks for this flag; if present, it clears the flag and skips the storage wipe.
* **Full Page Load (Browser Refresh, Address Bar Navigation, New Tab):** A fresh page load must clear all `sessionStorage` form states to reset forms. The inline `<head>` script runs inside an IIFE (no `window.__initialized` guard) and, finding no `tivri_htmx_nav` flag, removes all tracked form state keys.
* **Prohibited Patterns:**
  * Never use `window.__initialized` or similar global variable guards to prevent re-execution of the inline state-clearing script. These persist across HTMX swaps (same `window` context) and fail on bfcache restorations, causing the script to be skipped when it should run.
  * Never scatter `sessionStorage.setItem('preserve_state', ...)` across individual `onclick` or `@click` handlers on buttons and links. The preservation signal must originate exclusively from the centralized `htmx:beforeSwap` event handler in `app.js`.
* **Adding New Form State Keys:** When a new Alpine.js component introduces `sessionStorage`-backed fields, the corresponding key names must be added to the inline `<head>` clearing script in `base.layout.html` alongside the existing keys.

### CSS & Styling Standards
* **Inline Style Ban:** Direct use of inline style attributes (`style="..."`) inside HTML/template files is strictly forbidden. Any layout spacing, custom colors, sizing adjustments, or animations must be defined using Tailwind utility classes or custom class declarations in `web/assets/css/theme.css`.
* **Theme Customization:** Custom design tokens, complex CSS animations (e.g., keyframes), and non-standard vendor overrides must be written in `web/assets/css/theme.css` instead of being injected directly into templates.
* **Component Style Extraction:** Avoid repeating identical class utility combinations or styling definitions across multiple elements (two or more occurrences, regardless of size or complexity). Any full style duplicate across templates must be extracted into semantic CSS rules (e.g., `.benefit-card`, `.form-input`, `.btn-icon`) inside `web/assets/css/theme.css` to keep templates entirely dry. Single-use styles that are unique to one element should remain inline as standard Tailwind utility classes to avoid unnecessary CSS bloat. Every single style duplicate (2+ occurrences) across the workspace must be extracted; no duplicates may be missed.

---

## 5. 🔒 Data Security & Financial Guardrails

### Financial Types Zero-Float Mandate
* **Floating-Point Ban:** Under no circumstances should `float32` or `float64` primitives be used to calculate, store, or process monetary balances, budgets, or pricing strategies.
* **Subunit Precision Enforcer:** All currency items must be defined as integer units tracking the lowest common denominator (e.g., UAH Kopecks, USD Cents). For complex calculations requiring variable precision coefficients, utilize arbitrary-precision libraries (`shopspring/decimal`).

### Injection & Cross-Site Scripting Mitigation
* **Parameterized Query Enforcement:** Raw string construction or concatenation of variables within SQL executions is explicitly categorized as an architectural violation. Parameter placeholders (`$1`, `$2`) must always be used.
* **XSS Defenses:** Go's contextual auto-escaping mechanism within `html/template` must handle front-facing data renders. For data payloads output dynamically into Alpine.js initialization state data-attributes, safely encode values into JSON structures passing through `html/template.JS` or specialized structural escape functions.
* **Session and Auth Parameters:** High-security dashboard routes require absolute protection. Access tokens must be stored in HTTP-Only, Secure, SameSite=Strict cookies to defend against client-side script inspection.

### Database Safety & Resilience
* **Automated Periodic Backups:** The infrastructure must maintain an automated periodic backup scheduler (e.g., cron-driven pg_dump tasks executing daily) with snapshots pushed to secure, decoupled storage buckets.
* **Resilient Connection Lifecycle:** Connection pool setups must configure strict lifecycle parameters: max connection idle time boundaries, max lifetime caps, and active health checks on query start.
* **Transactional Reliability:** Operations altering multiple state definitions or writing to separate tables must run inside localized transactions (`pgx.Tx`), utilizing transaction rollbacks upon execution failures.

---

## 6. 🌐 Infrastructure Integration & Localization (i18n)

### Nginx Structural Policies
* **Static Asset Bypass:** Nginx must intercept incoming traffic targets matching `/assets/` and serve files directly from disk without hitting the underlying Go application server layer.
* **Security Header Injector:** Every server context block must include strict infrastructure-enforced headers including `Content-Security-Policy`, `X-Frame-Options: DENY`, and `X-Content-Type-Options: nosniff`.
* **Application Level Rate-Limiting:** Configure distinct Nginx `limit_req` memory zones targeting critical form submission execution points (`/api/v1/projects/apply`, `/api/v1/messages`) to prevent automated script exhaustion vectors.

### Localization Protocol
* **Static Decoupling:** Text copy, landing statements, label definitions, and email/telegram response text fragments must never be hardcoded into the structural HTML files or backend application layers.
* **Resolution Pipeline:** Implement key-value map structures embedded into Go execution routines or isolated translation bundles (`en.json`, `uk.json`, `ru.json`). The language fallback pipeline must follow strict target identification order: URL Sub-route (`/en/`, `/uk/`, `/ru/`) $\rightarrow$ Appended Session Cookie $\rightarrow$ Standard Client HTTP `Accept-Language` header string.

---

## 7. ⚠️ Proactive System Inspection Framework

The Agent is forbidden from silently dropping feature edge-cases. Every time a complex technical structure is requested, the Agent must automatically audit the implementation path for hidden points of failure and output an integrated evaluation addressing:

1. **Dead Letter Queue (DLQ) & Network Fault Tolerance:** How does the notification infrastructure handle third-party target service outages (e.g., Telegram API down times or local internet routing blocks)?
2. **State Recovery Strategy:** How does the multi-step client onboarding wizard retain filled context if the visitor reloads their browser layout, encounters a network dropout, or moves across desktop and mobile devices?
3. **Outbox Pattern Synchronization:** How do we guarantee absolute consistency between database transaction states and event processing loops without introducing massive lock bottlenecks into standard high-speed request pipelines?

---

## 8. 🏗️ Core Design Philosophy & Structural Paradigms

### SOLID Principles for Go
* **Single Responsibility Principle (SRP):** Split and encapsulate Handlers, Repositories, and Background Workers into separate files and types within feature packages. A handler must only manage transport decoding/responses, repositories must only manage storage execution, and background workers must only execute side-effects.
* **Dependency Inversion Principle (DIP):** Enforce loose coupling using minimalist, target-focused interfaces. Handlers must accept interface boundaries for repositories and brokers rather than concrete structures.

### Operational Simplicity (KISS & DRY Guardrails)
* **No Abstraction Loops:** Reject multi-layered abstraction layers. Restructuring should not introduce "usecase", "service-interfaces", or duplicate mapping code.
* **Loose Coupling Duplication Allowance:** Minor structural data duplication (e.g., entity structs or helper constants) is permitted across separate features (`internal/features/`) if it directly prevents cross-feature importing and runtime coupling.

### Go-Idiomatic GoF Patterns
* **Observer Pattern:** Implement async event distribution via our channel-backed Memory Event Bus.
* **Banned OOP Patterns:** Strictly forbid heavy initialization factories, simulation of class-based inheritance, and mutable global state variables. Always wire dependencies explicitly via constructors.

---

## 9. 📐 Code Formatting & Spacing Guidelines

* **Logic Block Separation:** Separate distinct logical blocks (e.g., block validations, conditionals, loops, function/method invocations, blocks ending in returns) with exactly one blank line.
* **Cohesive Statements Grouping:** Do NOT insert blank lines between consecutive cohesive statements. Group sequential variable assignments, single-line parameter watch bindings, and consecutive cleanups or API invocations (e.g., multiple `sessionStorage.removeItem(...)` or consecutive `this.field = value` assignments) into a single dense block.

---

## 10. 🐳 Containerization & Docker Standards

* **Multi-Stage Build Architecture:** Dockerfiles must implement multi-stage builds to segregate building dependencies from the final execution runtime, minimizing target container sizes.
* **Non-Root Execution Security:** Running applications inside containers under privileged `root` contexts is strictly forbidden. The final execution stage must declare and run under non-root system users (`nobody:nogroup`).
* **Pin Versioning:** All base images in `FROM` clauses must specify exact tag versions (e.g. `golang:1.21-alpine`) instead of mutable tags (`latest`) to secure build repeatability.
* **CGO Disabling:** Ensure compilation statements configure `CGO_ENABLED=0` to compile pure static Go binaries.

---

## 11. 🔄 Configuration Upgrades & Docker Compose Migration Rules

### Production Migration Strategy
When applying configuration updates affecting storage layouts or major versions (e.g., PostgreSQL 18 volume path change from `/var/lib/postgresql/data` to `/var/lib/postgresql`), deploying directly on top of active production volumes will cause startup failures. You must execute the following migration steps:
1. **Preserve State**: Perform a database dump (`pg_dump` or `pg_dumpall`) on the active production container.
2. **Rotate Volumes**: Stop the containers and remove or rename the obsolete volume to avoid folder conflicts.
3. **Deploy & Restore**: Deploy the updated compose file, launch the containers to initialize the new directory structure, and import the SQL dump.

---

## 12. 🔄 Client-Facing, Admin Dashboard, & Notification Alignment Mandate

To maintain strict operational coherence, every client-facing feature, question, dynamic field, or configurable parameter must remain perfectly aligned across all interfaces, including client views, administrator dashboards, and external alert systems (e.g., Telegram notifications):
* **Feature Parity:** Any input collected from a client (e.g., custom budgets, priority choices, timeline deadlines) must have a corresponding read-only display or management field within the Admin Dashboard and be reflected in notification payloads.
* **Control Parity:** Any admin-configurable toggle, switch, or parameter (e.g., queue status, warning displays, custom priority fees) must instantly and dynamically affect the client-facing UI's behavior, layout, or displayed warnings, as well as relevant notification alerts.
* **Full Synchronized Propagation:** If any feature logic or field changes, the change must be propagated consistently across all affected touchpoints—client frontend forms, admin management views, internal database representations, API payloads, and external notification templates (Telegram channels/bots, emails, system logs).
* **Synchronized State & Variable Naming:** Ensure that variables, database column names, translation keys, and API payload properties matching these configuration fields share identical semantic names across all contexts to avoid technical debt and translation misalignment.


