package monitor

func New() Monitor {
	rv := &monitor{}
	return rv
}

type monitor struct {
}

func (m *monitor) Status() []Status {
	return nil
}
