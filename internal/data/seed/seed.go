package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func SeedRole(ctx context.Context, cfg config.Config, name string) (*int32, error) {
	role, err := cfg.Queries.CreateRole(ctx, name)
	if err == nil {
		fmt.Println("Role created successfully!")
		return &role.ID, nil
	}

	if err == pgx.ErrNoRows {
		fmt.Println("Role already exists skipping")
		return nil, err
	}
	return &role.ID, nil
}

func SeedAdmin(ctx context.Context, cfg config.Config, roleID int32) (uuid.UUID, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("changeme123"), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := cfg.Queries.CreateUser(
		ctx,
		dm_svc_db.CreateUserParams{
			FirstName:    "admin",
			LastName:     "user",
			PhoneNumber:  "09876567890",
			PasswordHash: hashedPassword,
			RoleID:       roleID,
		},
	)

	if err == nil {
		fmt.Println("Admin created successfully!")
		return user.ID, nil
	}

	if err == pgx.ErrNoRows {
		fmt.Println("Admin already exists, skipping...")
		return uuid.Nil, nil
	}

	return uuid.UUID{}, fmt.Errorf("create admin: %w", err)
}

func UpdateAdminStatus(ctx context.Context, cfg config.Config, userID uuid.UUID) error {
	userNew, err := cfg.Queries.UpdateUserStatus(ctx, userID)
	if err != nil {
		panic(err)
	}
	log.Printf("New user status: %v", userNew.Status)
	return nil
}

func RunAll(ctx context.Context, cfg config.Config) error {
	roleID, err := SeedRole(ctx, cfg, "fleet_manager")
	if err != nil {
		return err
	}

	adminID, err := SeedAdmin(ctx, cfg, *roleID)
	if err != nil {
		return err
	}

	err = UpdateAdminStatus(ctx, cfg, adminID)
	if err != nil {
		return err
	}

	return nil
}
