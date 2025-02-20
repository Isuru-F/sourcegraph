package writer

import (
	"context"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type DatabaseWriter interface {
	WriteDBFile(ctx context.Context, args types.SearchArgs, tempDBFile string) error
}

type databaseWriter struct {
	path            string
	gitserverClient gitserver.GitserverClient
	parser          parser.Parser
}

func NewDatabaseWriter(
	path string,
	gitserverClient gitserver.GitserverClient,
	parser parser.Parser,
) DatabaseWriter {
	return &databaseWriter{
		path:            path,
		gitserverClient: gitserverClient,
		parser:          parser,
	}
}

func (w *databaseWriter) WriteDBFile(ctx context.Context, args types.SearchArgs, dbFile string) error {
	if newestDBFile, oldCommit, ok, err := w.getNewestCommit(ctx, args); err != nil {
		return err
	} else if ok {
		if ok, err := w.writeFileIncrementally(ctx, args, dbFile, newestDBFile, oldCommit); err != nil || ok {
			return err
		}
	}

	return w.writeDBFile(ctx, args, dbFile)
}

func (w *databaseWriter) getNewestCommit(ctx context.Context, args types.SearchArgs) (dbFile string, commit string, ok bool, err error) {
	newest, err := findNewestFile(filepath.Join(w.path, diskcache.EncodeKeyComponent(string(args.Repo))))
	if err != nil || newest == "" {
		return "", "", false, err
	}

	err = store.WithSQLiteStore(newest, func(db store.Store) (err error) {
		if commit, ok, err = db.GetCommit(ctx); err != nil {
			return errors.Wrap(err, "store.GetCommit")
		}

		return nil
	})

	return newest, commit, ok, err
}

func (w *databaseWriter) writeDBFile(ctx context.Context, args types.SearchArgs, dbFile string) error {
	return w.parseAndWriteInTransaction(ctx, args, nil, dbFile, func(tx store.Store, symbols <-chan result.Symbol) error {
		if err := tx.CreateMetaTable(ctx); err != nil {
			return errors.Wrap(err, "store.CreateMetaTable")
		}
		if err := tx.CreateSymbolsTable(ctx); err != nil {
			return errors.Wrap(err, "store.CreateSymbolsTable")
		}
		if err := tx.InsertMeta(ctx, string(args.CommitID)); err != nil {
			return errors.Wrap(err, "store.InsertMeta")
		}
		if err := tx.WriteSymbols(ctx, symbols); err != nil {
			return errors.Wrap(err, "store.WriteSymbols")
		}
		if err := tx.CreateSymbolIndexes(ctx); err != nil {
			return errors.Wrap(err, "store.CreateSymbolIndexes")
		}

		return nil
	})
}

// The maximum number of paths when doing incremental indexing. Diffs with more paths than this will
// not be incrementally indexed, and instead we will process all symbols.
const maxTotalPaths = 999

// The maximum sum of bytes in paths in a diff when doing incremental indexing. Diffs bigger than this
// will not be incrementally indexed, and instead we will process all symbols. Without this limit, we
// could hit HTTP 431 (header fields too large) when sending the list of paths `git archive paths...`.
// The actual limit is somewhere between 372KB and 450KB, and we want to be well under that.
// 100KB seems safe.
const maxTotalPathsLength = 100000

func (w *databaseWriter) writeFileIncrementally(ctx context.Context, args types.SearchArgs, dbFile, newestDBFile, oldCommit string) (bool, error) {
	changes, err := w.gitserverClient.GitDiff(ctx, args.Repo, api.CommitID(oldCommit), args.CommitID)
	if err != nil {
		return false, errors.Wrap(err, "gitserverClient.GitDiff")
	}

	// Paths to re-parse
	addedOrModifiedPaths := append(changes.Added, changes.Modified...)

	// Paths to modify in the database
	addedModifiedOrDeletedPaths := append(addedOrModifiedPaths, changes.Deleted...)

	// Too many entries
	if len(addedModifiedOrDeletedPaths) > maxTotalPaths {
		return false, nil
	}

	totalPathsLength := 0
	for _, path := range addedModifiedOrDeletedPaths {
		totalPathsLength += len(path)
	}
	// Argument lists too long
	if totalPathsLength > maxTotalPathsLength {
		return false, nil
	}

	if err := copyFile(newestDBFile, dbFile); err != nil {
		return false, err
	}

	return true, w.parseAndWriteInTransaction(ctx, args, addedOrModifiedPaths, dbFile, func(tx store.Store, symbols <-chan result.Symbol) error {
		if err := tx.UpdateMeta(ctx, string(args.CommitID)); err != nil {
			return errors.Wrap(err, "store.UpdateMeta")
		}
		if err := tx.DeletePaths(ctx, addedModifiedOrDeletedPaths); err != nil {
			return errors.Wrap(err, "store.DeletePaths")
		}
		if err := tx.WriteSymbols(ctx, symbols); err != nil {
			return errors.Wrap(err, "store.WriteSymbols")
		}

		return nil
	})
}

func (w *databaseWriter) parseAndWriteInTransaction(ctx context.Context, args types.SearchArgs, paths []string, dbFile string, callback func(tx store.Store, symbols <-chan result.Symbol) error) error {
	symbols, err := w.parser.Parse(ctx, args, paths)
	if err != nil {
		return errors.Wrap(err, "parser.Parse")
	}

	return store.WithSQLiteStoreTransaction(ctx, dbFile, func(tx store.Store) error {
		return callback(tx, symbols)
	})
}
