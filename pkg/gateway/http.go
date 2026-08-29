package gateway

import (
	"database/sql"
	"embed"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/uptime"
)

//go:embed static
var staticFiles embed.FS

func RegisterHandlers(e *echo.Echo, i *do.Injector) {
	e.Renderer = &echo.TemplateRenderer{Template: NewTemplateRenderer()}

	e.StaticFS("/static", echo.MustSubFS(staticFiles, "static"))

	e.GET("/health", func(c *echo.Context) error {
		var health struct {
			Uptime string `json:"uptime"`
			DB     string `json:"db"`
		}

		health.Uptime = uptime.Get().String()

		conn := do.MustInvoke[*sql.DB](i)
		if err := conn.Ping(); err != nil {
			health.DB = "unhealthy"
		} else {
			health.DB = "healthy"
		}

		return c.JSON(http.StatusOK, health)
	})

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index", map[string]any{
			"Title":   "Hello World",
			"Message": "Hello World",
		})
	})
}
