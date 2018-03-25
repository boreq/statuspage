package monitor

func New(configDirectory string) Monitor {
	rv := &monitor{}
	return rv
}

type monitor struct {
}

func (m *monitor) Status() []Status {
	return nil
}
