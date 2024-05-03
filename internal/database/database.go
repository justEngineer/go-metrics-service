package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"embed"

	"github.com/cenkalti/backoff"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	config "github.com/justEngineer/go-metrics-service/internal/http/server/config"
	"github.com/justEngineer/go-metrics-service/internal/storage"
)

type Database struct {
	Connections *pgxpool.Pool
	mainContext *context.Context
}

//go:embed migrations/*.sql
var migrationSQL embed.FS

const migrationsDir = "migrations"

const (
	insertGaugeSQL = `INSERT INTO
		gauge_metrics (id, value)
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE
	SET value = EXCLUDED.value`

	insertCounterSQL = `INSERT INTO
		counter_metrics (id, value)
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE
	SET value = counter_metrics.value + EXCLUDED.value`

	selectGaugeSQL = `SELECT value FROM gauge_metrics WHERE id = $1`

	selectCounterSQL = `SELECT value FROM counter_metrics WHERE id = $1`
)

// Получаем одно соединение для базы данных
func NewConnection(ctx context.Context, cfg *config.ServerConfig) (*Database, error) {
	connect, err := pgxpool.New(ctx, cfg.DBConnection)
	db := Database{Connections: connect, mainContext: &ctx}
	if err != nil {
		return &db, err
	} else {
		err = db.applyMigrations(cfg)
	}
	return &db, err
}

// ApplyMigrations применяет миграции к базе данных.
func (d *Database) applyMigrations(cfg *config.ServerConfig) error {
	srcDriver, err := iofs.New(migrationSQL, migrationsDir)
	if err != nil {
		return fmt.Errorf("unable to apply db migrations: %v", err)
	}
	// Создаем экземпляр драйвера базы данных для PostgreSQL.
	db, err := sql.Open("postgres", cfg.DBConnection)
	if err != nil {
		return fmt.Errorf("unable to create db driver: %v", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("unable to create db instance: %v", err)
	}
	// Создаем новый экземпляр мигратора с использованием драйвера источника и драйвера базы данных PostgreSQL.
	migrator, err := migrate.NewWithInstance("migration_embedded_sql_files", srcDriver, "psql_db", driver)
	if err != nil {
		return fmt.Errorf("unable to create migration: %v", err)
	}
	// Закрываем мигратор в конце работы функции.
	defer migrator.Close()
	// Применяем миграции.
	if err = migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("unable to apply migrations %v", err)
	}
	return nil
}

func (d *Database) Ping() error {
	ctx, cancel := context.WithTimeout(*d.mainContext, 1*time.Second)
	defer cancel()
	err := d.Connections.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) SetGaugeMetric(ctx context.Context, key string, value float64) error {
	f := func() error {
		if _, err := d.Connections.Exec(ctx, insertGaugeSQL, key, value); err != nil {
			return err
		}
		return nil
	}
	err := executeWithBackoff(f)
	return err
}

func (d *Database) SetCounterMetric(ctx context.Context, key string, value int64) error {
	f := func() error {
		if _, err := d.Connections.Exec(ctx, insertCounterSQL, key, value); err != nil {
			return err
		}
		return nil
	}
	err := executeWithBackoff(f)
	return err
}

func (d *Database) GetGaugeMetric(ctx context.Context, key string) (float64, error) {
	var value float64 = 0
	f := func() error {
		var result float64
		row := d.Connections.QueryRow(ctx, selectGaugeSQL, key)
		if err := row.Scan(&result); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("gauge metric is not found: %w, id: %v", err, key)
			}
			return err
		}
		value = result
		return nil
	}
	err := executeWithBackoff(f)
	return value, err
}

func (d *Database) GetCounterMetric(ctx context.Context, key string) (int64, error) {
	var value int64 = 0
	f := func() error {
		var result int64
		row := d.Connections.QueryRow(ctx, selectCounterSQL, key)
		if err := row.Scan(&result); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("counter metric is not found: %w, id: %v", err, key)
			}
			return err
		}
		value = result
		return nil
	}
	err := executeWithBackoff(f)
	return value, err
}

func (d *Database) SetMetricsBatch(ctx context.Context, gaugesBatch []storage.GaugeMetric, countersBatch []storage.CounterMetric) error {
	f := func() error {
		tx, err := d.Connections.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		for _, gaugeMetric := range gaugesBatch {
			if _, err := tx.Exec(ctx, insertGaugeSQL, gaugeMetric.Name, gaugeMetric.Value); err != nil {
				errRollback := tx.Rollback(ctx)
				if errRollback != nil {
					return errRollback
				}
				return err
			}
		}
		for _, counterMetric := range countersBatch {
			if _, err := tx.Exec(ctx, insertCounterSQL, counterMetric.Name, counterMetric.Value); err != nil {
				errRollback := tx.Rollback(ctx)
				if errRollback != nil {
					return errRollback
				}
				return err
			}
		}
		return tx.Commit(ctx)
	}
	return executeWithBackoff(f)
}

func executeWithBackoff(f func() error) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 3 * time.Second
	expBackoff.RandomizationFactor = 0
	err := backoff.Retry(f, expBackoff)
	if err != nil {
		return fmt.Errorf("failed to connect to database after retrying: %v", err)
	}
	return err
}
