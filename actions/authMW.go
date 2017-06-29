package actions

import (
	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// UserExistsMW is an Auth middleware
func UserExistsMW(h buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if App().Env == "test" {
			return h(c)
		}

		userID := c.Session().Get("UserID")

		if userID == nil {
			c.Flash().Set("error", []string{"Need to login first."})
			return c.Redirect(302, "/auth/github")
		}

		user := models.User{}
		tx := c.Value("tx").(*pop.Connection)
		count, err := tx.Where("id = ?", userID).Count(&user)

		if err != nil {
			return errors.WithStack(err)
		}

		if count == 0 {
			c.Flash().Set("error", []string{"Need to login first."})
			return c.Redirect(302, "/auth/github")
		}

		return h(c)
	}
}
