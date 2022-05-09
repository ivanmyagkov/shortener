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

type Server struct {
	storage interfaces.Storage
	cfg     interfaces.Config
}

func New(storage interfaces.Storage, config interfaces.Config) *Server {
	return &Server{
		storage: storage,
		cfg:     config,
	}
}

func PostURL(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil || len(body) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}
		shortURL := s.cfg.HostName() + "/" + utils.MD5(body)
		s.storage.SetShortURL(shortURL, string(body))

		//db.ShortURL[shortURL] = string(body)

		return c.String(http.StatusCreated, shortURL)
	}
}

func GetURL(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {

		if c.Param("id") == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		shortURL := s.cfg.HostName() + "/" + c.Param("id")

		if s.storage.GetURL(shortURL) == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		c.Response().Header().Set("Location", s.storage.GetURL(shortURL))

		return c.NoContent(http.StatusTemporaryRedirect)
	}
}

func PostJSON(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request struct {
			URL string `json:"url"`
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

		if request.URL == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		response.ShortURL = s.cfg.HostName() + "/" + utils.MD5([]byte(request.URL))
		s.storage.SetShortURL(response.ShortURL, request.URL)

		return c.JSON(http.StatusCreated, response)

	}
}
