package handlers

import (
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

const host = "http://localhost:8080/"

func PostURL(db *storage.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil || len(body) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}
		shortURL := host + utils.MD5(body)

		db.ShortURL[shortURL] = string(body)

		return c.String(http.StatusCreated, shortURL)
	}
}

func GetURL(db *storage.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		if c.Param("id") == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		shortURL := host + c.Param("id")

		if db.ShortURL[shortURL] == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		c.Response().Header().Set("Location", db.ShortURL[shortURL])

		return c.NoContent(http.StatusTemporaryRedirect)
	}
}
