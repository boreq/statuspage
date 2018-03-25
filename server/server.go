package server

import (
	"github.com/boreq/statuspage-backend/logging"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/server/api"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var log = logging.GetLogger("server")

type handler struct {
	m monitor.Monitor
}

func (h *handler) Status(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	var response []monitor.Status = h.m.Status()
	return response, nil
}

func Serve(m monitor.Monitor, address string) error {
	h := &handler{m: m}

	router := httprouter.New()
	router.GET("/status.json", api.Wrap(h.Status))

	return http.ListenAndServe(address, router)
}
