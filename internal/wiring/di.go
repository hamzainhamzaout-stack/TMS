package wiring

import (
	"github.com/Robotech-Org/driver-management/internal/auth"
	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/driver"
)

type DependencyContainer struct {
	AuthHandler   *auth.AuthHandler
	DriverHandler *driver.DriverHandler
}

func NewDependencyContainer(cfg config.Config) *DependencyContainer {
	dbQueries := dm_svc_db.New(cfg.DB)

	authSvc := auth.NewAuthService(cfg.Logger, *dbQueries, cfg)
	driverSvc := driver.NewDriverService(cfg.Logger, *dbQueries, cfg)
	return &DependencyContainer{
		AuthHandler:   auth.NewAuthHandler(cfg.Logger, *authSvc),
		DriverHandler: driver.NewDriverHandler(cfg.Logger, *driverSvc, cfg.CLOUDINARY),
	}
}
