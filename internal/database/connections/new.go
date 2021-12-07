package connections

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// NewFrontendDB creates a new connection to the frontend database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
//
// This connection is not expected to be closed but last the life of the calling application.
func NewFrontendDB(dsn, appName string, migrate bool) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(dsn, appName, "frontend")
	if err != nil {
		return nil, err
	}

	if !migrate {
		return db, nil
	}

	return db, runMigrations(db, schemas.Frontend)
}

// NewCodeIntelDB creates a new connection to the codeintel database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
//
// This connection is not expected to be closed but last the life of the calling application.
func NewCodeIntelDB(dsn, appName string, migrate bool) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(dsn, appName, "codeintel")
	if err != nil {
		return nil, err
	}

	if !migrate {
		return db, nil
	}

	return db, runMigrations(db, schemas.CodeIntel)
}

// NewCodeInsightsDB creates a new connection to the codeinsights database. After successful
// connection, the schema version of the database will be compared against an expected version and
// migrations may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
//
// This connection is not expected to be closed but last the life of the calling application.
func NewCodeInsightsDB(dsn, appName string, migrate bool) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(dsn, appName, "codeinsight")
	if err != nil {
		return nil, err
	}

	if !migrate {
		return db, nil
	}

	return db, runMigrations(db, schemas.CodeInsights)
}

func NewTestDB(dsn string, schemas ...*schemas.Schema) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(dsn, "test", "")
	if err != nil {
		return nil, err
	}

	return db, runMigrations(db, schemas...)
}

func runMigrations(db *sql.DB, schemas ...*schemas.Schema) error {
	ctx := context.Background()

	for _, schema := range schemas {
		// TODO - combine these operations
		store := store.NewWithDB(db, schema.MigrationsTableName, store.NewOperations(&observation.TestContext))
		if err := store.EnsureSchemaTable(ctx); err != nil {
			return err
		}

		storeFactory := map[string]runner.StoreFactory{schema.Name: func() (runner.Store, error) { return store, nil }}
		options := runner.Options{Up: true, SchemaNames: []string{schema.Name}}

		fmt.Printf("RUNNING\n")
		// TODO - can do this just once
		if err := runner.NewRunner(storeFactory).Run(ctx, options); err != nil {
			fmt.Printf("FAILING\n")
			return err
		}

		fmt.Printf("COMPLETED\n")
	}

	return nil
}
