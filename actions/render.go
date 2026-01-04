package actions

import (
	"embed"
	"html/template"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/helpers/hctx"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/tags"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var r *render.Engine
var templatesFS embed.FS
var assetsFS embed.FS

// SetEmbedFS sets the embedded filesystems for templates and assets
func SetEmbedFS(templates, assets embed.FS) {
	templatesFS = templates
	assetsFS = assets
	initRender()
}

func initRender() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "templates/application.html",

		// Embedded filesystems for templates and assets:
		TemplatesFS: templatesFS,
		AssetsFS:    assetsFS,

		// Add template helpers here:
		// https://github.com/gobuffalo/plush/issues/111
		Helpers: render.Helpers{
			"paginator": func(pagination *pop.Paginator, opts map[string]interface{}) (template.HTML, error) {
				t, err := tags.Pagination(pagination, opts)
				if err != nil {
					return "", errors.WithStack(err)
				}
				return t.HTML(), nil
			},
		},
	})
}


