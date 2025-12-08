package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/domain"
	"github.com/Robotech-Org/driver-management/utils/gens"
	"github.com/Robotech-Org/driver-management/utils/tokenGen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	logger  *slog.Logger
	queries dm_svc_db.Queries
	config  config.Config
}

func NewAuthService(logger *slog.Logger, queries dm_svc_db.Queries, config config.Config) *AuthService {
	return &AuthService{
		logger:  logger,
		queries: queries,
		config:  config,
	}
}

func (as *AuthService) Login(ctx context.Context, phoneNumber, plainTextPassword string) (string, string, string, error) {
	as.logger.Info("Login service started")
	var access, refresh, reset string

	user, err := as.queries.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return "", "", "", err
	}
	log.Printf("User: %v", user)

	switch user.Status {
	case "uninitialized":
		userID, err := as.ValidateOTPService(ctx, phoneNumber, plainTextPassword)
		if err != nil {
			return "", "", "", err
		}

		reset, err = tokenGen.GeneratePasswordResetToken(userID)
		if err != nil {
			return "", "", "", err
		}

		return "", "", reset, nil

	case "active":
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainTextPassword)); err != nil {
			return "", "", "", fmt.Errorf("invalid credentials")
		}
		log.Printf("User: %v", user)

		roles, err := as.queries.GetMyRole(ctx, user.ID)
		if err != nil {
			return "", "", "", err
		}

		access, refresh, err = tokenGen.GenerateTokens(&user, roles)
		if err != nil {
			return "", "", "", err
		}

	case "inactive":
		return "", "", "", fmt.Errorf("user is inactive")

	case "deleted":
		return "", "", "", fmt.Errorf("user is deleted")

	default:
		return "", "", "", fmt.Errorf("invalid user status")
	}

	return access, refresh, reset, nil
}

func (as *AuthService) RefreshTokenService(ctx context.Context, refreshToken string) (string, string, error) {
	as.logger.Info("RefreshTokenService service started")
	isBlacklisted, err := as.queries.IsTokenBlacklisted(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	if isBlacklisted {
		return "", "", errors.New("refresh token is blacklisted")
	}

	access, refresh, err := tokenGen.GenerateNewAccessToken(refreshToken, &as.queries, ctx)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (as *AuthService) ValidateOTPService(ctx context.Context, phoneNumber string, userOTP string) (string, error) {
	as.logger.Info("OTP validation service started")
	user, err := as.queries.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		log.Println("Error when getting user by email")
		return "", err
	}

	fmt.Printf("UserID: %v", user.ID)

	otp, err := as.queries.GetOTPByUserID(ctx, user.ID)
	if err != nil {
		log.Printf("Error when getting OTP for user id: %v", user.ID)
		return "", err
	}
	log.Printf("OTP found in the db: %v", otp)

	if err := bcrypt.CompareHashAndPassword([]byte(otp.OtpHash), []byte(userOTP)); err != nil {
		log.Println("Couldn't compare users otp and otp hash from the database")
		return "", err
	}

	if time.Now().After(otp.ExpiresAt.Time) {
		return "", fmt.Errorf("otp has expired")
	}

	err = as.queries.UpdateOTPStatus(ctx, dm_svc_db.UpdateOTPStatusParams{
		Status: dm_svc_db.NullOtpStatus{
			OtpStatus: dm_svc_db.OtpStatusUsed,
			Valid:     true,
		},
		ID: otp.ID,
	})

	if err != nil {
		fmt.Println("Couldn't update/change the status of the otp in the table")
		return "", err
	}

	fmt.Println("Success")

	return user.ID.String(), nil
}

func (as *AuthService) CreateUser(ctx context.Context, user domain.User) (*dm_svc_db.User, error) {
	as.logger.Info("CreateUser service started")
	plainTextPassword, err := gens.GenerateRandomPassword()
	if err != nil {
		return nil, err
	}
	log.Printf("Generated password: %s", plainTextPassword)
	tx, err := as.config.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := as.queries.WithTx(tx)

	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return nil, err
	}

	userDB, err := qtx.CreateUser(ctx, dm_svc_db.CreateUserParams{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		PhoneNumber:  user.PhoneNumber,
		PasswordHash: hash,
		RoleID:       user.RoleID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("user with email '%s' already exists", user.PhoneNumber)
			}
		}
		as.logger.Error("failed to create user", slog.Any("err", err))
		return nil, err
	}
	as.logger.Info("User Created", "user detail", user)

	deliveryMethod := "sms"
	_, err = qtx.CreateOTP(ctx, dm_svc_db.CreateOTPParams{
		UserID:         userDB.ID,
		OtpHash:        string(hash),
		ExpiresAt:      pgtype.Timestamp{Time: time.Now().Add(5 * time.Minute), Valid: true},
		DeliveryMethod: &deliveryMethod,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		as.logger.Error("failed to commit transaction", slog.Any("err", err))
		return nil, err
	}

	return &userDB, nil
}

