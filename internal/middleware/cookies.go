package middleware

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
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
		cookie, err := c.Request().Cookie("cookie")
		if err != nil {
			log.Println(2)
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
