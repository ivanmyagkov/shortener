//	Package for receiving data compression/decompression cookies.
package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
)

type MW struct {
	users interfaces.Users
}

//	Creating a user.
func New(users interfaces.Users) *MW {
	return &MW{
		users: users,
	}
}

//	Intermediate function for validating and creating cookies.
func (M *MW) SessionWithCookies(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("cookie")
		if err != nil {
			uid := utils.CreateID(16)
			cookie := new(http.Cookie)
			cookie.Name = "cookie"
			cookie.Path = "/"
			cookie.Value, _ = M.users.CreateSissionID(uid)
			c.SetCookie(cookie)
			c.Request().AddCookie(cookie)
		} else {
			if _, err := M.users.ReadSessionID(cookie.Value); err != nil {
				uid := utils.CreateID(16)
				cookie := new(http.Cookie)
				cookie.Name = "cookie"
				cookie.Path = "/"
				cookie.Value, _ = M.users.CreateSissionID(uid)
				c.SetCookie(cookie)
				c.Request().AddCookie(cookie)
			}
		}
		return next(c)
	}
}
