//	Server's handlers
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

// Server struct
type Server struct {
	storage  interfaces.Storage
	cfg      interfaces.Config
	user     interfaces.Users
	inWorker interfaces.InWorker
}

//	Server constructor.
func New(storage interfaces.Storage, config interfaces.Config, user interfaces.Users, inWorker interfaces.InWorker) *Server {
	return &Server{
		storage:  storage,
		cfg:      config,
		user:     user,
		inWorker: inWorker,
	}
}

//	Post request handler.
//	Adding a link to an abbreviation.
//	We get an abbreviated link.
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

//	GET request handler.
//	We get original link.
func (s Server) GetURL(c echo.Context) error {
	if c.Param("id") == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	shortURL := c.Param("id")
	su, err := s.storage.GetURL(shortURL)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return c.NoContent(http.StatusBadRequest)
		} else if errors.Is(err, interfaces.ErrWasDeleted) {
			return c.NoContent(http.StatusGone)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	c.Response().Header().Set("Location", su)
	return c.NoContent(http.StatusTemporaryRedirect)
}

//	Post request handler.
//	Adding a link to an abbreviation.
//	Passing the link in the form of json.
//	We get an abbreviated link.
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

// Auxiliary link shortening functionю
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

//	Ping handler.
func (s Server) GetPing(c echo.Context) error {
	if err := s.storage.Ping(); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

//	Get request handler.
//	Getting all the user's links.
func (s Server) GetURLsByUserID(c echo.Context) error {

	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userID, _ := s.user.ReadSessionID(cookie.Value)

	URLs, err := s.storage.GetAllURLsByUserID(userID)
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

//	Post request handler.
//	Adding a links to an abbreviation.
//	Passing the link in the form array  of json.
//	We get an array of abbreviated link.
func (s Server) PostBatch(c echo.Context) error {
	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	userID, _ := s.user.ReadSessionID(cookie.Value)
	batchReq := make([]interfaces.BatchRequest, 1000)
	batchArr := make([]interfaces.BatchResponse, 1000)
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

//	DELETE request handler.
//	delete user links.
func (s Server) DelURLsBATCH(c echo.Context) error {
	cookie, err := c.Request().Cookie("cookie")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	userID, _ := s.user.ReadSessionID(cookie.Value)
	var model interfaces.Task
	model.ID = userID

	body, err := io.ReadAll(c.Request().Body)
	if err != nil || len(body) == 0 {
		return c.NoContent(http.StatusBadRequest)
	}
	deleteURLs := make([]string, 1000)
	err = json.Unmarshal(body, &deleteURLs)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	for _, deleteURL := range deleteURLs {
		model.ShortURL = deleteURL
		s.inWorker.Do(model)
	}

	return c.NoContent(http.StatusAccepted)
}
