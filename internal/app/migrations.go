package app

import "embed"

//go:embed migrations/*.sql
var postgresMigrationFS embed.FS
