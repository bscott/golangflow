package actions

import (
	"net/http"
	"os"

	//"sync"

	"github.com/bscott/golangflow/locales"
	"github.com/bscott/golangflow/models"
	"github.com/bscott/golangflow/public"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/middleware/csrf"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/middleware/paramlogger"
	basicauth "github.com/gobuffalo/mw-basicauth"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/envy"

	// Used for Heroku metrics
	// _ "github.com/heroku/x/hmetrics/onload"
	"github.com/markbates/goth/gothic"
	// newrelic "github.com/newrelic/go-agent"

	"sync"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
)

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	appOnce.Do(func() {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_golangflow_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		if ENV == "development" {
			app.Use(paramlogger.ParameterLogger)
		}

		// NewRelic Integration
		// if ENV == "production" {

		// 	// Ensure SESSION_SECRET is set to secure the session storage in production
		// 	if os.Getenv("SESSION_SECRET") == "" {
		// 		panic("SESSION_SECRET is not set")
		// 	}

		// 	config := newrelic.NewConfig("golangflow", os.Getenv("NEW_RELIC_LICENSE_KEY"))
		// 	config.Enabled = ENV == "production"
		// 	na, _ := newrelic.NewApplication(config)
		// }

		// 	app.Use(func(next buffalo.Handler) buffalo.Handler {
		// 		return func(c buffalo.Context) error {
		// 			req := c.Request()
		// 			txn := na.StartTransaction(req.URL.String(), c.Response(), req)
		// 			ri := c.Value("current_route").(buffalo.RouteInfo)
		// 			txn.AddAttribute("PathName", ri.PathName)
		// 			txn.AddAttribute("RequestID", c.Value("request_id"))
		// 			defer txn.End()
		// 			err := next(c)
		// 			if err != nil {
		// 				txn.NoticeError(err)
		// 				return err
		// 			}
		// 			return nil
		// 		}
		// 	})
		// }

		// Wraps each request in a transaction.
		//   c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))
		// Setup and use translations:
		app.Use(translations())

		app.Use(SetCurrentUser)

		app.Use(T.Middleware())
		app.Use(Authorize)

		app.GET("/", HomeHandler)
		app.GET("/rss", RSSFeed)
		app.GET("/json", JSONFeed)
		app.GET("/privacy", Privacy)
		app.Middleware.Skip(Authorize, HomeHandler, RSSFeed, JSONFeed, Privacy)

		//app.ServeFiles("/assets", assetsBox)

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

		app.ServeFiles("/", http.FS(public.FS())) // serve files from the public directory
	})

	return app
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}
