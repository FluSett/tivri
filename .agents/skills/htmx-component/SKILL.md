---
name: htmx-component
description: "Guidelines for creating and modifying HTMX-powered Go HTML template components and partial responses."
---

# HTMX & Go Template Component Standards

## Trigger
Use this skill when editing or creating Go HTML templates in `web/templates/` or implementing HTMX partial responses.

## Key Invariants

### 1. HTMX Native Operations
- Rely natively on HTMX attributes (`hx-post`, `hx-target`, `hx-swap`) for form submissions and server communication.
- Do not build custom global event routing in JavaScript for form handling.

### 2. DRY UI Components
- Extract reusable UI elements into Go template components under `web/templates/partials/components/`.
- Inject component data cleanly using the `dict` helper.

### 3. Server-Side Data Formatting
- Never duplicate date or currency formatting logic in JavaScript.
- Render all formatted data natively on the server using Go templates (e.g. `{{.CreatedAt.Format "2006-01-02 15:04"}}`).

### 4. Stateful Tailwind Utilities
- Keep JavaScript-toggled state utility classes (e.g., `hidden`, `opacity-0`, `translate-x-full`) explicitly inline on HTML elements.
- Never abstract stateful Tailwind utilities into custom CSS via `@apply`, as JS `classList` manipulations rely on exact class names.
