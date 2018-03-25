package monitor

import (
	"encoding/json"
	"github.com/boreq/statuspage-backend/logging"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

var log = logging.GetLogger("monitor")

func New(configDirectory string) Monitor {
	rv := &monitor{
		configDirectory: configDirectory,
	}
	go rv.run()
	return rv
}

type monitor struct {
	configDirectory string
	status          map[string]Status
}

func (m *monitor) run() {
	t := time.NewTicker(15 * time.Second)

	for range t.C {
		// New
		filenames := make([]string, 0)
		files, err := ioutil.ReadDir(m.configDirectory)
		if err != nil {
			log.Printf("Error: %s", err)
		}

		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".json") {
				filenames = append(filenames, getScriptName(f.Name()))
				err := m.execute(f.Name())
				if err == nil {
					log.Printf("Error: %s", err)
					m.status[getScriptName(f.Name())] = Status{Status: FAILURE}
				}
			}
		}

		// Cleanup
		for k, _ := range m.status {
			if !stringInSlice(k, filenames) {
				delete(m.status, k)
			}
		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (m *monitor) execute(filename string) error {
	var status Status

	// Load
	status.Config = new(Config)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, status.Config)
	if err != nil {
		return err
	}

	// Execute
	scriptFilename := getScriptName(filename)
	cmd := exec.Command(scriptFilename)
	err = cmd.Run()
	if err == nil {
		status.Status = UP
	} else {
		status.Status = DOWN
	}

	m.status[scriptFilename] = status

	return nil
}

func (m *monitor) Status() []Status {
	return nil
}

func getScriptName(filename string) string {
	return strings.TrimSuffix(filename, ".json")
}
