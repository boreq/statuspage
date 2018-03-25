package monitor

import (
	"encoding/json"
	"github.com/boreq/statuspage-backend/logging"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

var log = logging.GetLogger("monitor")

func New(scriptsDirectory string) Monitor {
	rv := &monitor{
		scriptsDirectory: scriptsDirectory,
		status:           make(map[string]Status),
	}
	go rv.run()
	return rv
}

type monitor struct {
	scriptsDirectory string
	status           map[string]Status
}

func (m *monitor) run() {
	m.rerun()

	t := time.NewTicker(15 * time.Second)

	for range t.C {
		m.rerun()
	}
}

func (m *monitor) rerun() {
	// New
	filenames := make([]string, 0)
	files, err := ioutil.ReadDir(m.scriptsDirectory)
	if err != nil {
		log.Printf("Error: %s", err)
	}

	wg := sync.WaitGroup{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			filenames = append(filenames, getScriptName(f.Name()))
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := m.execute(f.Name())
				if err != nil {
					log.Printf("Error: %s", err)
					m.status[getScriptName(f.Name())] = Status{Status: FAILURE}
				}
			}()
		}
	}
	wg.Wait()

	// Cleanup
	for k, _ := range m.status {
		if !stringInSlice(k, filenames) {
			delete(m.status, k)
		}
	}
}

func (m *monitor) execute(filename string) error {
	var status Status

	// Load
	pth := path.Join(m.scriptsDirectory, filename)
	status.Config = new(Config)
	content, err := ioutil.ReadFile(pth)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, status.Config)
	if err != nil {
		return err
	}

	// Execute
	scriptFilename := getScriptName(filename)
	cmd := exec.Command(getScriptName(pth))
	log.Debugf("Running %s", pth)
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
	var rv []Status = make([]Status, 0)

	for _, v := range m.status {
		rv = append(rv, v)
	}

	return rv
}

func getScriptName(filename string) string {
	return strings.TrimSuffix(filename, ".json")
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
