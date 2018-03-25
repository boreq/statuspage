package monitor

type Monitor interface {
	Status() []Status
}

type Config struct {
	Name string
}

type StatusEnum string

const (
	UP      StatusEnum = "UP"
	DOWN               = "DOWN"
	FAILURE            = "FAILURE"
)

type Status struct {
	Config Config
	Status StatusEnum
}
