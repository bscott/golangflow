package locales

import (
	"embed"
	"io/fs"

	"github.com/gobuffalo/buffalo"
)

//go:embed *.yaml
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "locales")
}
