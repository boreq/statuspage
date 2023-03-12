package db

import (
	"context"
	"errors"
	"github.com/boreq/statuspage-backend/logging"
	"github.com/dgraph-io/badger"
	"time"
)

var log = logging.GetLogger("gc")

const badgerGarbageCollectionErrorDelay = 1 * time.Minute

type GarbageCollector struct {
	db *badger.DB
}

func NewGarbageCollector(db *badger.DB) *GarbageCollector {
	return &GarbageCollector{db: db}
}

func (g *GarbageCollector) Run(ctx context.Context) {
	for {
		if err := g.gc(); err != nil {
			if !errors.Is(err, badger.ErrNoRewrite) {
				log.Printf("Error performing garbage collection: %s", err)
			}

			select {
			case <-time.After(badgerGarbageCollectionErrorDelay):
				continue
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func (g *GarbageCollector) gc() error {
	return g.db.RunValueLogGC(0.5)
}
