package main

import (
	"math/rand"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func isLoggedIn(c *gin.Context) bool {
	_, err := c.Cookie("user_id")
	return err == nil
}

func getUserID(c *gin.Context) string {
	userID, _ := c.Cookie("user_id")
	return userID
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func generateShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func (app *App) getUserLinks(userID string) ([]Link, error) {
	rows, err := app.DB.Query("SELECT id, original_url, short_url FROM links WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		err := rows.Scan(&link.ID, &link.OriginalURL, &link.ShortURL)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

func (app *App) createTable() error {
	_, err := app.DB.Exec(`CREATE TABLE IF NOT EXISTS links (
		id TEXT PRIMARY KEY NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		short_url TEXT NOT NULL UNIQUE,
		user_id TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}
