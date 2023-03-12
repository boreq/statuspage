package monitor

import "errors"

type Monitor struct {
	id         string
	name       string
	configPath string
	scriptPath string
}

func NewMonitor(
	id string,
	name string,
	configPath string,
	scriptPath string,
) (Monitor, error) {
	if id == "" {
		return Monitor{}, errors.New("empty id")
	}

	if name == "" {
		return Monitor{}, errors.New("empty name")
	}

	if configPath == "" {
		return Monitor{}, errors.New("empty config path")
	}

	if scriptPath == "" {
		return Monitor{}, errors.New("empty script path")
	}

	return Monitor{
		id:         id,
		name:       name,
		configPath: configPath,
		scriptPath: scriptPath,
	}, nil
}

func (m Monitor) Id() string {
	return m.id
}

func (m Monitor) ScriptPath() string {
	return m.scriptPath
}

func (m Monitor) Name() string {
	return m.name
}

func (m Monitor) ConfigPath() string {
	return m.configPath
}
