package db

import (
	"errors"
	"github.com/boreq/statuspage-backend/monitor"
)

//go:generate msgp

type PersistedMeasurement struct {
	Id          string  `msg:"id"`
	UpdateEvery float64 `msg:"updateEvery"`
	Timestamp   int64   `msg:"timestamp"`
	Duration    float64 `msg:"duration"`
	Status      string  `msg:"status"`
	Output      string  `msg:"string"`
}

const (
	statusFailure = "failure"
	statusUp      = "up"
	statusDown    = "down"
)

func encodeStatus(status monitor.StatusEnum) (string, error) {
	switch status {
	case monitor.FAILURE:
		return statusFailure, nil
	case monitor.UP:
		return statusUp, nil
	case monitor.DOWN:
		return statusDown, nil
	default:
		return "", errors.New("unknown status")
	}
}

func decodeStatus(status string) (monitor.StatusEnum, error) {
	switch status {
	case statusFailure:
		return monitor.FAILURE, nil
	case statusUp:
		return monitor.UP, nil
	case statusDown:
		return monitor.DOWN, nil
	default:
		return "", errors.New("unknown status")
	}
}
