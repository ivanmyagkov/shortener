package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
)

func TestGetUrl(t *testing.T) {
	type args struct {
		db       *storage.DB
		cfg      *config.Config
		usr      *storage.DBUsers
		URL      string
		shortURL string
		cookie   string
	}
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		args  args
		value string
		want  want
	}{
		{
			name: "without param",
			args: args{
				db:       storage.NewDBConn(),
				usr:      storage.New(),
				cfg:      config.NewConfig(":8080", "http://localhost:8080/", "", ""),
				URL:      "https://www.yandex.ru",
				shortURL: "http://localhost:8080/f845599b09851789",
				cookie:   "a07a35a622236b60753719fbc9a9ff0c",
			},
			value: "",
			want:  want{code: 400},
		},
		{
			name: "with empty bd",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080/", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			value: "f845599b09851789",
			want:  want{code: 400},
		},
		{
			name: "with param",
			args: args{
				db:       storage.NewDBConn(),
				usr:      storage.New(),
				cfg:      config.NewConfig(":8080", "http://localhost:8080", "", ""),
				URL:      "https://www.yandex.ru",
				shortURL: "f845599b09851789",
				cookie:   "a07a35a622236b60753719fbc9a9ff0c",
			},
			value: "f845599b09851789",
			want:  want{code: 307},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			s := New(tt.args.db, tt.args.cfg, tt.args.usr)
			ID, _ := tt.args.usr.ReadSessionID(tt.args.cookie)
			err := s.storage.SetShortURL(ID, tt.args.shortURL, tt.args.URL)
			if err != nil {
				log.Println(err)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.value)

			h := s.GetURL(c)
			if assert.NoError(t, h) {
				require.Equal(t, tt.want.code, rec.Code)
			}
		})
	}
}

func TestPostUrl(t *testing.T) {
	type args struct {
		db     *storage.DB
		cfg    *config.Config
		usr    *storage.DBUsers
		cookie string
	}
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name  string
		args  args
		value string
		want  want
	}{
		{
			name:  "body is empty",
			value: "",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "with body",
			value: "https://www.yandex.ru",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 201, body: "http://localhost:8080/f845599b09851789"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			s := New(tt.args.db, tt.args.cfg, tt.args.usr)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			cookies := new(http.Cookie)
			cookies.Name = "cookie"
			cookies.Path = "/"
			cookies.Value = tt.args.cookie
			c.SetCookie(cookies)
			c.Request().AddCookie(cookies)
			h := s.PostURL(c)
			if assert.NoError(t, h) {
				require.Equal(t, tt.want.code, rec.Code)
				body, err := io.ReadAll(rec.Body)
				if err != nil {
					require.Errorf(t, err, "can't read body")
				}
				require.Equal(t, tt.want.body, string(body))
			}
		})
	}
}

func TestPostJSON(t *testing.T) {

	type args struct {
		db     *storage.DB
		cfg    *config.Config
		usr    *storage.DBUsers
		cookie string
	}
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name  string
		args  args
		value string
		want  want
	}{
		{
			name:  "body is empty",
			value: "",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "body is wrong",
			value: `{"url": ""}`,
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "with body",
			value: `{"url" : "https://www.yandex.ru"}`,
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 201, body: "http://localhost:8080/f845599b09851789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response struct {
				ShortURL string `json:"result"`
			}
			e := echo.New()
			s := New(tt.args.db, tt.args.cfg, tt.args.usr)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			cookies := new(http.Cookie)
			cookies.Name = "cookie"
			cookies.Path = "/"
			cookies.Value = tt.args.cookie
			c.SetCookie(cookies)
			c.Request().AddCookie(cookies)
			h := s.PostJSON(c)
			if assert.NoError(t, h) {
				require.Equal(t, tt.want.code, rec.Code)
				body, err := io.ReadAll(rec.Body)
				if err != nil {
					require.Errorf(t, err, "can't read body")
				}
				err = json.Unmarshal(body, &response)

				if err != nil {
					require.Errorf(t, err, "can't read body")
				}
				require.Equal(t, tt.want.body, response.ShortURL)
			}
		})
	}
}

func TestServer_GetURLsByUserID(t *testing.T) {
	type args struct {
		db     *storage.DB
		cfg    *config.Config
		usr    *storage.DBUsers
		cookie string
	}
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		args  args
		value string
		want  want
	}{
		{
			name:  "body is empty",
			value: "",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", ""),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 204},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			e := echo.New()
			s := New(tt.args.db, tt.args.cfg, tt.args.usr)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			cookies := new(http.Cookie)
			cookies.Name = "cookie"
			cookies.Path = "/"
			cookies.Value = tt.args.cookie
			c.SetCookie(cookies)
			c.Request().AddCookie(cookies)

			h := s.GetURLsByUserID(c)
			if assert.NoError(t, h) {
				require.Equal(t, tt.want.code, rec.Code)
			}
		})
	}
}
