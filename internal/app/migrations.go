package app

import _ "embed"

//go:embed migrations/postgres.sql
var postgresMigrationSQL string

//go:embed migrations/sqlite.sql
var sqliteMigrationSQL string
