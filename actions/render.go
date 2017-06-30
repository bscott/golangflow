package actions

import (
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
		},
	})
}
