package actions

import (
	"fmt"
	"os"
	//"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/github/callback")),
	)
}

//func AuthCallback(c buffalo.Context) error {
//	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
//	if err != nil {
//		return c.Error(401, err)
//	}
//	// Do something with the user, maybe register them/sign them in
//	// Adding the userID to the session to remember the logged in user
//	session := c.Session()
//	session.Set("userID", user.UserID)
//	err = session.Save()
//	if err != nil {
//		return c.Error(401, err)
//	}
//	// The default value jus renders the data we get by GitHub
//	// return c.Render(200, r.JSON(user))
//
//	// After the user is logged in we add a redirect
//	return c.Redirect(http.StatusMovedPermanently, "/secure")
//}

func AuthCallback(c buffalo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}
	// Do something with the user, maybe register them/sign them in
	return c.Render(200, r.JSON(user))
}
