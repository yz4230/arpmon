package gateway

import (
	"embed"
	"html/template"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/labstack/echo/v5"
)

//go:embed templates
var templateFiles embed.FS

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func formatDateTime(value time.Time) string {
	return value.Format(time.DateTime)
}

func NewTemplateRenderer() *TemplateRenderer {
	const ext = ".html.tmpl"
	base := template.Must(template.New("base").Funcs(template.FuncMap{
		"formatDateTime": formatDateTime,
	}).ParseFS(templateFiles, path.Join("templates", "base"+ext)))

	mustParse := func(base *template.Template, page string) *template.Template {
		tmpl, err := base.Clone()
		if err != nil {
			panic(err)
		}
		tmpl, err = tmpl.ParseFS(templateFiles, path.Join("templates", "pages", page+ext))
		if err != nil {
			panic(err)
		}
		return tmpl
	}

	return &TemplateRenderer{
		templates: map[string]*template.Template{
			"index": mustParse(base, "index"),
		},
	}
}

func (r *TemplateRenderer) ExecuteTemplate(wr io.Writer, name string, data any) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Something went wrong XD")
	}
	return tmpl.ExecuteTemplate(wr, "base", data)
}
