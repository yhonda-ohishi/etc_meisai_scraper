package interceptors

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserClaims represents the JWT claims structure
type UserClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthConfig holds configuration for the auth interceptor
type AuthConfig struct {
	JWTSecret       string
	PublicMethods   []string
	AdminOnlyMethods []string
}

// NewAuthConfig creates a new auth configuration
func NewAuthConfig() *AuthConfig {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key" // Should be set in production
	}

	return &AuthConfig{
		JWTSecret: jwtSecret,
		PublicMethods: []string{
			"/etc.v1.EtcService/Health",
			"/etc.v1.EtcService/GetVersion",
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
		AdminOnlyMethods: []string{
			"/etc.v1.EtcService/DeleteAllRecords",
			"/etc.v1.EtcService/PurgeOldData",
		},
	}
}

// UnaryAuthInterceptor creates a unary server interceptor for JWT authentication
func UnaryAuthInterceptor(config *AuthConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if method is public
		if isPublicMethod(info.FullMethod, config.PublicMethods) {
			return handler(ctx, req)
		}

		// Extract and validate token
		userClaims, err := extractAndValidateToken(ctx, config.JWTSecret)
		if err != nil {
			return nil, err
		}

		// Check admin permissions for admin-only methods
		if isAdminOnlyMethod(info.FullMethod, config.AdminOnlyMethods) {
			if !hasAdminRole(userClaims.Roles) {
				return nil, status.Error(codes.PermissionDenied, "admin role required")
			}
		}

		// Inject user claims into context
		ctxWithUser := context.WithValue(ctx, "user_claims", userClaims)
		ctxWithUser = context.WithValue(ctxWithUser, "user_id", userClaims.UserID)
		ctxWithUser = context.WithValue(ctxWithUser, "username", userClaims.Username)

		return handler(ctxWithUser, req)
	}
}

// StreamAuthInterceptor creates a stream server interceptor for JWT authentication
func StreamAuthInterceptor(config *AuthConfig) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Check if method is public
		if isPublicMethod(info.FullMethod, config.PublicMethods) {
			return handler(srv, stream)
		}

		// Extract and validate token
		userClaims, err := extractAndValidateToken(stream.Context(), config.JWTSecret)
		if err != nil {
			return err
		}

		// Check admin permissions for admin-only methods
		if isAdminOnlyMethod(info.FullMethod, config.AdminOnlyMethods) {
			if !hasAdminRole(userClaims.Roles) {
				return status.Error(codes.PermissionDenied, "admin role required")
			}
		}

		// Create wrapped stream with user context
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          contextWithUser(stream.Context(), userClaims),
		}

		return handler(srv, wrappedStream)
	}
}

// wrappedServerStream wraps a grpc.ServerStream to override the context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// extractAndValidateToken extracts JWT token from metadata and validates it
func extractAndValidateToken(ctx context.Context, jwtSecret string) (*UserClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authorization := md.Get("authorization")
	if len(authorization) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	// Extract token from "Bearer <token>" format
	authHeader := authorization[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return nil, status.Error(codes.Unauthenticated, "empty token")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token claims")
	}

	// Check token expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, status.Error(codes.Unauthenticated, "token expired")
	}

	// Validate required claims
	if claims.UserID == "" || claims.Username == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid token: missing required claims")
	}

	return claims, nil
}

// isPublicMethod checks if a method is in the public methods list
func isPublicMethod(method string, publicMethods []string) bool {
	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

// isAdminOnlyMethod checks if a method requires admin role
func isAdminOnlyMethod(method string, adminOnlyMethods []string) bool {
	for _, adminMethod := range adminOnlyMethods {
		if method == adminMethod {
			return true
		}
	}
	return false
}

// hasAdminRole checks if user has admin role
func hasAdminRole(roles []string) bool {
	for _, role := range roles {
		if role == "admin" || role == "administrator" {
			return true
		}
	}
	return false
}

// contextWithUser adds user claims to context
func contextWithUser(ctx context.Context, claims *UserClaims) context.Context {
	ctx = context.WithValue(ctx, "user_claims", claims)
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "username", claims.Username)
	return ctx
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value("user_claims").(*UserClaims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// ValidateTokenString validates a JWT token string directly
func ValidateTokenString(tokenString, jwtSecret string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}