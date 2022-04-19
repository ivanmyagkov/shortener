package handlers

import (
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

const host = "http://localhost:8080/"

func PostUrl(url *storage.Db) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil || len(body) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}
		shortUrl := host + utils.MD5(body)

		url.ShortUrl[shortUrl] = string(body)

		return c.String(http.StatusCreated, shortUrl)
	}
}

func GetUrl(db *storage.Db) echo.HandlerFunc {
	return func(c echo.Context) error {

		if c.Param("id") == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		shortUrl := host + c.Param("id")

		if db.ShortUrl[shortUrl] == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		c.Response().Header().Set("Location", db.ShortUrl[shortUrl])

		return c.NoContent(http.StatusTemporaryRedirect)
	}
}
