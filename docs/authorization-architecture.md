# Authorization Architecture

This document describes the authorization system implemented in this microservice, built using Casbin for multi-tenant Role-Based Access Control (RBAC).

## Overview

The authorization system provides multi-tenant RBAC with the following key characteristics:

- **Middleware-based**: Authorization is handled at the middleware layer, not mixed into business logic
- **Hybrid storage**: Policies are version-controlled in YAML, role assignments are persisted in database
- **Multi-tenant**: Each tenant has isolated authorization domains
- **Clean separation**: Business logic handlers focus on core functionality, middleware handles authorization
- **Casbin-powered**: Uses Casbin v2 for flexible, scalable authorization engine

## Architecture Components

### 1. Casbin Service (`class/shared/authorization/casbin_service.go`)

The main authorization service wrapper that provides a clean API:

- `CanDo(userID, resource, action, tenantID)` - Check if user can perform action
- `AssignRole(userID, role, tenantID)` - Assign role to user in tenant
- `RemoveRole(userID, role, tenantID)` - Remove role assignment

**Design Decision**: Wraps Casbin's complex API with domain-specific methods to simplify usage and provide consistent error handling.

### 2. Authorization Middleware (`class/shared/authorization/middleware.go`)

gRPC interceptor that automatically checks authorization for all endpoints:

```go
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        authorization.AuthorizationInterceptor(authzService),
        loggingInterceptor,
    ),
)
```

**Endpoint Mapping**: Maps gRPC methods to resource+action combinations:
```go
var EndpointMapping = map[string]ResourceAction{
    "/auth.v1.AuthService/Signup": {Resource: "user", Action: "create"},
}
```

**Design Decision**: Middleware approach ensures:
- Authorization is enforced consistently across all endpoints
- Business logic handlers remain focused on core functionality
- Authorization logic is centralized and easier to audit
- Can't accidentally forget to add authorization checks

### 3. Hybrid Policy Storage

#### Policies in YAML (`policies.yaml`)
```yaml
roles:
  admin:
    permissions:
      all: [all]  # Admin can do everything
  instructor:
    permissions:
      assignment: [create, view, edit, grade]
      course: [view, edit]
      student: [view]
      grade: [assign, view]
  student:
    permissions:
      assignment: [view, submit]
      course: [view]
      grade: [view]
      profile: [edit]
```

**Design Decisions**:
- **Version controlled**: Policies are in Git, enabling change tracking and code review
- **Consistent across tenants**: Same role definitions for all tenants (for now)
- **Human readable**: Uses "all" syntax instead of "*" wildcards for clarity
- **Memory-based**: Loaded at startup for fast authorization checks

#### Role Assignments in Database (`casbin_rule` table)
- Stores user-role-tenant mappings dynamically
- Supports runtime role assignment/removal
- Persisted for durability across restarts

**Design Decision**: Hybrid approach balances:
- **Static policies**: Version controlled, consistent, auditable
- **Dynamic assignments**: Flexible user management at runtime

### 4. Custom Database Adapter (`class/shared/authorization/role_adapter.go`)

`RoleOnlyPostgresAdapter` extends sql-adapter to implement hybrid storage:

- **Only persists role assignments** (g records) to database
- **Skips policy persistence** (p records) - these stay in memory
- **Selective operations**: AddPolicy, RemovePolicy only work on grouping policies

**Design Decision**: Custom adapter ensures policies never get accidentally persisted to database, maintaining the intended hybrid architecture.

### 5. Policy Loader (`class/shared/authorization/policy_loader.go`)

Converts YAML policies to Casbin format:

- **"all" → "*" conversion**: Makes YAML more readable while supporting Casbin wildcards
- **Tenant expansion**: Creates explicit policies for each tenant domain
- **Validation**: Ensures policy structure is correct before loading

**Design Decision**: Abstraction layer allows human-friendly YAML while maintaining Casbin compatibility.

### 6. RBAC Model Configuration (`configs/rbac_model.conf`)

Casbin model definition with wildcard support:

```conf
[matchers]
m = g(r.sub, p.sub, r.dom) && (r.obj == p.obj || p.obj == "*") && (r.act == p.act || p.act == "*") && r.dom == p.dom
```

**Design Decision**: Supports both specific permissions and wildcard permissions, enabling both fine-grained and broad access patterns.

## Authorization Flow

1. **Request arrives** at gRPC server
2. **Authorization middleware** intercepts the request
3. **Endpoint mapping** determines required resource+action
4. **User/tenant extraction** from request context (placeholder - JWT middleware will handle this)
5. **Authorization check** via `CasbinService.CanDo()`
6. **Allow/deny** request based on result

## Multi-Tenant Design

Each tenant operates in its own authorization domain:

- **Tenant isolation**: Users in tenant A cannot access tenant B resources
- **Domain-specific roles**: Same role name can have different permissions per tenant (future enhancement)
- **No global permissions**: No wildcards across tenants for security

**Design Decision**: Explicit tenant domains prevent accidental cross-tenant access and support future tenant-specific customizations.

## Security Considerations

### 1. No Global Wildcards
- Admin roles are tenant-specific only
- Prevents accidental super-user permissions across all tenants

### 2. Middleware Enforcement
- Authorization cannot be bypassed - every request goes through middleware
- Centralized policy enforcement point

### 3. Policy Separation
- Static policies in version control enable audit trails
- Dynamic assignments in database for operational flexibility

### 4. Fail-Safe Defaults
- Unknown endpoints are denied by default
- Missing user/tenant information results in denial

## Future Enhancements

### 1. JWT Integration
Replace placeholder user/tenant extraction with proper JWT middleware:
```go
// TODO: Replace with actual JWT middleware
userID := extractUserFromJWT(ctx)
tenantID := extractTenantFromJWT(ctx)
```

### 2. Tenant-Specific Policies
Support tenant-specific policy customizations while maintaining YAML defaults:
- Base policies from YAML
- Tenant overrides in database
- Fallback chain: tenant-specific → default → deny

### 3. Dynamic Policy Management
API endpoints for runtime policy management (admin only):
- Update role permissions
- Create custom roles per tenant
- Policy validation and rollback

### 4. Authorization Caching
Cache authorization decisions for frequently accessed user-resource combinations:
- Redis-based cache
- Cache invalidation on role changes
- Performance optimization for high-throughput scenarios

### 5. Audit Logging
Comprehensive audit trail for authorization decisions:
- Who accessed what, when
- Authorization failures and reasons
- Policy changes and who made them

## Testing Strategy

### Unit Tests
- Test individual components (CasbinService, PolicyLoader, etc.)
- Mock database interactions
- Verify policy loading and conversion logic

### Integration Tests
- Test complete authorization flow end-to-end
- Real database with test data
- Verify middleware integration with gRPC

### Security Tests
- Attempt to bypass authorization
- Test cross-tenant access attempts
- Verify fail-safe behaviors

## Maintenance

### Adding New Endpoints
1. Add endpoint mapping in `middleware.go`
2. Define resource+action semantics
3. Update policies in `policies.yaml` if new permissions needed
4. Add tests for the new endpoint authorization

### Modifying Permissions
1. Update `policies.yaml`
2. Restart service to reload policies
3. Test changed permissions
4. Consider migration for existing role assignments if needed

### Troubleshooting
- Check logs for authorization decisions
- Verify endpoint mappings are correct
- Ensure user/tenant extraction is working
- Validate policy YAML syntax and loading

This architecture provides a solid foundation for scalable, secure, multi-tenant authorization while maintaining clean separation of concerns and operational flexibility.