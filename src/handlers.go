package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (app *App) homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"loggedIn": isLoggedIn(c),
	})
}

func (app *App) shortenURL(c *gin.Context) {
	originalURL := c.PostForm("url")
	customName := c.PostForm("custom_name")

	if !isValidURL(originalURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	shortURL := customName
	if shortURL == "" {
		shortURL = generateShortURL()
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

func (app *App) logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (app *App) dashboard(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	links, err := app.getUserLinks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
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
