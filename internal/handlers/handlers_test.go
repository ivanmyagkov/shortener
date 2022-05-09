package handlers

import (
	"encoding/json"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetUrl(t *testing.T) {
	type args struct {
		db       *storage.DB
		url      string
		shortUrl string
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
				url:      "https://www.yandex.ru",
				shortUrl: "http://localhost:8080/f845599b09851789",
			},
			value: "",
			want:  want{code: 400},
		},
		{
			name: "with empty bd",
			args: args{
				db: storage.NewDBConn(),
			},
			value: "f845599b09851789",
			want:  want{code: 400},
		},
		{
			name: "with param",
			args: args{
				db:       storage.NewDBConn(),
				url:      "https://www.yandex.ru",
				shortUrl: "http://localhost:8080/f845599b09851789",
			},
			value: "f845599b09851789",
			want:  want{code: 307},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			s := New(tt.args.db)
			s.storage.SetShortUrl(tt.args.shortUrl, tt.args.url)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.value)

			h := GetURL(s)
			if assert.NoError(t, h(c)) {
				require.Equal(t, tt.want.code, rec.Code)
			}
		})
	}
}

func TestPostUrl(t *testing.T) {
	type args struct {
		db *storage.DB
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
				db: storage.NewDBConn(),
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "with body",
			value: "https://www.yandex.ru",
			args: args{
				db: storage.NewDBConn(),
			},
			want: want{code: 201, body: "http://localhost:8080/f845599b09851789"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			s := New(tt.args.db)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := PostURL(s)
			if assert.NoError(t, h(c)) {
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
		db *storage.DB
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
				db: storage.NewDBConn(),
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "body is wrong",
			value: `{"url": ""}`,
			args: args{
				db: storage.NewDBConn(),
			},
			want: want{code: 400, body: ""},
		},
		{
			name:  "with body",
			value: `{"url": "https://www.yandex.ru"}`,
			args: args{
				db: storage.NewDBConn(),
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
			s := New(tt.args.db)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			h := PostJSON(s)
			if assert.NoError(t, h(c)) {
				require.Equal(t, tt.want.code, rec.Code)
				body, err := io.ReadAll(rec.Body)
				err = json.Unmarshal(body, &response)

				if err != nil {
					require.Errorf(t, err, "can't read body")
				}

				require.Equal(t, tt.want.body, response.ShortURL)
			}
		})
	}
}
