package monitor

import (
	"errors"
	"time"
)

type StatusEnum string

const (
	UP      StatusEnum = "UP"
	DOWN               = "DOWN"
	FAILURE            = "FAILURE"
)

type Measurement struct {
	id          string
	updateEvery time.Duration
	timestamp   time.Time
	duration    time.Duration
	status      StatusEnum
	output      string
}

func NewMeasurement(id string, updateEvery time.Duration, timestamp time.Time, duration time.Duration, status StatusEnum, output string) (Measurement, error) {
	if duration < 0 {
		return Measurement{}, errors.New("duration < 0")
	}

	if updateEvery < 0 {
		return Measurement{}, errors.New("updateEvery < 0")
	}

	return Measurement{
		id:          id,
		updateEvery: updateEvery,
		timestamp:   timestamp,
		duration:    duration,
		status:      status,
		output:      output,
	}, nil
}

func (m Measurement) Id() string {
	return m.id
}

func (m Measurement) UpdateEvery() time.Duration {
	return m.updateEvery
}

func (m Measurement) Timestamp() time.Time {
	return m.timestamp
}

func (m Measurement) Duration() time.Duration {
	return m.duration
}

func (m Measurement) Status() StatusEnum {
	return m.status
}

func (m Measurement) Output() string {
	return m.output
}
