package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	config "github.com/justEngineer/go-metrics-service/internal/http/server/config"
)

type Database struct {
	Connection *sql.DB
}

// Получаем одно соединение для базы данных
func NewConnection(cfg *config.ServerConfig) (*Database, error) {
	connect, err := sql.Open("pgx", cfg.DBConnection)
	db := Database{Connection: connect}
	//connect, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	return &db, err
}

func (db *Database) Ping() error {
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// defer cancel()
	// err := db.Connection.PingContext(ctx)
	err := db.Connection.Ping()
	if err != nil {
		return err
	}
	return nil
}
