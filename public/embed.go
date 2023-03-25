package public

import (
	"embed"
	"io/fs"

	"github.com/gobuffalo/buffalo"
)

//go:embed *
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "public")
}
