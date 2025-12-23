package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor_Unary(t *testing.T) {
	jwtConfig := &security.JWTConfig{
		SecretKey:     "test-secret",
		Expiration:    time.Hour,
		RefreshExp:    7 * 24 * time.Hour,
		Issuer:        "test-issuer",
		SigningMethod: "HS256",
	}
	jwtManager, err := security.NewJWTManager(jwtConfig)
	if err != nil {
		t.Fatalf("NewJWTManager error: %v", err)
	}

	adminToken, err := jwtManager.GenerateToken(1, "admin", true)
	if err != nil {
		t.Fatalf("GenerateToken(admin) error: %v", err)
	}
	userToken, err := jwtManager.GenerateToken(2, "user", false)
	if err != nil {
		t.Fatalf("GenerateToken(user) error: %v", err)
	}

	publicMethods := map[string]struct{}{
		"/service/Login": {},
	}
	adminMethods := map[string]struct{}{
		"/service/DeleteCategory": {},
	}

	interceptor := AuthInterceptor(jwtManager, publicMethods, adminMethods)

	tests := []struct {
		name     string
		method   string
		token    string
		wantCode codes.Code
		wantNext bool
	}{
		{name: "Health check - no token - allowed", method: "/grpc.health.v1.Health/Check", token: "", wantCode: codes.OK, wantNext: true},
		{name: "Health watch - no token - allowed", method: "/grpc.health.v1.Health/Watch", token: "", wantCode: codes.OK, wantNext: true},
		{name: "Public method - no token - allowed", method: "/service/Login", token: "", wantCode: codes.OK, wantNext: true},
		{name: "Normal method - no token - unauthenticated", method: "/service/GetPost", token: "", wantCode: codes.Unauthenticated, wantNext: false},
		{name: "Normal method - invalid token - unauthenticated", method: "/service/GetPost", token: "Bearer invalid-token", wantCode: codes.Unauthenticated, wantNext: false},
		{name: "Normal method - user token - allowed", method: "/service/GetPost", token: "Bearer " + userToken, wantCode: codes.OK, wantNext: true},
		{name: "Admin method - user token - permission denied", method: "/service/DeleteCategory", token: "Bearer " + userToken, wantCode: codes.PermissionDenied, wantNext: false},
		{name: "Admin method - admin token - allowed", method: "/service/DeleteCategory", token: "Bearer " + adminToken, wantCode: codes.OK, wantNext: true},
		{name: "Normal method - admin token - allowed", method: "/service/GetPost", token: "Bearer " + adminToken, wantCode: codes.OK, wantNext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.token != "" {
				md := metadata.New(map[string]string{"authorization": tt.token})
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			// 模拟 ContextInterceptor：从 metadata 提取 authorization 到 contextx
			ctx = contextx.FromMD(ctx)

			var nextCalled bool
			mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				nextCalled = true
				return "success", nil
			}

			info := &grpc.UnaryServerInfo{FullMethod: tt.method}
			_, err := interceptor(ctx, nil, info, mockHandler)

			gotCode := status.Code(err)
			if gotCode != tt.wantCode {
				t.Fatalf("code=%v want=%v err=%v", gotCode, tt.wantCode, err)
			}
			if nextCalled != tt.wantNext {
				t.Fatalf("nextCalled=%v want=%v", nextCalled, tt.wantNext)
			}
		})
	}
}
