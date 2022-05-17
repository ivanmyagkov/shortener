package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	_ "github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
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

func (s Server) PostURL(c echo.Context) error {

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	ShortURL, err := s.shortenURL(string(body))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	return c.String(http.StatusCreated, ShortURL)
}

func (s Server) GetURL(c echo.Context) error {
	if c.Param("id") == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	shortURL := c.Param("id")

	su, err := s.storage.GetURL(shortURL)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	c.Response().Header().Set("Location", su)
	return c.NoContent(http.StatusTemporaryRedirect)
}

func (s Server) PostJSON(c echo.Context) error {
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
	response.ShortURL, err = s.shortenURL(request.URL)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusCreated, response)

}

func (s Server) shortenURL(URL string) (string, error) {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return "", err
	}
	shortURL := utils.MD5([]byte(URL))
	err = s.storage.SetShortURL(shortURL, URL)
	if err != nil {
		return "", err
	}
	shortURL = utils.NewURL(s.cfg.HostName(), shortURL)
	return shortURL, nil
}
