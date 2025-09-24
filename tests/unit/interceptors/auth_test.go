package interceptors_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/interceptors"
)

// Mock implementations for testing
type mockUnaryHandler struct {
	mock.Mock
}

func (m *mockUnaryHandler) Handle(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

type mockStreamHandler struct {
	mock.Mock
}

func (m *mockStreamHandler) Handle(srv interface{}, stream grpc.ServerStream) error {
	args := m.Called(srv, stream)
	return args.Error(0)
}

type mockServerStream struct {
	mock.Mock
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *mockServerStream) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *mockServerStream) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

// Test data
const (
	testSecret         = "test-secret-key-12345"
	testUserID         = "user123"
	testUsername       = "testuser"
	testPublicMethod   = "/etc.v1.EtcService/Health"
	testPrivateMethod  = "/etc.v1.EtcService/GetRecords"
	testAdminMethod    = "/etc.v1.EtcService/DeleteAllRecords"
)

func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("JWT_SECRET", testSecret)
	code := m.Run()
	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}

func TestNewAuthConfig(t *testing.T) {
	tests := []struct {
		name      string
		envSecret string
		want      *interceptors.AuthConfig
	}{
		{
			name:      "with JWT_SECRET environment variable",
			envSecret: "custom-secret",
			want: &interceptors.AuthConfig{
				JWTSecret: "custom-secret",
				PublicMethods: []string{
					"/etc.v1.EtcService/Health",
					"/etc.v1.EtcService/GetVersion",
					"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
				},
				AdminOnlyMethods: []string{
					"/etc.v1.EtcService/DeleteAllRecords",
					"/etc.v1.EtcService/PurgeOldData",
				},
			},
		},
		{
			name:      "without JWT_SECRET environment variable",
			envSecret: "",
			want: &interceptors.AuthConfig{
				JWTSecret: "default-secret-key",
				PublicMethods: []string{
					"/etc.v1.EtcService/Health",
					"/etc.v1.EtcService/GetVersion",
					"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
				},
				AdminOnlyMethods: []string{
					"/etc.v1.EtcService/DeleteAllRecords",
					"/etc.v1.EtcService/PurgeOldData",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalSecret := os.Getenv("JWT_SECRET")
			defer os.Setenv("JWT_SECRET", originalSecret)

			if tt.envSecret == "" {
				os.Unsetenv("JWT_SECRET")
			} else {
				os.Setenv("JWT_SECRET", tt.envSecret)
			}

			// Execute
			got := interceptors.NewAuthConfig()

			// Assert
			assert.Equal(t, tt.want.JWTSecret, got.JWTSecret)
			assert.Equal(t, tt.want.PublicMethods, got.PublicMethods)
			assert.Equal(t, tt.want.AdminOnlyMethods, got.AdminOnlyMethods)
		})
	}
}

func TestUnaryAuthInterceptor_PublicMethods(t *testing.T) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{testPublicMethod},
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

	interceptor := interceptors.UnaryAuthInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: testPublicMethod,
	}

	ctx := context.Background()
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	handler.AssertExpectations(t)
}

func TestUnaryAuthInterceptor_ValidToken(t *testing.T) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{testPublicMethod},
	}

	// Create valid token
	token := createTestToken(t, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour))

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)

		// Verify user context was added
		userClaims, ok := interceptors.GetUserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testUserID, userClaims.UserID)
		assert.Equal(t, testUsername, userClaims.Username)

		userID, ok := interceptors.GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testUserID, userID)

		username, ok := interceptors.GetUsernameFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testUsername, username)
	})

	interceptor := interceptors.UnaryAuthInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: testPrivateMethod,
	}

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	handler.AssertExpectations(t)
}

