package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/models"
	"github.com/ARED-Group/dynamic-token-manager/internal/services"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	config        *config.Config
	tokenService  *services.TokenService
	deviceService *services.DeviceService
}

func NewAuthMiddleware(cfg *config.Config, tokenService *services.TokenService, deviceService *services.DeviceService) *AuthMiddleware {
	return &AuthMiddleware{
		config:        cfg,
		tokenService:  tokenService,
		deviceService: deviceService,
	}
}

// DeviceAuthMiddleware validates device authentication using serial number
func (a *AuthMiddleware) DeviceAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for device serial in header
		deviceSerial := r.Header.Get("X-Device-Serial")
		if deviceSerial == "" {
			http.Error(w, "Device serial number required", http.StatusUnauthorized)
			return
		}

		// Validate device
		if !a.deviceService.IsValidDevice(deviceSerial) {
			http.Error(w, "Invalid device", http.StatusForbidden)
			return
		}

		// Add device serial to context
		ctx := context.WithValue(r.Context(), "device_serial", deviceSerial)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWTAuthMiddleware validates JWT tokens
func (a *AuthMiddleware) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := a.tokenService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalDeviceAuth allows requests with or without device authentication
func (a *AuthMiddleware) OptionalDeviceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceSerial := r.Header.Get("X-Device-Serial")
		if deviceSerial != "" && a.deviceService.IsValidDevice(deviceSerial) {
			ctx := context.WithValue(r.Context(), "device_serial", deviceSerial)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
