package monitor

import (
	"context"
	"encoding/json"
	"github.com/boreq/errors"
	"github.com/boreq/statuspage-backend/logging"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

var log = logging.GetLogger("monitor")

const configPathSuffix = ".json"

type MeasurementsStorage interface {
	Add(m Measurement) error
}

type Runner struct {
	scriptsDirectory    string
	updateEvery         time.Duration
	measurementsStorage MeasurementsStorage

	monitors      []Monitor
	monitorsMutex sync.Mutex
}

func NewRunner(
	scriptsDirectory string,
	updateEverySeconds int,
	measurementsStorage MeasurementsStorage,
) *Runner {
	rv := &Runner{
		scriptsDirectory:    scriptsDirectory,
		updateEvery:         time.Duration(updateEverySeconds) * time.Second,
		measurementsStorage: measurementsStorage,
	}
	return rv
}

func (m *Runner) Monitors() []Monitor {
	m.monitorsMutex.Lock()
	defer m.monitorsMutex.Unlock()

	tmp := make([]Monitor, len(m.monitors))
	copy(tmp, m.monitors)
	return tmp
}

func (m *Runner) Run(ctx context.Context) {
	for {
		if err := m.rerun(); err != nil {
			log.Printf("Error: %s", err)
		}

		select {
		case <-time.After(m.updateEvery):
		case <-ctx.Done():
			return
		}
	}
}

func (m *Runner) rerun() error {
	monitors, err := m.loadMonitors()
	if err != nil {
		return errors.Wrap(err, "error loading monitors")
	}

	m.monitorsMutex.Lock()
	m.monitors = monitors
	m.monitorsMutex.Unlock()

	wg := sync.WaitGroup{}
	for _, monitor := range monitors {
		monitor := monitor
		wg.Add(1)
		go func() {
			defer wg.Done()

			var measurement Measurement
			var err error

			start := time.Now()
			result, err := m.execute(monitor)
			duration := time.Since(start)
			if err != nil {
				log.Printf("Error executing a monitor: %s", err)

				measurement, err = NewMeasurement(
					monitor.Id(),
					m.updateEvery,
					start,
					duration,
					FAILURE,
					"",
				)
			} else {
				var status StatusEnum
				if result.Up {
					status = UP
				} else {
					status = DOWN
				}

				measurement, err = NewMeasurement(
					monitor.Id(),
					m.updateEvery,
					start,
					duration,
					status,
					result.Output,
				)
			}

			if err := m.measurementsStorage.Add(measurement); err != nil {
				log.Printf("Error saving a measurement: %s", err)
			}

		}()
	}
	wg.Wait()
	return nil
}

func (m *Runner) loadMonitors() ([]Monitor, error) {
	files, err := os.ReadDir(m.scriptsDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "error reading directory")
	}

	var monitors []Monitor
	for _, f := range files {
		if strings.HasSuffix(f.Name(), configPathSuffix) {
			monitor, err := m.loadMonitor(f.Name())
			if err != nil {
				return nil, errors.Wrapf(err, "error loading monitor '%s'", f.Name())
			}
			monitors = append(monitors, monitor)
		}
	}
	return monitors, nil
}

func (m *Runner) loadMonitor(filename string) (Monitor, error) {
	configPath := path.Join(m.scriptsDirectory, filename)
	scriptPath := strings.TrimSuffix(configPath, configPathSuffix)

	content, err := os.ReadFile(configPath)
	if err != nil {
		return Monitor{}, errors.Wrap(err, "error reading config file")
	}

	var c config
	if err = json.Unmarshal(content, &c); err != nil {
		return Monitor{}, errors.Wrap(err, "error parsing json")
	}

	monitor, err := NewMonitor(
		strings.TrimSuffix(filename, configPathSuffix),
		c.Name,
		configPath,
		scriptPath,
	)
	if err != nil {
		return Monitor{}, errors.Wrap(err, "error creating a monitor")
	}

	return monitor, nil
}

func (m *Runner) execute(monitor Monitor) (executionResult, error) {
	cmd := exec.Command(monitor.ScriptPath())
	log.Debugf("Running %s", monitor.scriptPath)

	output, err := cmd.Output()
	if err != nil {
		if err == exec.ErrNotFound {
			return executionResult{}, errors.Wrap(err, "error executing a command")
		}

		exitErr := &exec.ExitError{}
		if !errors.As(err, &exitErr) {
			return executionResult{}, errors.Wrap(err, "exit error")
		}
	}

	return executionResult{
		Output: string(output),
		Up:     err == nil,
	}, nil
}

type config struct {
	Name string `json:"name"`
}

type executionResult struct {
	Up     bool
	Output string
}
