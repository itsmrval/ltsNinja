package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Link struct {
	ID          string
	OriginalURL string
	ShortURL    string
	UserID      string
}

type App struct {
	DB                *sql.DB
	GithubOAuthConfig *oauth2.Config
	Router            *gin.Engine
}

func NewApp() (*App, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := initDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	oauthConfig := initGithubOAuth()

	router := gin.Default()

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

func (app *App) SetupRoutes() {
	app.Router.LoadHTMLGlob("templates/*")
	app.Router.Static("/static", "./static")
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

func (app *App) Run() error {
	port := os.Getenv("PORT")
	return app.Router.Run(":" + port)
}

func (app *App) createTable() error {
	_, err := app.DB.Exec(`CREATE TABLE IF NOT EXISTS links (
		id TEXT PRIMARY KEY NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		short_url TEXT NOT NULL UNIQUE,
		user_id TEXT
	)`)
	return err
}

func (app *App) logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (app *App) homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"loggedIn": isLoggedIn(c),
	})
}

func (app *App) shortenURL(c *gin.Context) {
	originalURL := c.PostForm("url")
	customName := c.PostForm("custom_name")

	shortURL := customName
	if shortURL == "" {
		shortURL = uuid.New().String()[:8]
	}

	userID := getUserID(c)

	id := uuid.New().String()

	_, err := app.DB.Exec("INSERT INTO links (id, original_url, short_url, user_id) VALUES (?, ?, ?, ?)",
		id, originalURL, shortURL, userID)
	if err != nil {
		log.Printf("Error inserting link: %v", err)
		if err.Error() == "UNIQUE constraint failed: links.short_url" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Short URL already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shortURL": shortURL})
}

func (app *App) loginGithub(c *gin.Context) {
	url := app.GithubOAuthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusFound, url)
}

func (app *App) githubCallback(c *gin.Context) {
	code := c.Query("code")
	token, err := app.GithubOAuthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := app.GithubOAuthConfig.Client(c, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var githubUser struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(body, &githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	userID := fmt.Sprintf("%d", githubUser.ID)
	c.SetCookie("user_id", userID, 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (app *App) dashboard(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	rows, err := app.DB.Query("SELECT id, original_url, short_url FROM links WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		err := rows.Scan(&link.ID, &link.OriginalURL, &link.ShortURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		links = append(links, link)
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"links": links, "userId": userID})
}

func (app *App) deleteLink(c *gin.Context) {
	var payload struct {
		ID string `json:"id"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userID := getUserID(c)

	_, err := app.DB.Exec("DELETE FROM links WHERE id = ? AND user_id = ?", payload.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "id": payload.ID})
}

func (app *App) updateLink(c *gin.Context) {
	var payload struct {
		ID      string `json:"id"`
		NewName string `json:"new_name"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userID := getUserID(c)

	_, err := app.DB.Exec("UPDATE links SET short_url = ? WHERE id = ? AND user_id = ?", payload.NewName, payload.ID, userID)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: links.short_url" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Short URL already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (app *App) redirectToOriginal(c *gin.Context) {
	shortURL := c.Param("shortURL")

	var originalURL string
	err := app.DB.QueryRow("SELECT original_url FROM links WHERE short_url = ?", shortURL).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "Short URL not found")
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

func isLoggedIn(c *gin.Context) bool {
	_, err := c.Cookie("user_id")
	return err == nil
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Cookie("user_id")
	return userID
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := app.createTable(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	app.SetupRoutes()

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
