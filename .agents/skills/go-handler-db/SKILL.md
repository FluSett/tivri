---
name: go-handler-db
description: "Guidelines for developing Go HTTP handlers, database queries with context timeouts, slog structured logging, and Go 1.21+ standards."
---

# Go HTTP Handler & Datastore Standards

## Trigger
Use this skill when developing or modifying Go HTTP handlers, middleware, service logic, or datastore access layers in `internal/`.

## Key Invariants

### 1. Structured Logging (`slog`)
- Use Go 1.21+ `log/slog` for all application logging with contextual key-value attributes (`slog.Info`, `slog.Error`).
- Prohibit raw `fmt.Printf`, `fmt.Println`, or unstructured `log.Printf` calls.

### 2. Strict Context Timeouts
- Never pass an unbounded HTTP context directly into a database query.
- Wrap all datastore operations in explicit `context.WithTimeout` contexts (e.g. `ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)`).

### 3. Modular Handler Files
- Keep HTTP handler files focused by sub-domain responsibility.
- Split handlers exceeding ~300 lines into dedicated files within the package (e.g., `admin_auth.go`, `admin_leads.go`, `admin_portfolio.go`) while sharing the core handler struct.

### 4. Zero Hardcoded Magic Values
- Extract all duration multipliers or numeric constants to clear, localized `const` declarations.
- Inject dynamic configuration (domains, ports, endpoints) via environment variables.
