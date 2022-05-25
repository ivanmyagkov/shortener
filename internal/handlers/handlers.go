package handlers

import (
	"encoding/json"
	"io"
	"log"
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
	user    interfaces.Users
}

func New(storage interfaces.Storage, config interfaces.Config, user interfaces.Users) *Server {
	return &Server{
		storage: storage,
		cfg:     config,
		user:    user,
	}
}

func (s Server) PostURL(c echo.Context) error {
	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	userID, _ := s.user.ReadSessionID(cookie.Value)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	ShortURL, err := s.shortenURL(userID, string(body))
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
	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	log.Println(cookie.Value)
	userID, _ := s.user.ReadSessionID(cookie.Value)

	var request struct {
		URL string `json:"url"`
	}

	var response struct {
		ShortURL string `json:"result"`
	}
	err = json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	//userID, err := s.user.GetUserID(userName)

	response.ShortURL, err = s.shortenURL(userID, request.URL)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusCreated, response)
}

func (s Server) shortenURL(userID, URL string) (string, error) {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return "", err
	}
	shortURL := utils.MD5([]byte(URL))
	err = s.storage.SetShortURL(userID, shortURL, URL)
	if err != nil {
		return "", err
	}
	shortURL = utils.NewURL(s.cfg.HostName(), shortURL)
	return shortURL, nil
}

func (s Server) GetURLsByUserID(c echo.Context) error {

	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	userID, err := s.user.ReadSessionID(cookie.Value)
	if err != nil {
		return c.NoContent(http.StatusNoContent)
	}
	var URLs []interfaces.ModelURL
	if URLs, err = s.storage.GetAllURLsByUserID(userID); err != nil {
		return c.NoContent(http.StatusNoContent)
	}
	var URLArray []interfaces.ModelURL
	var model interfaces.ModelURL
	for _, v := range URLs {
		model.BaseURL = v.BaseURL
		model.ShortURL = utils.NewURL(s.cfg.HostName(), v.ShortURL)
		URLArray = append(URLArray, model)
	}

	return c.JSON(http.StatusOK, URLArray)
}