func (as *AuthService) ResetPasswordService(ctx context.Context, token string, password string) error {
	as.logger.Info("Reset password service started")
	userID, err := tokenGen.ValidatePasswordResetToken(token)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return as.queries.UpdateUserPassword(ctx, dm_svc_db.UpdateUserPasswordParams{
		ID:           parsedUUID,
		PasswordHash: hash,
	})
}

func (as *AuthService) LogoutService(ctx context.Context, refreshToken string) (string, error) {
	as.logger.Info("LogoutService service started")
	tx, err := as.config.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := as.queries.WithTx(tx)

	err = qtx.BlacklistToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return "logout successful", nil
}

func (as *AuthService) GetRolesService(ctx context.Context) ([]dm_svc_db.Role, error) {
	return as.queries.GetRoles(ctx)
}

func (ms *AuthService) GetUserProfile(ctx context.Context, userID string) (*dm_svc_db.UserProfileRow, error) {
	ms.logger.Info("GetUserProfile service started")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	user, err := ms.queries.UserProfile(ctx, userUUID)
	if err != nil {
		ms.logger.Error("Failed to get user profile", "error", err)
		return nil, err
	}
	return &user, nil
}

func (as *AuthService) CreateRoleService(ctx context.Context, roleData domain.Role) (string, error) {
	as.logger.Info("CreateRoleService service started")
	tx, err := as.config.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := as.queries.WithTx(tx)

	role, err := qtx.CreateRole(ctx, roleData.Name)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	log.Printf("Role created: %s", role.Name)

	return role.Name, nil
}

func (as *AuthService) UpdateRoleService(ctx context.Context, roleData domain.Role) (string, error) {
	as.logger.Info("UpdateRoleService service started")
	tx, err := as.config.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := as.queries.WithTx(tx)

	role, err := qtx.UpdateRole(ctx, dm_svc_db.UpdateRoleParams{
		ID:   roleData.ID,
		Name: roleData.Name,
	})
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return role.Name, nil
}

func (s *AuthService) DeleteRoleService(ctx context.Context, roleID int32) error {
	return s.queries.DeleteRole(ctx, roleID)
}

func (s *AuthService) GetUsersService(ctx context.Context) ([]dm_svc_db.User, error) {
	return s.queries.GetUsers(ctx)
}

func (as *AuthService) UpdateUserProfile(ctx context.Context, userID string, req domain.UpdateUserRequest) (*dm_svc_db.User, error) {
	as.logger.Info("UpdateUserProfile service started")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	params := dm_svc_db.UpdateUserProfileParams{
		ID:        userUUID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := as.queries.UpdateUserProfile(ctx, params)
	if err != nil {
		as.logger.Error("Failed to update user profile", "error", err)
		return nil, err
	}

	return &user, nil
}
func (s *AuthService) UpdateUserService(ctx context.Context, userID string, userProfileData domain.UserUpdateRequest) (*dm_svc_db.User, error) {
	s.logger.Info("UpdateUserService service started")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	params := dm_svc_db.UpdateUserParams{
		ID:          userUUID,
		FirstName:   userProfileData.FirstName,
		LastName:    userProfileData.LastName,
		PhoneNumber: userProfileData.PhoneNumber,
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update user", "error", err)
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) DeleteUserService(ctx context.Context, userID string) error {
	s.logger.Info("DeleteUserService service started")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	err = s.queries.DeleteUser(ctx, userUUID)
	if err != nil {
		s.logger.Error("Failed to delete user", "error", err)
		return err
	}

	return nil
}
