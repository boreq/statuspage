package commands

import (
	"context"
	"github.com/boreq/errors"
	"github.com/boreq/guinea"
	"github.com/boreq/statuspage-backend/config"
	"github.com/boreq/statuspage-backend/db"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/query"
	"github.com/boreq/statuspage-backend/server"
	"github.com/dgraph-io/badger"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "config",
			Multiple:    false,
			Optional:    false,
			Description: "Config file",
		},
	},
	ShortDescription: "runs the program",
}

func runRun(c guinea.Context) error {
	if err := config.Load(c.Arguments[0]); err != nil {
		return err
	}

	b, err := badger.Open(badger.DefaultOptions(config.Config.DataDirectory))
	if err != nil {
		return errors.Wrap(err, "error opening badger")
	}

	s := db.NewMeasurementsStorage(b)
	m := monitor.NewRunner(config.Config.ScriptsDirectory, 60, s)
	q := query.NewQuery(s)
	gc := db.NewGarbageCollector(b)

	go m.Run(context.Background())
	go gc.Run(context.Background())

	if err := server.Serve(q, m, config.Config.ServeAddress); err != nil {
		return err
	}

	return nil
}
