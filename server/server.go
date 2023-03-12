package server

import (
	"errors"
	"github.com/boreq/statuspage-backend/logging"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/query"
	"github.com/boreq/statuspage-backend/server/api"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

const showUptimeFromTimePeriod = 90 // [days]

var log = logging.GetLogger("server")

type handler struct {
	q *query.Query
	r *monitor.Runner
}

func (h *handler) Monitors(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	end := query.NewDate(time.Now())
	start := end.AddDate(0, 0, -showUptimeFromTimePeriod)

	response := MonitorsResponse{
		Monitors: make([]MonitorsResponseMonitor, 0),
	}

	for _, monitor := range h.r.Monitors() {
		monitor := monitor

		v := MonitorsResponseMonitor{
			Id:      monitor.Id(),
			Name:    monitor.Name(),
			Uptimes: make([]MonitorsResponseMonitorUptime, 0),
		}

		uptimes, err := h.q.List(monitor.Id(), start, end)
		if err != nil {
			log.Printf("Error listing: %s", err)
			return nil, api.InternalServerError
		}

		for _, uptime := range uptimes {
			uptime := uptime

			v.Uptimes = append(v.Uptimes, MonitorsResponseMonitorUptime{
				Date:   NewDate(uptime.Date),
				Uptime: uptime.Uptime,
			})
		}

		last, err := h.q.Latest(monitor.Id())
		if err != nil {
			if !errors.Is(err, query.ErrMeasurementNotFound) {
				log.Printf("Error getting last: %s", err)
				return nil, api.InternalServerError
			}
		} else {
			v.Status = &MonitorsResponseMonitorStatus{
				Timestamp: last.Timestamp.Unix(),
				Status:    EncodeStatus(last.Status),
			}
		}

		response.Monitors = append(response.Monitors, v)
	}

	return response, nil
}

type MonitorsResponse struct {
	Monitors []MonitorsResponseMonitor
}

type MonitorsResponseMonitor struct {
	Id      string
	Name    string
	Uptimes []MonitorsResponseMonitorUptime
	Status  *MonitorsResponseMonitorStatus
}

type MonitorsResponseMonitorStatus struct {
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}

type MonitorsResponseMonitorUptime struct {
	Date   Date     `json:"date"`
	Uptime *float64 `json:"uptime"`
}

type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func NewDate(d query.Date) Date {
	return Date{
		Year:  d.Year,
		Month: int(d.Month),
		Day:   d.Day,
	}
}

func EncodeStatus(status monitor.StatusEnum) string {
	switch status {
	case monitor.UP:
		return "UP"
	case monitor.DOWN:
		return "DOWN"
	case monitor.FAILURE:
		return "FAILURE"
	default:
		panic("unknown status")
	}
}

func Serve(q *query.Query, r *monitor.Runner, address string) error {
	h := &handler{r: r, q: q}

	router := httprouter.New()
	router.GET("/monitors/", api.Wrap(h.Monitors))

	return http.ListenAndServe(address, router)
}
