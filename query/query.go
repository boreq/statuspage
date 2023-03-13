package query

import (
	"github.com/boreq/errors"
	"github.com/boreq/statuspage-backend/monitor"
	"time"
)

type Date struct {
	Day   int
	Month time.Month
	Year  int
}

func NewDate(t time.Time) Date {
	local := t.Local()
	return Date{
		Day:   local.Day(),
		Month: local.Month(),
		Year:  local.Year(),
	}
}

func (d Date) AddDate(years int, months int, days int) Date {
	return NewDate(d.time().AddDate(years, months, days))
}

func (d Date) time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.Local)
}

func (d Date) Before(o Date) bool {
	return d.time().Before(o.time())
}

var ErrMeasurementNotFound = errors.New("measurement not found")

type MeasurementsStorage interface {
	Get(id string, start, end Date) ([]monitor.Measurement, error)
	Last(id string) (monitor.Measurement, error)
}

type ListResult struct {
	Date   Date
	Uptime *float64
}

type LastResult struct {
	Status    monitor.StatusEnum
	Timestamp time.Time
}

type Query struct {
	measurementsStorage MeasurementsStorage
}

func NewQuery(measurementsStorage MeasurementsStorage) *Query {
	return &Query{measurementsStorage: measurementsStorage}
}

func (q *Query) Latest(id string) (LastResult, error) {
	last, err := q.measurementsStorage.Last(id)
	if err != nil {
		return LastResult{}, errors.Wrap(err, "error getting last measurement")
	}

	return LastResult{
		Status:    last.Status(),
		Timestamp: last.Timestamp(),
	}, nil
}
func (q *Query) List(id string, start, end Date) ([]ListResult, error) {
	if end.Before(start) {
		return nil, errors.New("end before start")
	}

	measurements, err := q.measurementsStorage.Get(id, start, end)
	if err != nil {
		return nil, errors.Wrap(err, "error getting measurements")
	}

	aggr := make(map[Date]*aggregation)

	for _, m := range measurements {
		date := NewDate(m.Timestamp())
		if _, ok := aggr[date]; !ok {
			aggr[date] = &aggregation{}
		}

		switch m.Status() {
		case monitor.UP:
			aggr[date].Ups++
		case monitor.DOWN:
			aggr[date].Downs++
		}
	}

	var result []ListResult
	for d := start; !end.Before(d); d = d.AddDate(0, 0, 1) {
		r := ListResult{
			Date: d,
		}

		if a, ok := aggr[d]; ok && (a.Ups != 0 || a.Downs != 0) {
			uptime := float64(a.Ups) / float64(a.Ups+a.Downs)
			r.Uptime = &uptime
		}

		result = append(result, r)
	}
	return result, nil
}

type aggregation struct {
	Ups   int
	Downs int
}
