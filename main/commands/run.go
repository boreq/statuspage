package commands

import (
	"github.com/boreq/guinea"
	"github.com/boreq/statuspage-backend/config"
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
		return nil, err
	}

	// Serve the collected data
	if err := server.Serve(aggr, config.Config.ServeAddress); err != nil {
		return err
	}

	return nil
}
