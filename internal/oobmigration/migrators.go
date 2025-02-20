package oobmigration

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// ExternalServiceConfigMigrator is a background job that encrypts
// external services config on startup.
// It periodically waits until a keyring is configured to determine
// how many services it must migrate.
// Scheduling and progress report is deleguated to the out of band
// migration package.
// The migration is non destructive and can be reverted.
type ExternalServiceConfigMigrator struct {
	store        *basestore.Store
	BatchSize    int
	AllowDecrypt bool
}

func NewExternalServiceConfigMigrator(store *basestore.Store) *ExternalServiceConfigMigrator {
	// not locking too many external services at a time to prevent congestion
	return &ExternalServiceConfigMigrator{store: store, BatchSize: 50}
}

func NewExternalServiceConfigMigratorWithDB(db dbutil.DB) *ExternalServiceConfigMigrator {
	return NewExternalServiceConfigMigrator(basestore.NewWithDB(db, sql.TxOptions{}))
}

// ID of the migration row in in the out_of_band_migrations table.
// This ID was defined arbitrarily in this migration file: frontend/1528395802_external_service_config_migration.up.sql.
func (m *ExternalServiceConfigMigrator) ID() int {
	return 3
}

// Progress returns a value from 0 to 1 representing the percentage of configuration already migrated.
func (m *ExternalServiceConfigMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				CAST(c1.count AS float) / CAST(c2.count AS float)
			END
		FROM
			(SELECT COUNT(*) AS count FROM external_services WHERE encryption_key_id != '') c1,
			(SELECT COUNT(*) AS count FROM external_services) c2
	`)))
	return progress, err
}

// Up loads BatchSize external services, locks them, and encrypts their config using the
// key returned by keyring.Default().
// If there is no ring, it will periodically try again until the key is setup in the config.
// Up ensures the configuration can be decrypted with the same key before overwitting it.
// The key id is stored alongside the encrypted configuration.
func (m *ExternalServiceConfigMigrator) Up(ctx context.Context) (err error) {
	key := keyring.Default().ExternalServiceKey
	if key == nil {
		return nil
	}

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	services, err := m.listConfigsForUpdate(ctx, tx, false)
	if err != nil {
		return err
	}

	for _, svc := range services {
		encryptedCfg, err := key.Encrypt(ctx, []byte(svc.Config))
		if err != nil {
			return err
		}

		version, err := key.Version(ctx)
		if err != nil {
			return err
		}
		keyIdent := version.JSON()

		// ensure encryption round-trip is valid with keyIdent
		decrypted, err := key.Decrypt(ctx, encryptedCfg)
		if err != nil {
			return err
		}
		if decrypted.Secret() != svc.Config {
			return errors.New("invalid encryption round-trip")
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE external_services SET config = %s, encryption_key_id = %s WHERE id = %s",
			encryptedCfg,
			keyIdent,
			svc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (m *ExternalServiceConfigMigrator) Down(ctx context.Context) (err error) {
	key := keyring.Default().ExternalServiceKey
	if key == nil {
		return nil
	}

	if !m.AllowDecrypt {
		return nil
	}

	// For records that were encrypted, we need to decrypt the configuration,
	// store it in plain text and remove the encryption_key_id.
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	services, err := m.listConfigsForUpdate(ctx, tx, true)
	if err != nil {
		return err
	}

	for _, svc := range services {
		secret, err := key.Decrypt(ctx, []byte(svc.Config))
		if err != nil {
			return err
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE external_services SET config = %s, encryption_key_id = '' WHERE id = %s",
			secret.Secret(),
			svc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (m *ExternalServiceConfigMigrator) listConfigsForUpdate(ctx context.Context, tx *basestore.Store, encrypted bool) ([]*types.ExternalService, error) {
	// Select and lock a few records within this transaction. This ensures
	// that many frontend instances can run the same migration concurrently
	// without them all trying to convert the same record.
	q := "SELECT id, config FROM external_services "
	if encrypted {
		q += "WHERE encryption_key_id != ''"
	} else {
		q += "WHERE encryption_key_id = ''"
	}

	q += "ORDER BY id ASC LIMIT %s FOR UPDATE SKIP LOCKED"

	rows, err := tx.Query(ctx, sqlf.Sprintf(q, m.BatchSize))

	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var services []*types.ExternalService

	for rows.Next() {
		var svc types.ExternalService
		if err := rows.Scan(&svc.ID, &svc.Config); err != nil {
			return nil, err
		}
		services = append(services, &svc)
	}

	return services, nil
}

// ExternalAccountsMigrator is a background job that encrypts
// external accounts data on startup.
// It periodically waits until a keyring is configured to determine
// how many services it must migrate.
// Scheduling and progress report is delegated to the out of band
// migration package.
// The migration is non destructive and can be reverted.
type ExternalAccountsMigrator struct {
	store        *basestore.Store
	BatchSize    int
	AllowDecrypt bool
}

func NewExternalAccountsMigrator(store *basestore.Store) *ExternalAccountsMigrator {
	// not locking too many external accounts at a time to prevent congestion
	return &ExternalAccountsMigrator{store: store, BatchSize: 50}
}

func NewExternalAccountsMigratorWithDB(db dbutil.DB) *ExternalAccountsMigrator {
	return NewExternalAccountsMigrator(basestore.NewWithDB(db, sql.TxOptions{}))
}

// ID of the migration row in the out_of_band_migrations table.
// This ID was defined arbitrarily in this migration file: frontend/1528395809_external_account_migration.up.sql
func (m *ExternalAccountsMigrator) ID() int {
	return 6
}

// Progress returns a value from 0 to 1 representing the percentage of configuration already migrated.
func (m *ExternalAccountsMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				CAST(c1.count AS float) / CAST(c2.count AS float)
			END
		FROM
			(SELECT COUNT(*) AS count FROM user_external_accounts WHERE encryption_key_id != '' OR (account_data IS NULL AND auth_data IS NULL)) c1,
			(SELECT COUNT(*) AS count FROM user_external_accounts) c2
	`)))
	return progress, err
}