func TestUnaryAuthInterceptor_AdminMethod(t *testing.T) {
	tests := []struct {
		name          string
		roles         []string
		expectError   bool
		expectedCode  codes.Code
	}{
		{
			name:         "admin role allowed",
			roles:        []string{"admin"},
			expectError:  false,
		},
		{
			name:         "administrator role allowed",
			roles:        []string{"administrator"},
			expectError:  false,
		},
		{
			name:         "user role denied",
			roles:        []string{"user"},
			expectError:  true,
			expectedCode: codes.PermissionDenied,
		},
		{
			name:         "no roles denied",
			roles:        []string{},
			expectError:  true,
			expectedCode: codes.PermissionDenied,
		},
		{
			name:         "multiple roles with admin",
			roles:        []string{"user", "admin", "viewer"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &interceptors.AuthConfig{
				JWTSecret:        testSecret,
				PublicMethods:    []string{testPublicMethod},
				AdminOnlyMethods: []string{testAdminMethod},
			}

			token := createTestToken(t, testSecret, testUserID, testUsername, tt.roles, time.Now().Add(time.Hour))

			handler := &mockUnaryHandler{}
			if !tt.expectError {
				handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)
			}

			interceptor := interceptors.UnaryAuthInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: testAdminMethod,
			}

			md := metadata.Pairs("authorization", "Bearer "+token)
			ctx := metadata.NewIncomingContext(context.Background(), md)
			req := "request"

			resp, err := interceptor(ctx, req, info, handler.Handle)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "response", resp)
				handler.AssertExpectations(t)
			}
		})
	}
}

func TestUnaryAuthInterceptor_InvalidTokens(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedCode  codes.Code
		expectedMsg   string
	}{
		{
			name: "missing metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "missing metadata",
		},
		{
			name: "missing authorization header",
			setupContext: func() context.Context {
				md := metadata.Pairs("other-header", "value")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "missing authorization header",
		},
		{
			name: "invalid authorization header format",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Invalid token")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "invalid authorization header format",
		},
		{
			name: "empty token",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer ")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "empty token",
		},
		{
			name: "malformed token",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer invalid.token.format")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "expired token",
			setupContext: func() context.Context {
				token := createTestToken(t, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(-time.Hour))
				md := metadata.Pairs("authorization", "Bearer "+token)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token is expired",
		},
		{
			name: "token with wrong signing method",
			setupContext: func() context.Context {
				// Create token with RS256 instead of HS256
				claims := &interceptors.UserClaims{
					UserID:   testUserID,
					Username: testUsername,
					Roles:    []string{"user"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString([]byte(testSecret)) // This will be invalid for RS256
				md := metadata.Pairs("authorization", "Bearer "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "token missing required claims",
			setupContext: func() context.Context {
				claims := &interceptors.UserClaims{
					UserID: "", // Missing UserID
					Username: testUsername,
					Roles:    []string{"user"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(testSecret))
				md := metadata.Pairs("authorization", "Bearer "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "invalid token: missing required claims",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &interceptors.AuthConfig{
				JWTSecret:     testSecret,
				PublicMethods: []string{testPublicMethod},
			}

			handler := &mockUnaryHandler{}
			interceptor := interceptors.UnaryAuthInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: testPrivateMethod,
			}

			ctx := tt.setupContext()
			req := "request"

			resp, err := interceptor(ctx, req, info, handler.Handle)

			assert.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedMsg != "" {
				assert.Contains(t, st.Message(), tt.expectedMsg)
			}
		})
	}
}

func TestStreamAuthInterceptor_PublicMethods(t *testing.T) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{testPublicMethod},
	}

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil)

	stream := &mockServerStream{}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamAuthInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: testPublicMethod,
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestStreamAuthInterceptor_ValidToken(t *testing.T) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{testPublicMethod},
	}

	token := createTestToken(t, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour))

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		wrappedStream := args.Get(1).(grpc.ServerStream)
		ctx := wrappedStream.Context()

		// Verify user context was added
		userClaims, ok := interceptors.GetUserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testUserID, userClaims.UserID)
		assert.Equal(t, testUsername, userClaims.Username)
	})

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	stream := &mockServerStream{ctx: ctx}

	interceptor := interceptors.StreamAuthInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: testPrivateMethod,
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestStreamAuthInterceptor_InvalidToken(t *testing.T) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{testPublicMethod},
	}

	handler := &mockStreamHandler{}

	stream := &mockServerStream{ctx: context.Background()}

	interceptor := interceptors.StreamAuthInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: testPrivateMethod,
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestValidateTokenString(t *testing.T) {
	tests := []struct {
		name        string
		tokenString string
		secret      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid token",
			tokenString: createTestToken(t, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour)),
			secret:      testSecret,
			expectError: false,
		},
		{
			name:        "expired token",
			tokenString: createTestToken(t, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(-time.Hour)),
			secret:      testSecret,
			expectError: true,
			errorMsg:    "token is expired",
		},
		{
			name:        "invalid token format",
			tokenString: "invalid.token.format",
			secret:      testSecret,
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name:        "wrong secret",
			tokenString: createTestToken(t, "wrong-secret", testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour)),
			secret:      testSecret,
			expectError: true,
			errorMsg:    "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := interceptors.ValidateTokenString(tt.tokenString, tt.secret)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUserID, claims.UserID)
				assert.Equal(t, testUsername, claims.Username)
			}
		})
	}
}

