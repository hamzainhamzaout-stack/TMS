package seed

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func SeedRole(ctx context.Context, cfg config.Config, name string) (*int32, error) {
	role, err := cfg.Queries.CreateRole(ctx, name)
	if err == nil {
		fmt.Println("Role created successfully!")
		return &role.ID, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		fmt.Println("Role already exists, getting ID...")
		roles, err := cfg.Queries.GetRoles(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get roles: %w", err)
		}
		for _, r := range roles {
			if r.Name == name {
				return &r.ID, nil
			}
		}
		return nil, fmt.Errorf("role '%s' not found after creation failed", name)
	}

	return nil, err
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
			UserName:     "admin",
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

func SeedExpenses(ctx context.Context, cfg config.Config) error {
	expenses := []dm_svc_db.Expense{
		{
			Name: "Food",
		},
		{
			Name: "Transport",
		},
	}

	for _, expense := range expenses {
		_, err := cfg.Queries.CreateExpenseType(ctx, expense.Name)
		if err != nil {
			log.Printf("Error creating expense: %v", err)
			return fmt.Errorf("create expense: %w", err)
		}
	}

	fmt.Println("Expenses seeded successfully!")
	return nil
}

func UpdateAdminStatus(ctx context.Context, cfg config.Config, userID uuid.UUID) error {
	userNew, err := cfg.Queries.UpdateUserStatus(ctx, userID)
	if err != nil {
		return err
	}
	log.Printf("New user status: %v", userNew.Status)
	return nil
}

func RunAll(ctx context.Context, cfg config.Config) error {
	roleID, err := SeedRole(ctx, cfg, "fleet_manager")
	if err != nil {
		return errors.New("error seeding fleet manager role")
	}

	_, err = SeedRole(ctx, cfg, "finance_manager")
	if err != nil {
		return errors.New("error seeding finance manager role")
	}

	adminID, err := SeedAdmin(ctx, cfg, *roleID)
	if err != nil {
		return errors.New("error seeding admin")
	}

	if adminID != uuid.Nil {
		err = UpdateAdminStatus(ctx, cfg, adminID)
		if err != nil {
			return errors.New("error updating admin status")
		}
	}

	err = SeedExpenses(ctx, cfg)
	if err != nil {
		return errors.New("error seeding expenses")
	}

	return nil
}
