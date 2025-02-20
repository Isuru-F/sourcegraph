package writer

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
)

type CachedDatabaseWriter interface {
	GetOrCreateDatabaseFile(ctx context.Context, args types.SearchArgs) (string, error)
}

type cachedDatabaseWriter struct {
	databaseWriter DatabaseWriter
	cache          *diskcache.Store
}

func NewCachedDatabaseWriter(databaseWriter DatabaseWriter, cache *diskcache.Store) CachedDatabaseWriter {
	return &cachedDatabaseWriter{
		databaseWriter: databaseWriter,
		cache:          cache,
	}
}

// The version of the symbols database schema. This is included in the database filenames to prevent a
// newer version of the symbols service from attempting to read from a database created by an older and
// likely incompatible symbols service. Increment this when you change the database schema.
const symbolsDBVersion = 4

func (w *cachedDatabaseWriter) GetOrCreateDatabaseFile(ctx context.Context, args types.SearchArgs) (string, error) {
	key := []string{
		string(args.Repo),
		fmt.Sprintf("%s-%d", args.CommitID, symbolsDBVersion),
	}

	cacheFile, err := w.cache.OpenWithPath(ctx, key, func(fetcherCtx context.Context, tempDBFile string) error {
		if err := w.databaseWriter.WriteDBFile(fetcherCtx, args, tempDBFile); err != nil {
			return errors.Wrap(err, "databaseWriter.WriteDBFile")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	defer cacheFile.File.Close()

	return cacheFile.File.Name(), err
}
