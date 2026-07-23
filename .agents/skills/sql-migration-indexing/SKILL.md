---
name: sql-migration-indexing
description: "Guidelines for writing SQL migrations, composite indexing, and database performance optimization."
---

# SQL Migration & Composite Indexing Standards

## Trigger
Use this skill when creating database migration files, modifying schemas, or adding query filters in `internal/datastore/`.

## Key Invariants

### 1. Composite Database Indexes
- All query-filtered database columns (e.g. `client_status`, `internal_status`, `created_at`) must have composite indexes in SQL migrations for sub-millisecond query execution.
- Order composite index columns by cardinality and common query filtering combinations.

### 2. Migration Reversibility & Safety
- Ensure migrations apply cleanly without locking tables for extended periods.
- Avoid breaking existing API contracts or dropping columns in use without multi-phase migration steps.
