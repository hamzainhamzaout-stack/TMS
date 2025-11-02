package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/data/seed"
	"github.com/Robotech-Org/driver-management/internal/transport"
	"github.com/Robotech-Org/driver-management/internal/wiring"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	basectx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	godotenv.Load()

	baseURL := os.Getenv("DM_BASE_URL")
	if baseURL == "" {
		panic("DM_BASE_URL not set")
	}

	dbDsn := os.Getenv("DM_DB_DSN")
	if dbDsn == "" {
		panic("DM_DB_DSN not set")
	}

	serverPort := os.Getenv("DM_SERVER_PORT")
	if serverPort == "" {
		panic("DM_SERVER_PORT not set")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		panic("SECRET_KEY not set")
	}

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	if cloudName == "" {
		panic("CLOUDINARY_CLOUD_NAME not set")
	}
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	if apiKey == "" {
		panic("CLOUDINARY_API_KEY not set")
	}
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	if apiSecret == "" {
		panic("CLOUDINARY_API_SECRET not set")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		panic("Cloudinary initialization failed")
	}

	pool, err := pgxpool.New(basectx, dbDsn)
	if err != nil {
		panic("DB connection failed, err:" + err.Error())
	}
	defer pool.Close()

	if err := pool.Ping(basectx); err != nil {
		panic("DB ping failed")
	}

	queries := dm_svc_db.New(pool)

	config := config.Config{
		DB:         pool,
		Logger:     logger,
		BaseURL:    baseURL,
		ServerPort: serverPort,
		Queries:    *queries,
		SECRETKEY:  secretKey,
		CLOUDINARY: cld,
	}

	di := wiring.NewDependencyContainer(config)

	if err := seed.RunAll(basectx, config); err != nil {
		panic("Seed failed")
	}

	routes := transport.NewRoutes(di)

	log.Printf("Server starting on port %s", config.ServerPort)
	svc := http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: routes,
	}

	if err := svc.ListenAndServe(); err != nil {
		log.Panicf("error starting server: %v", err)
	}

}
