package config

import (
	"log/slog"

	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DB         *pgxpool.Pool
	Logger     *slog.Logger
	BaseURL    string
	ServerPort string
	Queries    dm_svc_db.Queries
	SECRETKEY  string
}

func Load() (*Config, error) {
	return nil, nil
}
