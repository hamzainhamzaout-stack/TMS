package tokenGen

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	UserID              string              `json:"user_id"`
	Status              dm_svc_db.UserState    `json:"user_state"`
	RoleName string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// Generate Access And Refresh Tokens For Users in the Authentication Flow
func GenerateTokens(user *dm_svc_db.User, role dm_svc_db.Role) (string, string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return "", "", errors.New("no secret key")
	}

	iat := time.Now()

	// Generate Access Token (15 minutes expiry)
	accessClaims := &Claims{
		UserID:              user.ID.String(),
		Status:              user.Status,
		RoleName: role.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(iat.Add(15 * time.Minute)),
			Issuer:    "RMS-Users",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token (24 hours expiry)
	refreshClaims := &Claims{
		UserID: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(iat.Add(24 * time.Hour)),
			Issuer:    "RMS-Users",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// Parse Token and Retrieve Token Claims
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(*Claims), nil
}

func GenerateNewAccessToken(refreshToken string, queries *dm_svc_db.Queries, ctx context.Context) (string, string, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil || !token.Valid {
		return "", "", err
	}
	claims, ok := token.Claims.(*Claims)

	if !ok || claims.ExpiresAt.Unix() < time.Now().Unix() {
		return "", "", fmt.Errorf("refresh token has expired")
	}

	iat := time.Now()

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", err
	}
	role, err := queries.GetMyRole(ctx, userID)
	if err != nil {
		return "", "", err
	}

	newAccessClaims := &Claims{
		UserID:              claims.UserID,
		Status:              claims.Status,
		RoleName: role.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(iat.Add(15 * time.Minute)),
			Issuer:    "RMS-Users",
		},
	}
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessClaims)
	newAccessTokenString, err := newAccessToken.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", "", err
	}

	newRefreshClaims := &Claims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(iat.Add(24 * time.Hour)),
			Issuer:    "RMS-Users",
		},
	}
	newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newRefreshClaims)
	newRefreshTokenString, err := newRefreshToken.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", "", err
	}
	return newAccessTokenString, newRefreshTokenString, nil
}

func ValidateRefreshToken(refreshToken string, context context.Context) (int, error) {
	if refreshToken == "" {
		return http.StatusBadRequest, fmt.Errorf("refresh token is empty")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return http.StatusBadRequest, fmt.Errorf("invalid refresh token")
	}

	if claims.ExpiresAt.Unix() < time.Now().Unix() {
		return http.StatusBadRequest, fmt.Errorf("refresh token has expired")
	}

	return http.StatusOK, nil
}

type PasswordResetClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GeneratePasswordResetToken(userID string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("no secret key")
	}

	iat := time.Now()

	claims := &PasswordResetClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iat),
			NotBefore: jwt.NewNumericDate(iat),
			ExpiresAt: jwt.NewNumericDate(iat.Add(15 * time.Minute)), // Short-lived token
			Issuer:    "RMS-Users-PasswordReset",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidatePasswordResetToken(tokenString string) (string, error) {
	claims := &PasswordResetClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid password reset token")
	}

	if claims.ExpiresAt.Unix() < time.Now().Unix() {
		return "", fmt.Errorf("password reset token has expired")
	}

	return claims.UserID, nil
}
