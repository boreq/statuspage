package monitor

type Monitor interface {
	Status() []Status
}

type Config struct {
	Name string `json:"name"`
}

type StatusEnum string

const (
	UP      StatusEnum = "UP"
	DOWN               = "DOWN"
	FAILURE            = "FAILURE"
)

type Status struct {
	Config *Config    `json:"config,omitempty"`
	Output *string    `json:"output,omitempty"`
	Status StatusEnum `json:"status"`
}
