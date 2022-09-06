package handlers

import (
	"context"
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
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/workerpool"
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
				cfg:      config.NewConfig(":8080", "http://localhost:8080/", "", "", false),
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
				cfg:    config.NewConfig(":8080", "http://localhost:8080/", "", "", false),
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
				cfg:      config.NewConfig(":8080", "http://localhost:8080", "", "", false),
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
			recordCh := make(chan interfaces.Task, 50)
			doneCh := make(chan struct{})

			inWorker := workerpool.NewInputWorker(recordCh, doneCh, context.Background())
			s := New(tt.args.db, tt.args.cfg, tt.args.usr, inWorker)
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
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 500, body: ""},
		},
		{
			name:  "cookie is wrong",
			value: "https://www.yandex.ru",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "with body",
			value: "https://www.yandex.ru/1",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 201, body: "http://localhost:8080/34a7a94a3c659110"},
		},
		{
			name:  "conflict",
			value: "https://www.yandex.ru/1",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 409, body: "http://localhost:8080/34a7a94a3c659110"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordCh := make(chan interfaces.Task, 50)
			doneCh := make(chan struct{})
			db, _ := storage.NewDB(tt.args.cfg.DatabasePath)
			inWorker := workerpool.NewInputWorker(recordCh, doneCh, context.Background())
			e := echo.New()
			s := New(db, tt.args.cfg, tt.args.usr, inWorker)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			cookies := new(http.Cookie)
			if tt.name == "cookie is wrong" {
				cookies.Name = "cookiee"
			} else {
				cookies.Name = "cookie"
			}

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
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
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
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 500, body: ""},
		},
		{
			name:  "cookie is wrong",
			value: `{"url" : "https://www.yandex.ru"}`,
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
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
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 201, body: "http://localhost:8080/f845599b09851789"},
		},
		{
			name:  "conflict",
			value: `{"url" : "https://www.yandex.ru"}`,
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: http.StatusConflict, body: "http://localhost:8080/f845599b09851789"},
		},
	}

	for _, tt := range tests {
		e := echo.New()
		recordCh := make(chan interfaces.Task, 50)
		doneCh := make(chan struct{})

		inWorker := workerpool.NewInputWorker(recordCh, doneCh, context.Background())
		t.Run(tt.name, func(t *testing.T) {
			var response struct {
				ShortURL string `json:"result"`
			}

			db, _ := storage.NewDB(tt.args.cfg.DatabasePath)
			s := New(db, tt.args.cfg, tt.args.usr, inWorker)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			cookies := new(http.Cookie)
			if tt.name == "cookie is wrong" {
				cookies.Name = "cookiee"
			} else {
				cookies.Name = "cookie"
			}
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
		//{
		//	name:  "with body",
		//	value: "",
		//	args: args{
		//		db:     storage.NewDBConn(),
		//		usr:    storage.New(),
		//		cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
		//		cookie: "33a578d973226bffb1ecd6ba6b9f179c",
		//	},
		//	want: want{code: 200},
		//},
		{
			name:  "body is empty",
			value: "",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 204},
		},
		{
			name:  "cookie is wrong",
			value: "",
			args: args{
				db:     storage.NewDBConn(),
				usr:    storage.New(),
				cfg:    config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
				cookie: "a07a35a622236b60753719fbc9a9ff0c",
			},
			want: want{code: 400},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			e := echo.New()
			recordCh := make(chan interfaces.Task, 50)
			doneCh := make(chan struct{})

			inWorker := workerpool.NewInputWorker(recordCh, doneCh, context.Background())
			s := New(tt.args.db, tt.args.cfg, tt.args.usr, inWorker)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			cookies := new(http.Cookie)
			if tt.name == "cookie is wrong" {
				cookies.Name = "cookiee"
			} else {
				cookies.Name = "cookie"
			}
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

func TestServer_GetPing(t *testing.T) {
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
			name:  "status ok",
			value: "",
			args: args{
				usr: storage.New(),
				cfg: config.NewConfig(":8080", "http://localhost:8080", "", "postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable", false),
			},
			want: want{code: 200},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			recordCh := make(chan interfaces.Task, 50)
			doneCh := make(chan struct{})

			inWorker := workerpool.NewInputWorker(recordCh, doneCh, context.Background())
			db, _ := storage.NewDB(tt.args.cfg.DatabasePath)
			s := New(db, tt.args.cfg, tt.args.usr, inWorker)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := s.GetPing(c)
			if assert.NoError(t, h) {
				require.Equal(t, tt.want.code, rec.Code)
			}
		})
	}
}
