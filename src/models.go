package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type App struct {
	DB                *sql.DB
	GithubOAuthConfig *oauth2.Config
	Router            *gin.Engine
}

type Link struct {
	ID          string
	OriginalURL string
	ShortURL    string
	UserID      string
	CreatedAt   string
}
