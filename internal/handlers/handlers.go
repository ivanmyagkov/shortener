package handlers

import (
	"encoding/json"
	"errors"
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
		if errors.Is(err, interfaces.ErrAlreadyExists) {
			return c.String(http.StatusConflict, ShortURL)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}
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
		if errors.Is(err, interfaces.ErrNotFound) {
			return c.NoContent(http.StatusBadRequest)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	c.Response().Header().Set("Location", su)
	return c.NoContent(http.StatusTemporaryRedirect)
}

func (s Server) PostJSON(c echo.Context) error {
	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	userID, _ := s.user.ReadSessionID(cookie.Value)

	var request struct {
		URL string `json:"url"`
	}

	var response struct {
		Result string `json:"result"`
	}
	err = json.NewDecoder(c.Request().Body).Decode(&request)

	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	response.Result, err = s.shortenURL(userID, request.URL)

	if err != nil {
		if errors.Is(err, interfaces.ErrAlreadyExists) {
			return c.JSON(http.StatusConflict, response)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}
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
		if errors.Is(err, interfaces.ErrAlreadyExists) {
			shortURL = utils.NewURL(s.cfg.HostName(), shortURL)
			return shortURL, interfaces.ErrAlreadyExists
		} else {
			return "", err
		}
	}

	shortURL = utils.NewURL(s.cfg.HostName(), shortURL)
	return shortURL, nil
}

func (s Server) GetPing(c echo.Context) error {
	if err := s.storage.Ping(); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s Server) GetURLsByUserID(c echo.Context) error {

	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userID, _ := s.user.ReadSessionID(cookie.Value)

	var URLs []interfaces.ModelURL
	URLs, err = s.storage.GetAllURLsByUserID(userID)
	if err != nil {
		return c.NoContent(http.StatusNoContent)
	}
	URLArray := make([]interfaces.ModelURL, 0, len(URLs))

	for _, v := range URLs {
		var model interfaces.ModelURL
		model.BaseURL = v.BaseURL
		model.ShortURL = utils.NewURL(s.cfg.HostName(), v.ShortURL)
		URLArray = append(URLArray, model)
	}

	return c.JSON(http.StatusOK, URLArray)
}

func (s Server) PostBatch(c echo.Context) error {
	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	userID, _ := s.user.ReadSessionID(cookie.Value)
	var batchReq []interfaces.BatchRequest
	var batchArr []interfaces.BatchResponse
	err = json.NewDecoder(c.Request().Body).Decode(&batchReq)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	for _, batch := range batchReq {
		var batchRes interfaces.BatchResponse
		batchRes.CorrelationID = batch.CorrelationID
		batchRes.ShortURL, err = s.shortenURL(userID, batch.OriginalURL)
		if err != nil {
			if errors.Is(err, interfaces.ErrAlreadyExists) {
				return c.NoContent(http.StatusBadRequest)
			}
			return err
		}

		batchArr = append(batchArr, batchRes)
	}
	return c.JSON(http.StatusCreated, batchArr)
}
