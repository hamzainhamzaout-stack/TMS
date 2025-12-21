package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Robotech-Org/driver-management/internal/domain"
	"github.com/Robotech-Org/driver-management/utils/responses"
	validator "github.com/Robotech-Org/driver-management/utils/validators"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	logger  *slog.Logger
	service AuthService
}

func NewAuthHandler(logger *slog.Logger, service AuthService) *AuthHandler {
	return &AuthHandler{
		logger:  logger,
		service: service,
	}
}

func (ah *AuthHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("CreateUserHandler handler started")
	var userProfileData domain.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&userProfileData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	user, err := ah.service.CreateUser(r.Context(), userProfileData)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with email '%s' already exists", userProfileData.PhoneNumber) {
			responses.ErrorJSONResponse(err, http.StatusConflict, w, nil)
			return
		}
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	userResponse := domain.User{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		RoleID:      user.RoleID,
		UserName:    user.UserName,
	}

	responses.SuccessJSONResponse("success", userResponse, http.StatusOK, w)
}

func (ah *AuthHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("UpdateUserHandler handler started")
	var userProfileData domain.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&userProfileData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	userID := chi.URLParam(r, "id")
	user, err := ah.service.UpdateUserService(r.Context(), userID, userProfileData)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with email '%s' already exists", userProfileData.PhoneNumber) {
			responses.ErrorJSONResponse(err, http.StatusConflict, w, nil)
			return
		}
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	userResponse := domain.User{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		RoleID:      user.RoleID,
	}

	responses.SuccessJSONResponse("success", userResponse, http.StatusOK, w)
}

func (ah *AuthHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("DeleteUserHandler handler started")
	userID := chi.URLParam(r, "id")
	err := ah.service.DeleteUserService(r.Context(), userID)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("success", nil, http.StatusOK, w)
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("Login handler started")
	if r.TLS != nil {
		ah.logger.Info("request is tls")
	}
	clientType := r.Header.Get("X-Client-Type")
	if clientType == "" {
		clientType = "web"
	}
	environment := os.Getenv("env")
	if environment == "" {
		environment = "development"
	}

	var userLoginData domain.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&userLoginData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	if err := validator.Validate(&userLoginData); err != nil {
		errorMsg := fmt.Errorf("validation error occurred")
		responses.ErrorJSONResponse(errorMsg, http.StatusBadRequest, w, nil)
		return
	}

	accessToken, refreshToken, passwordResetToken, err := ah.performLogin(r.Context(), userLoginData)
	if err != nil {
		ah.logger.Info("Error", "err", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	switch clientType {
	case "mobile":
		var responseData map[string]string
		if passwordResetToken != "" {
			responseData = map[string]string{
				"reset_token": passwordResetToken,
			}
		} else {
			responseData = map[string]string{
				"access_token":  accessToken,
				"refresh_token": refreshToken,
			}
		}

		responses.SuccessJSONResponse("Successfully authenticated", responseData, http.StatusOK, w)
	case "web":
		ah.logger.Info("Handling web client: Setting HttpOnly cookies.")
		if environment == "development" {

			if passwordResetToken != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     "passwordResetToken",
					Value:    passwordResetToken,
					Path:     "/",
					MaxAge:   int(time.Until(time.Now().Add(15 * time.Minute)).Seconds()),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				})
				responses.SuccessJSONResponse("Successfully authenticated", map[string]string{
					"reset_token": passwordResetToken,
				}, http.StatusOK, w)
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:     "accessToken",
					Value:    accessToken,
					Path:     "/",
					MaxAge:   int(time.Until(time.Now().Add(15 * time.Minute)).Seconds()),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				})

				http.SetCookie(w, &http.Cookie{
					Name:     "refreshToken",
					Value:    refreshToken,
					Path:     "/",
					MaxAge:   int(time.Until(time.Now().Add(3 * time.Hour)).Seconds()),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				})
				responses.SuccessJSONResponse("Successfully authenticated", nil, http.StatusOK, w)
			}
			ah.logger.Info("Cookies set for development environment with SameSite=None")

		} else {
			if passwordResetToken != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     "passwordResetToken",
					Value:    passwordResetToken,
					Path:     "/",
					MaxAge:   int(time.Until(time.Now().Add(15 * time.Minute)).Seconds()),
					HttpOnly: true,
					Secure:   r.TLS != nil,
				})
				responses.SuccessJSONResponse("Successfully authenticated", map[string]string{
					"reset_token": passwordResetToken,
				}, http.StatusOK, w)
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:     "accessToken",
					Value:    accessToken,
					Path:     "/",
					MaxAge:   int(time.Until(time.Now().Add(15 * time.Minute)).Seconds()),
					HttpOnly: true,
					Secure:   r.TLS != nil,
				})
				http.SetCookie(w, &http.Cookie{
					Name:     "refreshToken",
					Value:    refreshToken,
					Path:     "/",
					MaxAge:   int(time.Until(time.Now().Add(3 * time.Hour)).Seconds()),
					HttpOnly: true,
					Secure:   r.TLS != nil,
				})
				responses.SuccessJSONResponse("Successfully authenticated", nil, http.StatusOK, w)
			}
		}

	default:
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
}

