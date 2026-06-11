package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/app"
	coursemigrations "github.com/luxeave/entropy-course/migrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const schemaMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

func newMigrateCommand(ctx context.Context, config *viper.Viper) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
	}

	command.AddCommand(
		migrateActionCommand(ctx, config, "up", "Apply pending migrations", runMigrationsUp),
		migrateActionCommand(ctx, config, "down", "Rollback the latest migration", runMigrationsDown),
		migrateActionCommand(ctx, config, "status", "Show migration status", runMigrationsStatus),
	)

	return command
}

func migrateActionCommand(
	ctx context.Context,
	config *viper.Viper,
	use string,
	short string,
	run func(context.Context, app.Config, io.Writer) error,
) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, err := app.LoadConfig(config)
			if err != nil {
				return err
			}

			return run(ctx, cfg, command.OutOrStdout())
		},
	}
}

func runMigrationsUp(ctx context.Context, cfg app.Config, writer io.Writer) error {
	return withMigrator(ctx, cfg, func(migrator migrationRunner) error {
		return migrator.Up(ctx, writer)
	})
}

func runMigrationsDown(ctx context.Context, cfg app.Config, writer io.Writer) error {
	return withMigrator(ctx, cfg, func(migrator migrationRunner) error {
		return migrator.Down(ctx, writer)
	})
}

func runMigrationsStatus(ctx context.Context, cfg app.Config, writer io.Writer) error {
	return withMigrator(ctx, cfg, func(migrator migrationRunner) error {
		return migrator.Status(ctx, writer)
	})
}

func withMigrator(ctx context.Context, cfg app.Config, run func(migrationRunner) error) error {
	if strings.TrimSpace(cfg.DBURL) == "" {
		return fmt.Errorf("connect db: %w", app.ErrMissingDatabaseURL)
	}

	registered, err := coursemigrations.Registered()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	return run(migrationRunner{
		pool:       pool,
		migrations: registered,
	})
}

type migrationRunner struct {
	pool       *pgxpool.Pool
	migrations []coursemigrations.Migration
}

func (runner migrationRunner) Up(ctx context.Context, writer io.Writer) error {
	if err := runner.ensureSchema(ctx); err != nil {
		return err
	}

	applied, err := runner.applied(ctx)
	if err != nil {
		return err
	}

	appliedCount := 0
	for _, migration := range runner.migrations {
		if applied[migration.Version] {
			continue
		}
		if err := runner.apply(ctx, migration); err != nil {
			return err
		}
		appliedCount++
		_, _ = fmt.Fprintf(writer, "applied %s\n", migration.Version)
	}
	if appliedCount == 0 {
		_, _ = fmt.Fprintln(writer, "no pending migrations")
	}

	return nil
}

func (runner migrationRunner) Down(ctx context.Context, writer io.Writer) error {
	if err := runner.ensureSchema(ctx); err != nil {
		return err
	}

	applied, err := runner.applied(ctx)
	if err != nil {
		return err
	}

	for index := len(runner.migrations) - 1; index >= 0; index-- {
		migration := runner.migrations[index]
		if !applied[migration.Version] {
			continue
		}
		if err := runner.rollback(ctx, migration); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(writer, "rolled back %s\n", migration.Version)
		return nil
	}

	_, _ = fmt.Fprintln(writer, "no applied migrations")
	return nil
}

func (runner migrationRunner) Status(ctx context.Context, writer io.Writer) error {
	if err := runner.ensureSchema(ctx); err != nil {
		return err
	}

	applied, err := runner.applied(ctx)
	if err != nil {
		return err
	}

	for _, migration := range runner.migrations {
		status := "pending"
		if applied[migration.Version] {
			status = "applied"
		}
		_, _ = fmt.Fprintf(writer, "%s\t%s\n", migration.Version, status)
	}

	return nil
}

func (runner migrationRunner) ensureSchema(ctx context.Context) error {
	if _, err := runner.pool.Exec(ctx, schemaMigrationsTableSQL); err != nil {
		return fmt.Errorf("ensure schema migrations table: %w", err)
	}

	return nil
}

func (runner migrationRunner) applied(ctx context.Context) (map[string]bool, error) {
	rows, err := runner.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}

	return applied, nil
}

func (runner migrationRunner) apply(ctx context.Context, migration coursemigrations.Migration) error {
	return runner.inTransaction(ctx, func(tx pgx.Tx) error {
		if err := execSQL(ctx, tx, migration.UpSQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, migration.Version); err != nil {
			return fmt.Errorf("record migration %s: %w", migration.Version, err)
		}

		return nil
	})
}

func (runner migrationRunner) rollback(ctx context.Context, migration coursemigrations.Migration) error {
	return runner.inTransaction(ctx, func(tx pgx.Tx) error {
		if err := execSQL(ctx, tx, migration.DownSQL); err != nil {
			return fmt.Errorf("rollback migration %s: %w", migration.Version, err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, migration.Version); err != nil {
			return fmt.Errorf("unrecord migration %s: %w", migration.Version, err)
		}

		return nil
	})
}

func (runner migrationRunner) inTransaction(ctx context.Context, run func(pgx.Tx) error) error {
	tx, err := runner.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := run(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration transaction: %w", err)
	}

	return nil
}

func execSQL(ctx context.Context, tx pgx.Tx, sql string) error {
	for _, statement := range strings.Split(sql, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := tx.Exec(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}
