package commands

import (
	"github.com/boreq/guinea"
	"github.com/boreq/statuspage-backend/config"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/server"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{"config", false, "Config file"},
	},
	ShortDescription: "runs the program",
}

func runRun(c guinea.Context) error {
	if err := config.Load(c.Arguments[0]); err != nil {
		return err
	}

	m := monitor.New(config.Config.ScriptsDirectory, config.Config.UpdateEverySeconds)

	if err := server.Serve(m, config.Config.ServeAddress); err != nil {
		return err
	}

	return nil
}
