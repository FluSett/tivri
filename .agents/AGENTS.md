# 🤖 AGENTS.md

Defines absolute runtime behavior and constraints. No exceptions.

## 1. 🇬🇧 English Coaching
- **MANDATORY**: ALWAYS append an `### 🇬🇧 English Coach` section at the end of your response to gently correct the user's grammar, typos, and phrasing from their latest message.
- **Natural Flow**: Prioritize fluent, natural phrasing over rigid jargon (e.g., "build a website", not "make a website"). Suggest these improvements in the English Coach section.

## 2. 🤫 Code Quality & Self-Documenting Code
- **No Obvious Comments**: Comments restating logic are banned; comment only hidden business logic, bug workarounds, or critical optimizations.
- **Zero Warnings**: Code must pass `golangci-lint` without warnings. Proactively resolve all linter, compiler, and static code analysis notices.

## 3. 📄 Standards, Docs & Git
- **Sync Docs**: Update `README.md` concurrently with feature or architecture changes.
- **Granular & Compact Git**: Always split work into multiple logical, atomic commits rather than one monolithic commit. Keep commit messages ultra-concise using Conventional Commits under 50 characters (e.g. `feat: short desc`). Exclude compiled assets and binaries from git tracking.
- **Compact PR Specifications**: When generating PR titles and descriptions, use Conventional Commit PR titles under 60 characters and ultra-compact bullet points in descriptions without unnecessary Summary or Verification headers.

## 4. 🚫 Dynamic Configuration & Named Constants
- **No Magical Variables**: Never hardcode domains, ports, emails, or API endpoints. Inject dynamic parameters via environment variables.
- **Named Constants**: Never leave raw numeric literals or time multipliers scattered in business logic. Extract them to clear, localized `const` declarations.

## 5. 🎯 Domain Execution Skills
- Refer to active skills in `.agents/skills/` (`go-handler-db`, `htmx-component`, `vanilla-js-module`, `i18n-locale-management`, `security-form-protection`, `sql-migration-indexing`, `container-ci-hardening`) for domain-specific implementation rules and invariants.

## 6. 🔄 Unification & Standardization
- **Zero Raw Storage Calls**: Never invoke `localStorage` or `sessionStorage` methods directly in components or handlers. Always use unified wrapper functions from `web/assets/js/core/storage.js` (`setLocalItem`, `getLocalItem`, `setSessionItem`, `getSessionItem`).
- **Unified Abstractions**: Always use centralized helper modules (`core/storage.js`, `core/dom.js`, `core/state.js`, `core/validators.js`) instead of re-implementing DOM manipulation, storage lookups, or validation routines.

