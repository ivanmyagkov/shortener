package handlers

import (
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
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		db    *storage.DB
		value string
		want  want
	}{
		{
			name:  "without param",
			db:    &storage.DB{ShortURL: map[string]string{"http://localhost:8080/f845599b09851789": "https://www.yandex.ru"}},
			value: "",
			want:  want{code: 400},
		},
		{
			name:  "with empty bd",
			db:    &storage.DB{ShortURL: map[string]string{}},
			value: "f845599b09851789",
			want:  want{code: 400},
		},
		{
			name:  "with param",
			db:    &storage.DB{ShortURL: map[string]string{"http://localhost:8080/f845599b09851789": "https://www.yandex.ru"}},
			value: "f845599b09851789",
			want:  want{code: 307},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.value)

			h := GetURL(tt.db)
			if assert.NoError(t, h(c)) {
				require.Equal(t, tt.want.code, rec.Code)
			}
		})
	}
}

func TestPostUrl(t *testing.T) {
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name  string
		db    *storage.DB
		value string
		want  want
	}{
		{
			name:  "body is empty",
			value: "",
			db:    storage.NewDBConn(),
			want:  want{code: 400, body: ""},
		},
		{
			name:  "with body",
			value: "https://www.yandex.ru",
			db:    storage.NewDBConn(),
			want:  want{code: 201, body: "http://localhost:8080/f845599b09851789"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.value))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := PostURL(tt.db)
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
