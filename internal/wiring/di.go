package wiring

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/Robotech-Org/driver-management/internal/auth"
	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/driver"
	"github.com/Robotech-Org/driver-management/internal/tracking"
)

type DependencyContainer struct {
	AuthHandler     *auth.AuthHandler
	DriverHandler   *driver.DriverHandler
	TrackingHandler *tracking.Handler
}

func NewDependencyContainer(cfg config.Config) *DependencyContainer {
	dbQueries := dm_svc_db.New(cfg.DB)

	authSvc := auth.NewAuthService(cfg.Logger, *dbQueries, cfg)
	driverSvc := driver.NewDriverService(cfg.Logger, *dbQueries, cfg)
	trackingSvc := tracking.NewService(cfg.Logger, cfg.DB)

	// GPS points older than TRACKING_RETENTION_DAYS (default 90) are deleted daily.
	retentionDays := 90
	if v, err := strconv.Atoi(os.Getenv("TRACKING_RETENTION_DAYS")); err == nil && v > 0 {
		retentionDays = v
	}
	go trackingSvc.RunRetention(context.Background(), time.Duration(retentionDays)*24*time.Hour)

	return &DependencyContainer{
		AuthHandler:     auth.NewAuthHandler(cfg.Logger, *authSvc),
		DriverHandler:   driver.NewDriverHandler(cfg.Logger, *driverSvc, cfg.CLOUDINARY),
		TrackingHandler: tracking.NewHandler(cfg.Logger, trackingSvc),
	}
}
