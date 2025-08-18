# Class Backend POC

## TLDR

A Go monolith demonstrating **Clean Architecture** + **Hexagonal Architecture** + **Multi-Tenant RBAC Authorization**. Features user authentication, role-based permissions (admin/instructor/student), and comprehensive security with Casbin. Exposes APIs via both gRPC and HTTP REST with automatic authorization enforcement.

**Tech Stack:** Go, gRPC, HTTP REST, PostgreSQL, SQLC, Atlas, Casbin RBAC, JWT, Protocol Buffers, Docker, Testify

---

## Architecture Overview

This project demonstrates a modern Go monolith using **Clean Architecture** (Domain-Driven Design) combined with **Hexagonal Architecture** principles and **Multi-Tenant RBAC Authorization**. The codebase showcases enterprise-grade patterns with comprehensive security, role-based permissions, and tenant isolation - all within a well-structured monolithic architecture.

### Core Architectural Principles

- **Domain-First Design**: Business logic lives in the core domain layer
- **Dependency Inversion**: Core domain has no dependencies on infrastructure  
- **Interface Segregation**: Clean port/adapter pattern for external dependencies
- **Authorization as Middleware**: Security enforced at infrastructure layer, not business logic
- **Multi-Tenant Security**: Complete tenant isolation with role-based permissions
- **Error Handling**: Rich error types with stack traces and proper boundary handling
- **Type Safety**: Generated code for database queries and API contracts

## Project Structure

```
├── class/                           # Infrastructure Layer (Hexagonal Architecture "adapters")
│   ├── main.go                     # Application entry point with dual gRPC+HTTP servers
│   ├── auth/handlers/              # gRPC handlers (interface adapters)
│   ├── user/adapters/              # Database adapters (PostgreSQL implementation)
│   └── shared/                     # Infrastructure utilities
│       ├── authorization/          # Casbin authorization system
│       │   ├── casbin_service.go   # Main authorization service
│       │   ├── middleware.go       # gRPC authorization middleware
│       │   ├── policy_loader.go    # YAML policy loader
│       │   └── role_adapter.go     # Custom Casbin adapter
│       └── utils/                  # Infrastructure utilities (error conversion, etc.)
├── core/                           # Domain + Application Layer (Clean Architecture core)
│   ├── shared/errors/              # Domain error types and propagation
│   ├── auth/application/           # Auth use cases and commands
│   │   └── use-cases/signup-use-case/  # Signup business logic + tests
│   └── user/domain/                # User domain entities, ports, and business rules
│       ├── entities/               # Domain entities (User)
│       ├── ports/                  # Repository interfaces (dependency inversion)
│       └── errors/                 # Domain-specific errors
├── configs/                        # Configuration files
│   └── rbac_model.conf            # Casbin RBAC model configuration
├── policies.yaml                   # Role-based authorization policies
├── proto/                          # Schema-First API Design
│   ├── auth/v1/                    # Auth service definitions
│   ├── common/v1/                  # Shared proto messages
│   └── generated/                  # Generated gRPC + HTTP gateway code
├── migrations/                     # Database schema migrations (Atlas)
└── docs/                          # Documentation
    ├── authorization-architecture.md  # Authorization system documentation
    └── demo-implementation-plan.md   # Demo development roadmap
```

## Tech Stack

### Core Technologies

- **Go** - Primary language
- **gRPC** - High-performance RPC framework
- **HTTP REST** - Auto-generated from gRPC via gRPC-Gateway
- **PostgreSQL** - Primary database
- **Docker** - Containerization and development environment

### Development Tools

- **Protocol Buffers** - Schema-first API design
- **SQLC** - Type-safe SQL code generation
- **Atlas** - Database schema migrations
- **Casbin** - Authorization engine for RBAC
- **Testify** - Testing framework with mocking
- **Cockroach Errors** - Stack trace preservation
- **Buf** - Protocol buffer toolchain

## Key Features

### 🏗️ Clean Architecture Implementation

**Domain Layer** (`core/`):
- Pure business logic with no external dependencies
- Rich domain entities with validation
- Repository ports (interfaces) for dependency inversion

**Application Layer** (`core/auth/application/`):
- Use cases orchestrate business workflows
- Command objects with validation
- Application-specific error handling

**Infrastructure Layer** (`class/`):
- Database adapters implementing domain ports
- gRPC handlers converting between protocols
- Error conversion utilities

### 🛡️ Multi-Tenant RBAC Authorization

**Role-Based Permissions**:
- **Admin** - Full access to all resources within tenant
- **Instructor** - Course and assignment management, student viewing
- **Student** - Assignment submission, grade viewing, profile management

**Security Features**:
- Automatic authorization enforcement via middleware
- Complete tenant isolation (no cross-tenant access)
- Hybrid policy storage (YAML + database)
- Casbin-powered authorization engine

### 🔄 Dual API Exposure

Every endpoint is automatically available as:
- **gRPC** - High-performance binary protocol
- **HTTP REST** - JSON over HTTP via gRPC-Gateway  
- **OpenAPI** - Auto-generated documentation

### 🛡️ Comprehensive Error Handling

- **Domain Errors** - Business rule violations (e.g., email already exists)
- **Infrastructure Errors** - Technical failures (database, network)
- **Validation Errors** - Input validation failures
- **Stack Traces** - Full error context preserved with Cockroach errors

### 🧪 Testing Strategy

- **Unit Tests** - Use case testing with mocked dependencies
- **Type Safety** - All dependencies injected via interfaces
- **Factory Functions** - Validated object creation in tests
- **Error Scenarios** - Comprehensive error condition testing

## API Documentation

### Authentication Endpoints