// Up loads BatchSize external accounts, locks them, and encrypts their config using the
// key returned by keyring.Default().
// If there is no ring, it will periodically try again until the key is setup in the config.
// Up ensures the configuration can be decrypted with the same key before overwitting it.
// The key id is stored alongside the encrypted configuration.
func (m *ExternalAccountsMigrator) Up(ctx context.Context) (err error) {
	key := keyring.Default().UserExternalAccountKey
	if key == nil {
		return nil
	}

	version, err := key.Version(ctx)
	if err != nil {
		return err
	}

	keyIdent := version.JSON()

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalAccountsWith(tx)
	accounts, err := store.ListBySQL(ctx, sqlf.Sprintf("WHERE encryption_key_id = '' AND (account_data IS NOT NULL OR auth_data IS NOT NULL) ORDER BY id ASC LIMIT %s FOR UPDATE SKIP LOCKED", m.BatchSize))
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		var (
			encAuthData *string
			encData     *string
		)
		if acc.AuthData != nil {
			encrypted, err := key.Encrypt(ctx, *acc.AuthData)
			if err != nil {
				return err
			}

			// ensure encryption round-trip is valid
			decrypted, err := key.Decrypt(ctx, encrypted)
			if err != nil {
				return err
			}
			if decrypted.Secret() != string(*acc.AuthData) {
				return errors.New("invalid encryption round-trip")
			}

			encAuthData = strptr(string(encrypted))
		}

		if acc.Data != nil {
			encrypted, err := key.Encrypt(ctx, *acc.Data)
			if err != nil {
				return err
			}

			// ensure encryption round-trip is valid
			decrypted, err := key.Decrypt(ctx, encrypted)
			if err != nil {
				return err
			}
			if decrypted.Secret() != string(*acc.Data) {
				return errors.New("invalid encryption round-trip")
			}

			encData = strptr(string(encrypted))
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE user_external_accounts SET auth_data = %s, account_data = %s, encryption_key_id = %s WHERE id = %d",
			encAuthData,
			encData,
			keyIdent,
			acc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

func strptr(s string) *string {
	return &s
}

func (m *ExternalAccountsMigrator) Down(ctx context.Context) (err error) {
	key := keyring.Default().UserExternalAccountKey
	if key == nil {
		return nil
	}

	if !m.AllowDecrypt {
		return nil
	}

	// For records that were encrypted, we need to decrypt the configuration,
	// store it in plain text and remove the encryption_key_id.
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalAccountsWith(tx)
	accounts, err := store.ListBySQL(ctx, sqlf.Sprintf("WHERE encryption_key_id != '' ORDER BY id ASC LIMIT %s FOR UPDATE SKIP LOCKED", m.BatchSize))
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE user_external_accounts SET auth_data = %s, encryption_key_id = '' WHERE id = %s",
			acc.AuthData,
			acc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

// ExternalServiceWebhookMigrator is a background job that calculates the
// has_webhooks field on external services based on the external service
// configuration.
type ExternalServiceWebhookMigrator struct {
	store     *basestore.Store
	BatchSize int
}

var _ Migrator = &ExternalServiceWebhookMigrator{}

func NewExternalServiceWebhookMigrator(store *basestore.Store) *ExternalServiceWebhookMigrator {
	// Batch size arbitrarily chosen to match ExternalServiceConfigMigrator.
	return &ExternalServiceWebhookMigrator{store: store, BatchSize: 50}
}

func NewExternalServiceWebhookMigratorWithDB(db dbutil.DB) *ExternalServiceWebhookMigrator {
	return NewExternalServiceWebhookMigrator(basestore.NewWithDB(db, sql.TxOptions{}))
}

// ID returns the migration row ID in the out_of_band_migrations table.
//
// This ID was defined in the migration:
// migrations/frontend/1528395921_add_has_webhooks.up.sql
func (m *ExternalServiceWebhookMigrator) ID() int {
	return 13
}

// Progress returns a value from 0 to 1 representing the percentage of external
// services that have had their has_webhooks field calculated.
func (m *ExternalServiceWebhookMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				CAST(c1.count AS float) / CAST(c2.count AS float)
			END
		FROM
			(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NOT NULL) c1,
			(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL) c2
	`)))
	return progress, err
}

// Up loads BatchSize external services, locks them, and upserts them back into
// the database, which will calculate HasWebhooks along the way.
func (m *ExternalServiceWebhookMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalServicesWith(tx)

	svcs, err := store.List(ctx, database.ExternalServicesListOptions{
		OrderByDirection: "ASC",
		LimitOffset:      &database.LimitOffset{Limit: m.BatchSize},
		NoCachedWebhooks: true,
		ForUpdate:        true,
	})
	if err != nil {
		return err
	}

	err = store.Upsert(ctx, svcs...)
	return err
}

func (*ExternalServiceWebhookMigrator) Down(context.Context) error {
	// There's no sensible down migration here: if the SQL down migration has
	// been run, then the field no longer exists, and there's nothing to do.
	return nil
}
