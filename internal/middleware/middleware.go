//	Package middleware for receiving data compression/decompression cookies.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

//	CompressHandle - data compression function
func CompressHandle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") {
			return next(c)
		}
		gz, err := gzip.NewWriterLevel(c.Response().Writer, gzip.BestSpeed)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		defer gz.Close()
		c.Response().Writer = gzipWriter{ResponseWriter: c.Response().Writer, Writer: gz}
		c.Response().Header().Set("Content-Encoding", "gzip")
		return next(c)
	}
}

//	Decompress - data decompression function
func Decompress(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !strings.Contains(c.Request().Header.Get("Content-Encoding"), "gzip") {
			return next(c)
		}
		gz, err := gzip.NewReader(c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		c.Request().Body = gz
		return next(c)
	}
}
