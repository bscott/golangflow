package actions

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/pop/nulls"
	//"github.com/pkg/errors"

	"github.com/markbates/pop"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/github/callback")),
	)
}

// TODO: Refactor this whole function...
func AuthCallback(c buffalo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}
	// Do something with the user, maybe register them/sign them in
	// Adding the userID to the session to remember the logged in user

	u := models.User{
		Name:           user.Name,
		Email:          nulls.NewString(user.Email),
		ProviderUserid: user.UserID,
		GravatarID:     nulls.NewString(user.AvatarURL),
		Provider:       user.Provider,
	}

	// check if user already exists
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	eu := models.User{}
	var exists bool

	// To find the User the parameter user_id is used.
	query := tx.Where("provider = ?", user.Provider).Where("provider_userid = ?", user.UserID)
	err = query.First(&eu)

	if err != nil {
		exists = bool(false)
	}

	//exists, err := query.Exists(&models.User{})
	//
	//if err != nil {
	//	return c.Error(401, err)
	//}

	if exists == false && eu.Provider != user.Provider && eu.ProviderUserid != user.UserID {
		models.DB.Create(&u)

		// Build Session
		session := c.Session()
		session.Set("userID", user.UserID)
		err = session.Save()
		if err != nil {
			return c.Error(401, err)
		}
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	// Build Session
	session := c.Session()
	session.Set("userID", user.UserID)
	err = session.Save()
	if err != nil {
		return c.Error(401, err)
	}

	// The default value jus renders the data we get by GitHub
	// return c.Render(200, r.JSON(user))

	// After the user is logged in we add a redirect
	return c.Redirect(http.StatusMovedPermanently, "/")
}

//func AuthCallback(c buffalo.Context) error {
//	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
//	if err != nil {
//		return c.Error(401, err)
//	}
//
//
//	// Test User Ingestion
//
//
//
//
//
//
//
//	// Do something with the user, maybe register them/sign them in
//	return c.Render(200, r.JSON(user))
//}
