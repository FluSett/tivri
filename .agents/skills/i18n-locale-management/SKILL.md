---
name: i18n-locale-management
description: "Guidelines for managing internationalization, translation keys, and locale sync across en.json, uk.json, and ru.json."
---

# Internationalization & Locale Management Standards

## Trigger
Use this skill when adding UI copy, introducing new template strings, or editing translation JSON files in `locales/`.

## Key Invariants

### 1. Synchronous Key Updates
- Every new translation key added to `locales/en.json` must be added simultaneously to `locales/uk.json` and `locales/ru.json`.
- Maintain consistent nested JSON key structures across all locale dictionaries.

### 2. Template Key Safety
- Verify that every translation key referenced in Go templates exists in the i18n dictionary.
- Avoid dynamic string concatenation inside templates that can bypass translation lookup.

### 3. Server-Side Translation Resolution
- Resolve localized text strings on the server during template execution.
- Pass locale identifiers cleanly through the HTTP context/middleware.
