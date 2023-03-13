package types

import (
	"github.com/boreq/errors"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/query"
	"time"
)

const (
	StatusUp      = "UP"
	StatusDown    = "DOWN"
	StatusFailure = "FAILURE"
)

const (
	ShowUptimeFromTimePeriod = 90 // [days]
)

type Monitor struct {
	Name    string          `json:"name"`
	Uptimes []MonitorUptime `json:"uptimes"`
	Status  *MonitorStatus  `json:"status"`
}

type MonitorStatus struct {
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}

type MonitorUptime struct {
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
		return StatusUp
	case monitor.DOWN:
		return StatusDown
	case monitor.FAILURE:
		return StatusFailure
	default:
		panic("unknown status")
	}
}

func LoadMonitors(q *query.Query, r *monitor.Runner) ([]Monitor, error) {
	monitors := make([]Monitor, 0)

	end := query.NewDate(time.Now())
	start := end.AddDate(0, 0, -ShowUptimeFromTimePeriod)

	for _, monitor := range r.Monitors() {
		monitor := monitor

		v := Monitor{
			Name:    monitor.Name(),
			Uptimes: make([]MonitorUptime, 0),
		}

		uptimes, err := q.List(monitor.Id(), start, end)
		if err != nil {
			return nil, errors.Wrap(err, "error getting listing measurements")
		}

		for _, uptime := range uptimes {
			uptime := uptime

			v.Uptimes = append(v.Uptimes, MonitorUptime{
				Date:   NewDate(uptime.Date),
				Uptime: uptime.Uptime,
			})
		}

		last, err := q.Latest(monitor.Id())
		if err != nil {
			if !errors.Is(err, query.ErrMeasurementNotFound) {
				return nil, errors.Wrap(err, "error getting latest measurement")
			}
		} else {
			v.Status = &MonitorStatus{
				Timestamp: last.Timestamp.Unix(),
				Status:    EncodeStatus(last.Status),
			}
		}

		monitors = append(monitors, v)
	}

	return monitors, nil
}
