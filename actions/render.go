package actions

import (
	"html/template"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/plush"
	"github.com/markbates/pop"
	uuid "github.com/satori/go.uuid"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Box containing all of the templates:
		TemplatesBox: packr.NewBox("../templates"),

		// Add template helpers here:
		Helpers: render.Helpers{
			"getAvatar": func(id uuid.UUID, help plush.HelperContext) (string, error) {
				tx := help.Value("tx").(*pop.Connection)
				p := models.Post{}
				err := tx.Find(&p, id)
				if err != nil {
					return "", err
				}
				u := models.User{}
				erru := tx.Find(&u, p.UserID)
				if erru != nil {
					return "http://via.placeholder.com/140x100", nil
				}
				return u.GravatarID.String, nil
			},
			"getUser": func(id uuid.UUID, help plush.HelperContext) (string, error) {
				tx := help.Value("tx").(*pop.Connection)
				p := models.Post{}
				err := tx.Find(&p, id)
				if err != nil {
					return " ", err
				}
				u := models.User{}
				erru := tx.Find(&u, p.UserID)
				if erru != nil {
					return "User Not Found", erru
				}
				return u.Name, nil
			},
			"ownsPost": func(post *models.Post, help plush.HelperContext) (template.HTML, error) {
				if cu := help.Value("current_user_id"); cu != nil {
					if post.UserID == cu.(uuid.UUID) && help.HasBlock() {
						s, err := help.Block()
						return template.HTML(s), err
					}
				}
				return "", nil
			},
		},
	})
}
