package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	gmigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	embeddedmigrations "github.com/rawdah/rawdah-api/migrations"
)

const migrationsTable = "rawdah_schema_migrations"
const legacyMigrationsTable = "rawdah_schema_migrations_legacy"

func Run(ctx context.Context, db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	bootstrapVersion, err := bootstrapFromLegacyTable(ctx, db)
	if err != nil {
		return err
	}

	src, err := iofs.New(embeddedmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("init migration source: %w", err)
	}

	dbDriver, err := pgx.WithInstance(db.DB, &pgx.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		return fmt.Errorf("init migration database driver: %w", err)
	}

	m, err := gmigrate.NewWithInstance("iofs", src, "pgx5", dbDriver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	// Do not call m.Close() here when using pgx.WithInstance on a shared *sql.DB.
	// golang-migrate closes the provided database driver, which closes the app DB pool.
	// The source is embedded (iofs) and does not require explicit teardown in this path.

	if bootstrapVersion != nil {
		if err := m.Force(*bootstrapVersion); err != nil {
			return fmt.Errorf("force migration version %d: %w", *bootstrapVersion, err)
		}
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, gmigrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func bootstrapFromLegacyTable(ctx context.Context, db *sqlx.DB) (*int, error) {
	exists, err := tableExists(ctx, db, migrationsTable)
	if err != nil {
		return nil, err
	}
	if !exists {
		legacyExists, err := tableExists(ctx, db, legacyMigrationsTable)
		if err != nil {
			return nil, err
		}
		if !legacyExists {
			return nil, nil
		}
		return readMaxLegacyVersion(ctx, db, legacyMigrationsTable)
	}

	hasDirtyColumn, err := tableHasColumn(ctx, db, migrationsTable, "dirty")
	if err != nil {
		return nil, err
	}
	if hasDirtyColumn {
		return nil, nil
	}

	legacyVersion, err := readMaxLegacyVersion(ctx, db, migrationsTable)
	if err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx,
		fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, migrationsTable, legacyMigrationsTable),
	); err != nil {
		return nil, fmt.Errorf("rename legacy migration table: %w", err)
	}

	return legacyVersion, nil
}

func readMaxLegacyVersion(ctx context.Context, db *sqlx.DB, table string) (*int, error) {
	var legacyVersion sql.NullInt64
	if err := db.QueryRowxContext(ctx, fmt.Sprintf(`SELECT MAX(version) FROM %s`, table)).Scan(&legacyVersion); err != nil {
		return nil, fmt.Errorf("read legacy migration version: %w", err)
	}
	if !legacyVersion.Valid {
		return nil, nil
	}
	v := int(legacyVersion.Int64)
	return &v, nil
}

func tableExists(ctx context.Context, db *sqlx.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowxContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = current_schema()
			  AND table_name = $1
		)
	`, table).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check table existence for %s: %w", table, err)
	}
	return exists, nil
}

func tableHasColumn(ctx context.Context, db *sqlx.DB, table, column string) (bool, error) {
	var exists bool
	err := db.QueryRowxContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = $1
			  AND column_name = $2
		)
	`, table, column).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check column %s.%s: %w", table, column, err)
	}
	return exists, nil
}
