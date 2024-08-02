package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func initialize() (*App, error) {
	gin.SetMode(gin.ReleaseMode)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := initDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	oauthConfig := initGithubOAuth()

	router := gin.Default()
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	return &App{
		DB:                db,
		GithubOAuthConfig: oauthConfig,
		Router:            router,
	}, nil
}

func initDB() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DB_PATH not set in environment")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initGithubOAuth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func (app *App) Run() error {
	port := os.Getenv("PORT")
	log.Println("Server running on port :" + port)
	return app.Router.Run(":" + port)
}
