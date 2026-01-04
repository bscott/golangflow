package main

import "embed"

//go:embed all:templates
var TemplatesFS embed.FS

//go:embed all:assets
var AssetsFS embed.FS

//go:embed all:locales
var LocalesFS embed.FS
