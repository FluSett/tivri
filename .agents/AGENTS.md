# 🤖 AGENTS.md

Defines absolute runtime behavior and constraints. No exceptions.

## 1. 🇬🇧 English Coaching
- **MANDATORY**: ALWAYS append an `### 🇬🇧 English Coach` section at the end of your response to gently correct the user's grammar, typos, and phrasing from their latest message.
- **Natural Flow**: Prioritize fluent, natural phrasing over rigid jargon (e.g., "build a website", not "make a website"). Suggest these improvements in the English Coach section.

## 2. 🤫 Self-Documenting Code
- **No Obvious Comments**: Comments restating logic are banned.
- **"Why" Exception**: Only comment on hidden business logic, bug workarounds, or critical optimizations.
- **Refactoring**: Split complex logic, enforce descriptive naming instead of commenting.

## 3. 📄 Standards & CI/CD
- **Sync Docs**: Update `README.md` and CI/CD concurrently with changes.
- **No Compiled Assets in Git**: Exclude minified JS, compiled CSS, and binaries (`.gitignore`).
- **Compact Git Data**: Keep commit messages, branch names, merge titles, and merge descriptions extremely compact and filler-free. Use ultra-concise Conventional Commits (e.g., `feat: short desc`) under 50 characters. Use short `kebab-case` for branches.
- **Single Source of Truth**: Docs must reflect current production state.
- **Structured Logging (`slog`)**: Use Go 1.21+ `log/slog` for all application logging with contextual attributes (`slog.Info`, `slog.Error`), prohibiting raw `fmt.Printf` or unstructured `log.Printf` calls.

## 4. 🚫 No Magical Variables
- **Dynamic Configuration**: Never hardcode domains (e.g. `tivri.cc`), emails, ports, or API endpoints. Inject them via environment variables and pass them down into handlers/templates.
- **Named Constants**: Never leave arbitrary `time.Second` multipliers or numeric literals scattered in business logic. Extract them to clear, localized `const` declarations.
- **No Inline JavaScript**: Never use inline `<script>` tags in HTML templates (except for non-executable metadata like JSON-LD). Pass dynamic data using HTML5 data attributes (`data-*`) on elements and retrieve them via external modular Vanilla JS.

## 5. ♻️ DRY Architecture
- **HTML & Templates**: Never duplicate identical HTML structure. Extract reusable UI elements into Go template components and inject data via the `dict` helper.
- **Frontend State**: Use clean, modular Vanilla JS components and centralized state management to handle state storage natively without relying on heavy frameworks like Alpine.js or React.
- **Server-Side Formatting**: Never duplicate data formatting logic (like dates or currency) in JavaScript. Render all formatted data natively on the server using Go templates (e.g. `{{.CreatedAt.Format "2006-01-02 15:04"}}`) and pass it via HTMX.
- **Modular JavaScript**: Break large JavaScript logic into smaller, feature-specific modules (e.g., in `core/`) and import them into `app.js` using ES modules to prevent monolithic scripts. Always use the shared utilities in `core/` (like `validators.js`, `storage.js`, and `dom.js`) instead of re-implementing DOM manipulation, validation, or `sessionStorage` operations.
- **Modular Go Handlers**: Keep HTTP handler files focused by sub-domain responsibility. Split large handlers (over ~300 lines) into dedicated files within the package (e.g., `admin_auth.go`, `admin_leads.go`, `admin_portfolio.go`) while sharing the core handler struct.
- **Centralized CSS vs JS State**: Consolidate purely visual UI patterns into modular CSS files (e.g., `components.css`). However, NEVER abstract Tailwind utilities that are dynamically toggled by JavaScript (like `hidden`, `opacity-0`, `translate-x-full`) into custom CSS classes via `@apply`. These stateful utilities MUST remain explicitly inline on the HTML elements, otherwise `classList` manipulation will fail.

## 6. 🛠️ Lint & Static Warnings Code Quality
- **Always Resolve Warnings**: Never ignore warnings or errors reported by linters or compilation tools. Proactively refactor any hardcoded style hex values, invalid Tailwind classes, Go warnings, or structural code analysis notices.
- **Static Code Analysis (`golangci-lint`)**: All Go source files must pass `.golangci.yml` linting checks (`gosec`, `staticcheck`, `govet`, `errcheck`, `goconst`) without warnings.
- **Tailwind Conflicting Classes**: The Tailwind Intellisense plugin frequently flags mutually exclusive state classes (e.g., `text-white` vs `placeholder:text-neutral-500`, or `peer-checked:bg-primary` vs `peer-not-checked:bg-white/10`) as conflicts if they exist on the same element. To resolve these cleanly without writing inline styles or disabling the linter, completely separate the classes across different DOM elements (e.g., placing `text-white` on a wrapper `<div>` and `placeholder:text-neutral-500` on the child `<input>`, relying on Tailwind's native `inherit` behavior).

## 7. 🛡️ Security & Reliability Invariants
- **HTMX Native Operations**: Never build custom global event routing for form submissions. You must rely natively on HTMX attributes (`hx-post`, `hx-target`) for server communication. Vanilla JS is strictly reserved for visual micro-interactions and HTMX lifecycle hooks.
- **Strict Context Timeouts**: Never pass an unbounded HTTP context directly into a database query. All datastore operations must enforce `context.WithTimeout` wrappers.
- **Composite Database Indexes**: All query-filtered database columns (e.g. `client_status`, `internal_status`, `created_at`) must have composite indexes in SQL migrations for sub-millisecond query execution.
- **Content-Security-Policy & Response Headers**: Nginx & Go middleware must strictly enforce CSP, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, and `Permissions-Policy`.
- **Immutable Asset Caching**: Static assets must be requested with SHA-256 version parameters (`?v=hash`) via `AssetURL` and served with `Cache-Control: public, max-age=31536000, immutable`.
- **Graceful Shutdown**: The entry point MUST implement `os.Signal` interception to safely drain connections and active transactions before container termination.
