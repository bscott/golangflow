package main

import "embed"

//go:embed templates/*
var TemplatesFS embed.FS

//go:embed assets/*
var AssetsFS embed.FS

//go:embed locales/*
var LocalesFS embed.FS
