package commands

import (
	"github.com/boreq/statuspage-backend/config"
	"github.com/boreq/statuspage-backend/storage"
	"github.com/boreq/statuspage-backend/storage/bolt"
)

func initialize(configFilename string) (storage.Storage, error) {
	if err := config.Load(configFilename); err != nil {
		return nil, err
	}

	return bolt.New(config.Config.DatabaseFile)
}
