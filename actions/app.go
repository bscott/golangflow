package actions

import (
	"log"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/basicauth"
	"github.com/gobuffalo/buffalo/middleware/i18n"

	"github.com/bscott/golangflow/models"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"

	"github.com/markbates/goth/gothic"
	"github.com/newrelic/go-agent"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// T i18n Translator
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_golangflow_session",
		})
		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		// NewRelic Integration
		config := newrelic.NewConfig("golangflow", os.Getenv("NEW_RELIC_LICENSE_KEY"))
		config.Enabled = ENV == "production"
		na, _ := newrelic.NewApplication(config)

		app.Use(func(next buffalo.Handler) buffalo.Handler {
			return func(c buffalo.Context) error {
				req := c.Request()
				txn := na.StartTransaction(req.URL.String(), c.Response(), req)
				ri := c.Value("current_route").(buffalo.RouteInfo)
				txn.AddAttribute("PathName", ri.PathName)
				txn.AddAttribute("RequestID", c.Value("request_id"))
				defer txn.End()
				err := next(c)
				if err != nil {
					txn.NoticeError(err)
					return err
				}
				return nil
			}
		})

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		//app.Use(middleware.CSRF)

		app.Use(middleware.PopTransaction(models.DB))
		app.Use(SetCurrentUser)
		// Setup and use translations:
		var err error
		T, err = i18n.New(packr.NewBox("../locales"), "en-US")
		if err != nil {
			log.Fatal(err)
		}
		app.Use(T.Middleware())
		app.Use(Authorize)

		app.GET("/", HomeHandler)
		app.GET("/rss", RSSFeed)
		app.Middleware.Skip(Authorize, HomeHandler, RSSFeed)

		app.ServeFiles("/assets", packr.NewBox("../public/assets"))

		auth := app.Group("/auth")
		gothwap := buffalo.WrapHandlerFunc(gothic.BeginAuthHandler)
		auth.GET("/{provider}", gothwap)
		auth.GET("/{provider}/callback", AuthCallback)
		auth.DELETE("", AuthDestroy)
		auth.Middleware.Skip(Authorize, AuthCallback, gothwap)

		g := app.Resource("/users", UsersResource{&buffalo.BaseResource{}})
		g.Use(basicauth.Middleware(func(c buffalo.Context, u string, p string) (bool, error) {
			return (u == os.Getenv("ADMIN_USER") && p == os.Getenv("ADMIN_PASS")), nil
		}))

		pr := PostsResource{&buffalo.BaseResource{}}
		pg := app.Resource("/posts", pr)
		pg.Middleware.Skip(Authorize, pr.Show)
	}

	return app
}
