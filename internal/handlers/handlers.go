package handlers

import (
	"encoding/json"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	_ "github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

const host = "http://localhost:8080/"

type Server struct {
	storage interfaces.Storage
}

func New(storage interfaces.Storage) *Server {
	return &Server{
		storage: storage,
	}
}

func PostURL(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil || len(body) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}
		shortURL := host + utils.MD5(body)
		s.storage.SetShortUrl(shortURL, string(body))

		//db.ShortURL[shortURL] = string(body)

		return c.String(http.StatusCreated, shortURL)
	}
}

func GetURL(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {

		if c.Param("id") == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		shortURL := host + c.Param("id")

		if s.storage.GetUrl(shortURL) == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		c.Response().Header().Set("Location", s.storage.GetUrl(shortURL))

		return c.NoContent(http.StatusTemporaryRedirect)
	}
}

func PostJSON(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request struct {
			Url string `json:"url"`
		}

		var response struct {
			ShortURL string `json:"result"`
		}

		body, err := io.ReadAll(c.Request().Body)
		if err != nil || len(body) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		err = json.Unmarshal(body, &request)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		if request.Url == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		response.ShortURL = host + utils.MD5([]byte(request.Url))
		s.storage.SetShortUrl(response.ShortURL, request.Url)

		return c.JSON(http.StatusCreated, response)

	}
}
