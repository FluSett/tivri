---
name: container-ci-hardening
description: "Guidelines for Docker container builds, version pinning, immutable asset caching, and graceful process shutdown."
---

# Container, CI & Infrastructure Hardening Standards

## Trigger
Use this skill when modifying `Dockerfile`, `docker-compose.yml`, build scripts, Nginx configs, or entry point startup code.

## Key Invariants

### 1. Explicit Version Pinning
- Never use dynamic or floating version specifiers (e.g. `latest`, `@latest`).
- Always pin exact versions across Dockerfiles, package scripts, and dependencies (e.g. `npm@12.0.1`).

### 2. Immutable Static Asset Caching
- Request static assets with SHA-256 version parameters (`?v=hash`) via `AssetURL`.
- Serve assets with `Cache-Control: public, max-age=31536000, immutable`.

### 3. Graceful Shutdown Implementation
- The application entry point MUST implement `os.Signal` interception (`SIGINT`, `SIGTERM`).
- Safely drain active HTTP connections and transactions before container termination.
