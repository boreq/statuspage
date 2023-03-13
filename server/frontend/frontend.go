package frontend

import (
	"embed"
	"fmt"
	"github.com/boreq/errors"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/query"
	"github.com/boreq/statuspage-backend/server/types"
	"html/template"
	"net/http"
	"strings"
)

//go:embed index.tmpl
var indexTemplate string

//go:embed assets/*
var assets embed.FS

type Handler struct {
	tmpl *template.Template
	q    *query.Query
	r    *monitor.Runner
}

func NewHandler(
	q *query.Query,
	r *monitor.Runner,
) (*Handler, error) {
	var funcMap = template.FuncMap{
		"StatusUp":      func() string { return types.StatusUp },
		"StatusDown":    func() string { return types.StatusDown },
		"StatusFailure": func() string { return types.StatusFailure },

		"DerefFloat64": func(v *float64) float64 { return *v },
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(indexTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a template")
	}

	return &Handler{
		tmpl: tmpl,
		q:    q,
		r:    r,
	}, nil
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/assets") {
		http.FileServer(http.FS(assets)).ServeHTTP(writer, request)
		return
	}

	if err := h.renderIndex(writer, request); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Println("index render error", err)
	}
}

func (h *Handler) renderIndex(writer http.ResponseWriter, request *http.Request) error {
	monitors, err := types.LoadMonitors(h.q, h.r)
	if err != nil {
		return errors.Wrap(err, "error loading monitors")
	}

	if err := h.tmpl.Execute(writer, struct {
		NumberOfDays int
		Monitors     []types.Monitor
	}{
		NumberOfDays: types.ShowUptimeFromTimePeriod,
		Monitors:     monitors,
	}); err != nil {
		return errors.Wrap(err, "error executing the template")
	}

	return nil
}