func TestWrappedServerStream_Context(t *testing.T) {
	originalCtx := context.Background()
	stream := &mockServerStream{}

	// Test the wrapped stream context method
	// Since wrappedServerStream is not exported, we test through the interceptor
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{},
	}

	token := createTestToken(t, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour))

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		wrappedStream := args.Get(1).(grpc.ServerStream)
		ctx := wrappedStream.Context()

		// Verify the context has user information
		_, ok := interceptors.GetUserFromContext(ctx)
		assert.True(t, ok)
	})

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(originalCtx, md)

	stream.ctx = ctx

	interceptor := interceptors.StreamAuthInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: testPrivateMethod,
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestContextHelpers(t *testing.T) {
	t.Run("GetUserFromContext", func(t *testing.T) {
		// Test with valid context
		claims := &interceptors.UserClaims{
			UserID:   testUserID,
			Username: testUsername,
			Roles:    []string{"user"},
		}
		ctx := context.WithValue(context.Background(), "user_claims", claims)

		result, ok := interceptors.GetUserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, claims, result)

		// Test with empty context
		result, ok = interceptors.GetUserFromContext(context.Background())
		assert.False(t, ok)
		assert.Nil(t, result)

		// Test with wrong type
		ctx = context.WithValue(context.Background(), "user_claims", "wrong-type")
		result, ok = interceptors.GetUserFromContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("GetUserIDFromContext", func(t *testing.T) {
		// Test with valid context
		ctx := context.WithValue(context.Background(), "user_id", testUserID)

		result, ok := interceptors.GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testUserID, result)

		// Test with empty context
		result, ok = interceptors.GetUserIDFromContext(context.Background())
		assert.False(t, ok)
		assert.Empty(t, result)

		// Test with wrong type
		ctx = context.WithValue(context.Background(), "user_id", 123)
		result, ok = interceptors.GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Empty(t, result)
	})

	t.Run("GetUsernameFromContext", func(t *testing.T) {
		// Test with valid context
		ctx := context.WithValue(context.Background(), "username", testUsername)

		result, ok := interceptors.GetUsernameFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testUsername, result)

		// Test with empty context
		result, ok = interceptors.GetUsernameFromContext(context.Background())
		assert.False(t, ok)
		assert.Empty(t, result)

		// Test with wrong type
		ctx = context.WithValue(context.Background(), "username", 123)
		result, ok = interceptors.GetUsernameFromContext(ctx)
		assert.False(t, ok)
		assert.Empty(t, result)
	})
}

// Benchmark tests
func BenchmarkUnaryAuthInterceptor_PublicMethod(b *testing.B) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{testPublicMethod},
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := interceptors.UnaryAuthInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: testPublicMethod,
	}

	ctx := context.Background()
	req := "request"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkUnaryAuthInterceptor_ValidToken(b *testing.B) {
	config := &interceptors.AuthConfig{
		JWTSecret:     testSecret,
		PublicMethods: []string{},
	}

	token := createTestToken(b, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := interceptors.UnaryAuthInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: testPrivateMethod,
	}

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := "request"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkValidateTokenString(b *testing.B) {
	token := createTestToken(b, testSecret, testUserID, testUsername, []string{"user"}, time.Now().Add(time.Hour))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptors.ValidateTokenString(token, testSecret)
	}
}

// Helper functions
func createTestToken(t testing.TB, secret, userID, username string, roles []string, expiry time.Time) string {
	claims := &interceptors.UserClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}