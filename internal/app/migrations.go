package app

import _ "embed"

//go:embed migrations/postgres.sql
var postgresMigrationSQL string
