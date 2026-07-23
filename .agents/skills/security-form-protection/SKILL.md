---
name: security-form-protection
description: "Guidelines for form security, Cloudflare Turnstile integration, CSP headers, rate-limiting, and payload validation."
---

# Form Security & Turnstile Integration Standards

## Trigger
Use this skill when modifying intake forms, security headers, rate limiters, or Cloudflare Turnstile verification.

## Key Invariants

### 1. Cloudflare Turnstile Integration
- Manage Turnstile widgets via clean JS modules (e.g. `web/assets/js/core/turnstile.js`).
- Cleanly handle widget auto-hiding or transition states on successful token validation.
- Verify Turnstile tokens on the server endpoint before processing intake submissions.

### 2. HTTP Security Response Headers
- Enforce strict Content-Security-Policy (CSP) headers in middleware/Nginx.
- Include mandatory security headers: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, and `Permissions-Policy`.

### 3. Intake Form Sanitization & Validation
- Validate all user input server-side before persisting or forwarding intake requests.
- Rate-limit public form submissions to mitigate spam and automated abuse.
