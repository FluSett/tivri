package tivri

import "embed"

//go:embed locales/*
var LocalesFS embed.FS

//go:embed services/web/ui/*
var WebFS embed.FS
