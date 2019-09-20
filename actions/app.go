package actions

import (
	"encoding/json"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	tokenauth "github.com/gobuffalo/mw-tokenauth"
	"github.com/unrolled/secure"

	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo-pop/pop/popmw"
	contenttype "github.com/gobuffalo/mw-contenttype"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

func errorHandler() buffalo.ErrorHandler {
	return func(status int, err error, c buffalo.Context) error {
		c.Logger().Error(err)
		c.Response().WriteHeader(status)
		msg := fmt.Sprintf("%+v", err.Error())

		return json.NewEncoder(c.Response()).Encode(map[string]interface{}{
			"error":  msg,
			"status": status,
		})

	}

}

// App is the microsocial API's starting point.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_microsocial_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())
		app.Use(paramlogger.ParameterLogger)
		app.Use(contenttype.Set("application/json"))
		app.Use(popmw.Transaction(models.DB))

		app.GET("/fake_auth/{login}", LoginAsUser)

		// JWT authentication middleware
		auth_mw := tokenAuth()

		users := app.Group("/users")
		users.Use(auth_mw)
		users.GET("/", UsersList)
		users.POST("/", UsersCreate)
		users.GET("/{user_id}", UsersShow)
		users.PUT("/{user_id}", UsersUpdate)
		users.DELETE("/{user_id}", UsersDestroy)
		users.POST("/{user_id}/friend_request", FriendRequestsCreate)
		users.GET("/{user_id}/unfriend", FriendshipsDestroy)
		users.Middleware.Skip(auth_mw, UsersList, UsersCreate)

		frs := app.Group("/friend_requests")
		frs.Use(auth_mw)
		frs.GET("/{request_id}/accept", FriendRequestsAccept)
		frs.GET("/{request_id}/decline", FriendRequestsDecline)

		reports := app.Group("/reports")
		reports.Use(auth_mw)
		reports.POST("/", ReportsCreate)
		reports.GET("/", ReportsList)

		app.ErrorHandlers[400] = errorHandler()
		app.ErrorHandlers[401] = errorHandler()
		app.ErrorHandlers[403] = errorHandler()
		app.ErrorHandlers[404] = errorHandler()
		app.ErrorHandlers[409] = errorHandler()
		app.ErrorHandlers[500] = errorHandler()
	}

	return app
}

func tokenAuth() buffalo.MiddlewareFunc {
	return tokenauth.New(tokenauth.Options{})
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
