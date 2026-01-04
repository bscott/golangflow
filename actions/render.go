package actions

import (
	"embed"
	"html/template"
	"io/fs"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
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
	// Use fs.Sub to create a filesystem rooted at templates/
	templatesSubFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		panic(err)
	}

	// Wrap with Buffalo's FS for proper handling
	wrappedTemplatesFS := buffalo.NewFS(templatesSubFS, "templates")

	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Embedded filesystems for templates and assets:
		TemplatesFS: wrappedTemplatesFS,
		AssetsFS:    assetsFS,

		// Add template helpers here:
		// https://github.com/gobuffalo/plush/issues/111
		Helpers: render.Helpers{
			"getAvatar": func(id uuid.UUID, help hctx.HelperContext) (string, error) {
				tx := help.Value("tx").(*pop.Connection)
				u := models.User{}
				erru := tx.Find(&u, id)
				if erru != nil {
					return "http://via.placeholder.com/140x100", nil
				}
				return u.GravatarID.String, nil
			},
			"ownsPost": func(post *models.Post, help hctx.HelperContext) (template.HTML, error) {
				if cu := help.Value("current_user_id"); cu != nil {
					if post.UserID == cu.(uuid.UUID) && help.HasBlock() {
						s, err := help.Block()
						return template.HTML(s), err
					}
				}
				return "", nil
			},
			"byLine": byLineHelper,
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

func byLineHelper(id uuid.UUID, help hctx.HelperContext) (template.HTML, error) {
	tx := help.Value("tx").(*pop.Connection)
	u := models.User{}
	err := tx.Find(&u, id)
	if err != nil {
		return "", err
	}
	if !u.Nickname.Valid {
		return tags.New("span", tags.Options{
			"class": "fab fa-github",
			"body":  "&nbsp;" + u.Name,
		}).HTML(), nil
	}
	return tags.New("a", tags.Options{
		"class":  "fab fa-github",
		"href":   "https://github.com/" + u.Nickname.String,
		"target": "_blank",
		"body":   "&nbsp;" + u.Name,
	}).HTML(), nil
}
