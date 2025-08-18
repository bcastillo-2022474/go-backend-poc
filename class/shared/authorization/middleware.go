package authorization

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ResourceAction represents a resource and action for authorization
type ResourceAction struct {
	Resource string
	Action   string
}

// EndpointMapping maps gRPC methods to resource+action combinations
var EndpointMapping = map[string]ResourceAction{
	"/auth.v1.AuthService/Signup": {Resource: "user", Action: "create"},
}

// PublicEndpoints defines endpoints that don't require authorization
var PublicEndpoints = map[string]bool{
	"/auth.v1.AuthService/Signup": true,
}

func AuthorizationInterceptor(authzService *CasbinService) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if endpoint is public
		if PublicEndpoints[info.FullMethod] {
			log.Printf("Public endpoint accessed: %s", info.FullMethod)
			return handler(ctx, req)
		}

		userID, tenantID, err := extractUserAndTenant(ctx)
		if err != nil {
			log.Printf("Failed to extract user/tenant for %s: %v", info.FullMethod, err)
			return nil, status.Errorf(codes.Unauthenticated, "authentication required")
		}

		resourceAction, exists := EndpointMapping[info.FullMethod]
		if !exists {
			log.Printf("No authorization mapping for endpoint: %s", info.FullMethod)
			return nil, status.Errorf(codes.Internal, "authorization mapping not configured")
		}

		allowed, err := authzService.CanDo(userID, resourceAction.Resource, resourceAction.Action, tenantID)
		if err != nil {
			log.Printf("Failed to check authorization for %s: %v", info.FullMethod, err)
			return nil, status.Errorf(codes.Internal, "authorization error")
		}
		if !allowed {
			log.Printf("Access denied: user=%s, resource=%s, action=%s, tenant=%s",
				userID, resourceAction.Resource, resourceAction.Action, tenantID)
			return nil, status.Errorf(codes.PermissionDenied,
				"insufficient permissions for %s.%s", resourceAction.Resource, resourceAction.Action)
		}

		log.Printf("Access granted: user=%s, resource=%s, action=%s, tenant=%s",
			userID, resourceAction.Resource, resourceAction.Action, tenantID)

		// Add authorization context to request context
		authCtx := WithAuthorizationContext(ctx, userID, tenantID, resourceAction.Resource, resourceAction.Action)
		return handler(authCtx, req)
	}
}

// extractUserAndTenant extracts user ID and tenant ID from gRPC metadata
func extractUserAndTenant(ctx context.Context) (userID, tenantID string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// Extract user ID from X-User-Id header
	userHeaders := md.Get("x-user-id")
	if len(userHeaders) == 0 {
		return "", "", status.Errorf(codes.Unauthenticated, "missing user ID")
	}
	userID = userHeaders[0]

	// Extract tenant ID from X-Tenant-Id header
	tenantHeaders := md.Get("x-tenant-id")
	if len(tenantHeaders) == 0 {
		return "", "", status.Errorf(codes.Unauthenticated, "missing tenant ID")
	}
	tenantID = tenantHeaders[0]

	if userID == "" || tenantID == "" {
		return "", "", status.Errorf(codes.Unauthenticated, "empty user ID or tenant ID")
	}

	return userID, tenantID, nil
}

type AuthContext struct {
	UserID   string
	TenantID string
	Resource string
	Action   string
}

type authContextKey struct{}

// WithAuthorizationContext adds authorization context to the request context
func WithAuthorizationContext(ctx context.Context, userID, tenantID, resource, action string) context.Context {
	authCtx := &AuthContext{
		UserID:   userID,
		TenantID: tenantID,
		Resource: resource,
		Action:   action,
	}
	return context.WithValue(ctx, authContextKey{}, authCtx)
}

// GetAuthorizationContext extracts authorization context from request context
func GetAuthorizationContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(authContextKey{}).(*AuthContext)
	return authCtx, ok
}

func AddEndpointMapping(method string, resource string, action string) {
	EndpointMapping[method] = ResourceAction{
		Resource: resource,
		Action:   action,
	}
}

func AddPublicEndpoint(method string) {
	PublicEndpoints[method] = true
}

func GenerateMethodName(service, method string) string {
	return "/" + service + "/" + method
}