func (ah *AuthHandler) performLogin(ctx context.Context, userLoginData domain.UserLogin) (string, string, string, error) {
	return ah.service.Login(ctx, userLoginData.UserName, userLoginData.Password)
}

func (ah *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refreshToken")
	if err != nil {
		responses.ErrorJSONResponse(fmt.Errorf("error: %v", err), http.StatusBadRequest, w, nil)
		return
	}
	if refreshCookie == nil {
		responses.ErrorJSONResponse(errors.New("refresh token cookie not found"), http.StatusBadRequest, w, nil)
		return
	}
	refreshToken := refreshCookie.Value
	if refreshToken == "" {
		responses.ErrorJSONResponse(fmt.Errorf("error: %v", err), http.StatusBadRequest, w, nil)
		return
	}

	newToken, newRefreshToken, err := ah.service.RefreshTokenService(r.Context(), refreshToken)
	if err != nil {
		ah.logger.Error("Failed to refresh token", "error", err)
		responses.ErrorJSONResponse(fmt.Errorf("failed to refresh token, please log in again"), http.StatusUnauthorized, w, nil)
		return
	}

	environment := os.Getenv("ENV")
	clearCookie := func(name string) {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		}

		if environment == "development" {
			cookie.Secure = true
			cookie.SameSite = http.SameSiteNoneMode
		} else {
			cookie.Secure = r.TLS != nil
		}

		http.SetCookie(w, cookie)
	}

	clearCookie("accessToken")
	clearCookie("refreshToken")

	if environment == "development" {
		http.SetCookie(w, &http.Cookie{
			Name:     "accessToken",
			Value:    newToken,
			Path:     "/",
			MaxAge:   int(time.Until(time.Now().Add(15 * time.Minute)).Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "refreshToken",
			Value:    newRefreshToken,
			Path:     "/",
			MaxAge:   int(time.Until(time.Now().Add(3 * time.Hour)).Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		ah.logger.Info("Cookies set for development environment with SameSite=None")
	} else {

		http.SetCookie(w, &http.Cookie{
			Name:     "accessToken",
			Value:    newToken,
			Path:     "/",
			MaxAge:   int(time.Until(time.Now().Add(15 * time.Minute)).Seconds()),
			HttpOnly: true,
			Secure:   r.TLS != nil,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "refreshToken",
			Value:    newRefreshToken,
			Path:     "/",
			MaxAge:   int(time.Until(time.Now().Add(3 * time.Hour)).Seconds()),
			HttpOnly: true,
			Secure:   r.TLS != nil,
		})
	}

	responses.SuccessJSONResponse("Token refreshed successfully", map[string]string{
		"access_token": newToken,
	}, http.StatusOK, w)
}

func (ah *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("Reset Password Handler Initiated")

	resetToken, err := r.Cookie("passwordResetToken")
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	var reqBody struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	if err := ah.service.ResetPasswordService(r.Context(), resetToken.Value, reqBody.Password); err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	environment := os.Getenv("ENV")
	clearCookie := func(name string) {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		}

		if environment == "development" {
			cookie.Secure = true
			cookie.SameSite = http.SameSiteNoneMode
		} else {
			cookie.Secure = r.TLS != nil
		}

		http.SetCookie(w, cookie)
	}

	clearCookie("passwordResetToken")

	responses.SuccessJSONResponse("password reset successful", nil, http.StatusOK, w)
}

func (mh *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		responses.ErrorJSONResponse(errors.New("user_id not found in context"), http.StatusUnauthorized, w, nil)
		return
	}

	mh.logger.Info("ProfileHandler handler started", "user", userID)

	userProfile, err := mh.service.GetUserProfile(r.Context(), userID)
	if err != nil {
		mh.logger.Error("failed to get user profile", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("success", userProfile, http.StatusOK, w)
}

func (ah *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refreshToken")
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	refreshToken := refreshCookie.Value
	if refreshToken == "" {

		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	if refreshToken == "" {
		ah.logger.Error("Empty refresh token provided")
		responses.ErrorJSONResponse(errors.New("error getting the token"), http.StatusBadRequest, w, nil)
		return
	}

	logoutMessage, err := ah.service.LogoutService(r.Context(), refreshToken)
	if err != nil {
		ah.logger.Error("Failed to logout", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	environment := os.Getenv("ENV")
	clearCookie := func(name string) {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		}

		if environment == "development" {
			cookie.Secure = true
			cookie.SameSite = http.SameSiteNoneMode
		} else {
			cookie.Secure = r.TLS != nil
		}

		http.SetCookie(w, cookie)
	}

	clearCookie("accessToken")
	clearCookie("refreshToken")

	responses.SuccessJSONResponse("Logout successful", map[string]string{
		"message": logoutMessage,
	}, http.StatusOK, w)
}

func (ah *AuthHandler) GetRolesHandler(w http.ResponseWriter, r *http.Request) {
	roles, err := ah.service.GetRolesService(r.Context())
	if err != nil {
		ah.logger.Error("Failed to get roles", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("Roles retrieved successfully", roles, http.StatusOK, w)
}

func (ah *AuthHandler) CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	var role domain.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		ah.logger.Error("Failed to decode role", "error", err)
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	createdRole, err := ah.service.CreateRoleService(r.Context(), role)
	if err != nil {
		ah.logger.Error("Failed to create role", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("Role created successfully", createdRole, http.StatusCreated, w)
}

func (ah *AuthHandler) UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	var role domain.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		ah.logger.Error("Failed to decode role", "error", err)
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	updatedRole, err := ah.service.UpdateRoleService(r.Context(), role)
	if err != nil {
		ah.logger.Error("Failed to update role", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("Role updated successfully", updatedRole, http.StatusOK, w)
}

func (ah *AuthHandler) DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	roleID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ah.logger.Error("Failed to parse role ID", "error", err)
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	err = ah.service.DeleteRoleService(r.Context(), int32(roleID))
	if err != nil {
		ah.logger.Error("Failed to delete role", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("Role deleted successfully", nil, http.StatusOK, w)
}

func (ah *AuthHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := ah.service.GetUsersService(r.Context())
	if err != nil {
		ah.logger.Error("Failed to get users", "error", err)
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("Users retrieved successfully", users, http.StatusOK, w)
}

func (ah *AuthHandler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("UpdateProfileHandler handler started")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		responses.ErrorJSONResponse(errors.New("user_id not found in context"), http.StatusUnauthorized, w, nil)
		return
	}

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	user, err := ah.service.UpdateUserProfile(r.Context(), userID, req)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	data := map[string]interface{}{
		"message":      "Profile updated successfully",
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"status":       user.Status,
		"phone_number": user.PhoneNumber,
	}

	responses.SuccessJSONResponse("success", data, http.StatusOK, w)
}
