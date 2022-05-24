package middleware

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type MW struct {
	users interfaces.Users
}

func New(users interfaces.Users) *MW {
	return &MW{
		users: users,
	}
}

func (M *MW) SessionWithCookies(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var ok bool
		ok = true
		cookie, err := c.Cookie("cookie")
		if err != nil {
			cookie := new(http.Cookie)
			cookie.Name = "cookie"
			cookie.Path = "/"
			cookie.Value, _ = M.users.CreateSissionID()
			c.SetCookie(cookie)
			c.Request().AddCookie(cookie)
		} else {

			if _, err := M.users.ReadSessionID(cookie.Value); err != nil {
				cookie := new(http.Cookie)
				cookie.Name = "cookie"
				cookie.Path = "/"
				cookie.Value, _ = M.users.CreateSissionID()
				c.SetCookie(cookie)
				c.Request().AddCookie(cookie)
			}
		}
		log.Println(ok)
		return next(c)
	}
}
