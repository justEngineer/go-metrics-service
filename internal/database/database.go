package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	config "github.com/justEngineer/go-metrics-service/internal/http/server/config"
)

type Database struct {
	Connections *pgxpool.Pool
}

const (
	createTablesSQL = `
	CREATE TABLE IF NOT EXISTS gauge_metrics (
		id      TEXT PRIMARY KEY,
		value   DOUBLE PRECISION NOT NULL
	);
	CREATE TABLE IF NOT EXISTS counter_metrics (
		id      TEXT PRIMARY KEY,
		value   BIGSERIAL NOT NULL
	);`

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
	//	connect, err := sql.Open("pgx", cfg.DBConnection)
	connect, err := pgxpool.New(ctx, cfg.DBConnection)
	db := Database{Connections: connect}
	if err != nil {
		return &db, err
	}
	//connect, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	_, err = connect.Exec(ctx, createTablesSQL)
	return &db, err
}

func (d *Database) Ping() error {
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// defer cancel()
	// err := d.Connections.Ping(ctx)
	err := d.Connections.Ping(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) SetGaugeMetric(ctx context.Context, key string, value float64) error {
	if _, err := d.Connections.Exec(ctx, insertGaugeSQL, key, value); err != nil {
		return err
	}
	return nil
}

func (d *Database) SetCounterMetric(ctx context.Context, key string, value int64) error {
	if _, err := d.Connections.Exec(ctx, insertCounterSQL, key, value); err != nil {
		return err
	}
	return nil
}

func (d *Database) GetGaugeMetric(ctx context.Context, key string) (float64, error) {
	var result float64
	row := d.Connections.QueryRow(ctx, selectGaugeSQL, key)
	if err := row.Scan(&result); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("gauge metric is not found")
		}
		return 0, err
	}
	return result, nil
}

func (d *Database) GetCounterMetric(ctx context.Context, key string) (int64, error) {
	var result int64
	row := d.Connections.QueryRow(ctx, selectCounterSQL, key)
	if err := row.Scan(&result); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("counter metric is not found")
		}
		return 0, err
	}
	return result, nil
}
