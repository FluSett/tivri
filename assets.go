package tivri

import "embed"

//go:embed locales/*
var LocalesFS embed.FS

//go:embed web/*
var WebFS embed.FS
