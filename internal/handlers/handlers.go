package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	_ "github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
	"github.com/labstack/echo/v4"
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
	_, err = url.ParseRequestURI(string(body))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	shortURL := utils.MD5(body)
	err = s.storage.SetShortURL(shortURL, string(body))
	if err != nil {
		log.Println(err)
	}
	newURL := utils.NewURL(s.cfg.HostName(), shortURL)
	return c.String(http.StatusCreated, newURL)
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

	_, err = url.ParseRequestURI(request.URL)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	response.ShortURL = utils.MD5([]byte(request.URL))
	err = s.storage.SetShortURL(response.ShortURL, request.URL)
	if err != nil {
		log.Println(err)
	}
	response.ShortURL = utils.NewURL(s.cfg.HostName(), response.ShortURL)

	return c.JSON(http.StatusCreated, response)

}
