package main

import (
	"io/fs"
	"net/http"
)

func (app *App) SetupRoutes() {
	staticContent, _ := fs.Sub(staticFS, "static")
	app.Router.StaticFS("/static", http.FS(staticContent))

	app.Router.GET("/", app.homePage)
	app.Router.POST("/", app.shortenURL)

	app.Router.GET("/login", app.loginGithub)
	app.Router.GET("/callback", app.githubCallback)
	app.Router.GET("/logout", app.logout)

	app.Router.GET("/dashboard", app.dashboard)
	app.Router.DELETE("/dashboard", app.deleteLink)
	app.Router.PUT("/dashboard", app.updateLink)
	app.Router.GET("/:shortURL", app.redirectToOriginal)

}
