package actions

import (
	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {

	session := c.Session()
	providerID := session.Get("user_id")
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	posts := &models.Posts{}
	// You can order your list here. Just change
	err := tx.All(posts)
	// to:
	// err := tx.Order("create_at desc").All(posts)
	if err != nil {
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("posts", posts)
	c.Set("session", providerID)
	return c.Render(200, r.HTML("index.html"))
}
