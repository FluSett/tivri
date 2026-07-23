---
name: vanilla-js-module
description: "Guidelines for creating modular Vanilla JS frontend micro-interactions, avoiding inline scripts, and passing dynamic data via HTML5 data attributes."
---

# Modular Vanilla JS Standards

## Trigger
Use this skill when developing or refactoring frontend JavaScript in `web/assets/js/`.

## Key Invariants

### 1. No Inline JavaScript
- Never write inline `<script>` tags inside HTML templates (except non-executable metadata like JSON-LD).
- Pass dynamic server values using HTML5 `data-*` attributes and retrieve them via external modular JS.

### 2. Modular Architecture & Imports
- Break frontend logic into small, single-responsibility ES modules under `web/assets/js/core/`.
- Use shared utilities in `core/` (`validators.js`, `storage.js`, `dom.js`) instead of re-implementing DOM manipulation or `sessionStorage` routines.

### 3. HTMX Lifecycle Integration
- Reserve Vanilla JS for visual micro-interactions (e.g. scroll effects, theme toggles, widget management) and listening to HTMX lifecycle events (`htmx:afterSwap`, `htmx:configRequest`).
