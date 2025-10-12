package actions

import (
	"embed"
	"net/http"
	"os"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	basicauth "github.com/gobuffalo/mw-basicauth"
	forcessl "github.com/gobuffalo/mw-forcessl"
	i18n "github.com/gobuffalo/mw-i18n/v2"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/envy"

	// Used for Heroku metrics
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/markbates/goth/gothic"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// T i18n Translator
var T *i18n.Translator

var localesFS embed.FS

// SetLocalesFS sets the embedded filesystem for locales
func SetLocalesFS(locales embed.FS) {
	localesFS = locales
}

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
			app.Use(paramlogger.ParameterLogger)
		}

		// NewRelic Integration
		if ENV == "production" {

			// Ensure SESSION_SECRET is set to secure the session storage in production
			if os.Getenv("SESSION_SECRET") == "" {
				panic("SESSION_SECRET is not set")
			}

			na, _ := newrelic.NewApplication(
				newrelic.ConfigAppName("golangflow"),
				newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
				newrelic.ConfigEnabled(ENV == "production"),
			)

			app.Use(func(next buffalo.Handler) buffalo.Handler {
				return func(c buffalo.Context) error {
					req := c.Request()
					txn := na.StartTransaction(req.URL.String())
					defer txn.End()

					ri := c.Value("current_route").(buffalo.RouteInfo)
					txn.AddAttribute("PathName", ri.PathName)
					txn.AddAttribute("RequestID", c.Value("request_id"))

					err := next(c)
					if err != nil {
						txn.NoticeError(err)
						return err
					}
					return nil
				}
			})
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		//app.Use(middleware.CSRF)

		// Pop Trans
		app.Use(popmw.Transaction(models.DB))

		app.Use(SetCurrentUser)

		// Automatically redirect to SSL version
		app.Use(forceSSL())
		// Setup and use translations:
		var err error
		if T, err = i18n.New(localesFS, "en"); err != nil {
			app.Stop(err)
		}
		app.Use(T.Middleware())
		app.Use(Authorize)

		app.GET("/", HomeHandler)
		app.GET("/rss", RSSFeed)
		app.GET("/json", JSONFeed)
		app.GET("/privacy", Privacy)
		app.Middleware.Skip(Authorize, HomeHandler, RSSFeed, JSONFeed, Privacy)

		app.ServeFiles("/assets", http.FS(assetsFS))

		// Auth Group
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

func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