**Signup (Public)**:
```bash
# HTTP REST
curl -X POST http://localhost:8081/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'

# gRPC  
grpcurl -plaintext -d '{"name":"John","email":"john@example.com","password":"password123"}' \
  localhost:8080 auth.v1.AuthService/Signup
```

**Login (Public)**:
```bash
# Returns JWT token for authorization
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'
```

### Protected Endpoints (Require Authorization)

**List Users (Admin Only)**:
```bash
curl -X GET http://localhost:8081/api/v1/users \
  -H "Authorization: Bearer <jwt-token>" \
  -H "X-Tenant-Id: tenant1"
```

**Assign Role (Admin Only)**:
```bash
curl -X POST http://localhost:8081/api/v1/auth/assign-role \
  -H "Authorization: Bearer <jwt-token>" \
  -H "X-Tenant-Id: tenant1" \
  -d '{"user_id":"user-123","role":"instructor"}'
```

**List Students (Instructor Only)**:
```bash
curl -X GET http://localhost:8081/api/v1/users/students \
  -H "Authorization: Bearer <jwt-token>" \
  -H "X-Tenant-Id: tenant1"
```

## Database Schema

Managed via Atlas migrations with type-safe SQLC queries:

**Users Table**:
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Authorization Table**:
```sql
CREATE TABLE casbin_rule (
  ptype VARCHAR(100) NOT NULL,   -- Policy type: 'g' for groupings
  v0 VARCHAR(100),               -- User ID
  v1 VARCHAR(100),               -- Role
  v2 VARCHAR(100),               -- Tenant ID
  v3 VARCHAR(100),               -- Reserved
  v4 VARCHAR(100),               -- Reserved  
  v5 VARCHAR(100)                -- Reserved
);
```

**Role Definitions** (policies.yaml):
```yaml
roles:
  admin:
    permissions:
      all: [all]
  instructor:
    permissions:
      assignment: [create, view, edit, grade]
      course: [view, edit]
      student: [view]
  student:
    permissions:
      assignment: [view, submit]
      course: [view]
      grade: [view]
```

## Development Setup

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Atlas CLI
- Buf CLI

### Quick Start

1. **Start Database:**
   ```bash
   docker-compose up -d
   ```

2. **Run Migrations:**
   ```bash
   atlas migrate apply --env local
   ```

3. **Generate Code:**
   ```bash
   # Generate gRPC code
   buf generate
   
   # Generate database queries
   cd class/user && sqlc generate
   ```

4. **Setup Authorization:**
   ```bash
   # Create admin user via signup
   curl -X POST http://localhost:8081/api/v1/auth/signup \
     -H "Content-Type: application/json" \
     -d '{"name":"Admin User","email":"admin@example.com","password":"password123"}'
   
   # Manually assign admin role (for demo purposes)
   # In production, this would be done via secure admin interface
   ```

5. **Run Server:**
   ```bash
   go run class/main.go
   ```

6. **Test Authorization Flow:**
   ```bash
   # 1. Login and get JWT token
   JWT=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"password123"}' | jq -r '.token')
   
   # 2. Use token for protected endpoints
   curl -X GET http://localhost:8081/api/v1/users \
     -H "Authorization: Bearer $JWT" \
     -H "X-Tenant-Id: tenant1"
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific module tests
go test ./core/auth/application/use-cases/signup-use-case/ -v
```

## Configuration

Environment variables (`.env`):

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5437/edoo_class?sslmode=disable
GRPC_PORT=8080
HTTP_PORT=8081
JWT_SECRET=your-secret-key-here
TENANTS=tenant1,tenant2,tenant3
```

## Security & Authorization

### Authorization Flow
1. **Middleware Interception** - All requests go through authorization middleware
2. **JWT Token Validation** - Extract user/tenant from JWT claims
3. **Permission Check** - Casbin evaluates user permissions for requested resource+action
4. **Tenant Isolation** - Ensure user can only access their tenant's data

### Role Hierarchy
- **Admin** → Full tenant access (user management, role assignment)
- **Instructor** → Course management, student viewing, assignment grading
- **Student** → Assignment submission, grade viewing, profile editing

### Security Features
- **Multi-tenant isolation** - Complete data separation per tenant
- **Role-based permissions** - Granular resource+action controls
- **Policy version control** - Authorization rules in Git via YAML
- **Audit logging** - All authorization decisions logged

## Monitoring & Observability

- **Authorization Logging** - All access attempts and decisions logged
- **Structured Logging** - Request/response logging with timing
- **Error Tracking** - Full stack traces for debugging  
- **Health Checks** - Database connectivity verification
- **Graceful Shutdown** - Clean server termination

## Future Enhancements

- [ ] JWT middleware implementation (replace header-based auth)
- [ ] Complete demo endpoints (login, list users, assign roles, list students)
- [ ] Comprehensive test suite (unit + integration + authorization tests)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] API documentation with Swagger/OpenAPI
- [ ] Performance optimization and caching
- [ ] Event sourcing for audit trails

## Contributing

This project follows Clean Architecture principles with RBAC authorization. When adding new features:

1. **Domain Layer**: Start with domain entities and business rules
2. **Application Layer**: Define use cases and command objects
3. **Infrastructure Layer**: Add repository adapters and handlers
4. **Authorization**: Add endpoint mappings and role permissions
5. **Testing**: Write comprehensive tests including authorization scenarios

### Authorization for New Endpoints:
1. Add endpoint mapping in `middleware.go`:
   ```go
   EndpointMapping["/service.Method"] = ResourceAction{Resource: "resource", Action: "action"}
   ```
2. Update `policies.yaml` if new permissions needed
3. Test authorization scenarios (allowed/denied cases)