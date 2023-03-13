package server

import (
	"github.com/boreq/errors"
	"github.com/boreq/statuspage-backend/logging"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/query"
	"github.com/boreq/statuspage-backend/server/api"
	"github.com/boreq/statuspage-backend/server/frontend"
	"github.com/boreq/statuspage-backend/server/types"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var log = logging.GetLogger("server")

type handler struct {
	q *query.Query
	r *monitor.Runner
}

func (h *handler) Monitors(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	monitors, err := types.LoadMonitors(h.q, h.r)
	if err != nil {
		log.Printf("Error loading monitors: %s", err)
		return nil, api.InternalServerError
	}

	response := MonitorsResponse{
		Monitors: monitors,
	}

	return response, nil
}

type MonitorsResponse struct {
	Monitors []types.Monitor
}

func Serve(q *query.Query, r *monitor.Runner, address string) error {
	apiHandler := &handler{r: r, q: q}

	frontendHandler, err := frontend.NewHandler(q, r)
	if err != nil {
		return errors.Wrap(err, "error creating the frontend handler")
	}

	router := httprouter.New()
	router.GET("/api/monitors/", api.Wrap(apiHandler.Monitors))
	router.NotFound = frontendHandler

	return http.ListenAndServe(address, router)
}
