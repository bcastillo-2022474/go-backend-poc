# Class Backend POC - Complete Architecture Documentation

## Overview

This document provides a comprehensive overview of the **Class Backend POC**, a monolithic Go backend implementing **Hexagonal Architecture** with **Domain-Driven Design** patterns. The system features multi-tenant RBAC authorization, RESTful APIs via Huma + Gin, and production-ready infrastructure patterns.

**Architecture Style**: Hexagonal Architecture with Domain-Driven Design  
**Deployment Pattern**: Monolithic with modular internal structure  
**API Style**: RESTful HTTP with automatic OpenAPI 3.1 documentation  
**Authorization**: Multi-tenant RBAC using Casbin  

## Technology Stack

### Core Framework & Libraries
| Component | Technology | Version     | Purpose |
|-----------|------------|-------------|---------|
| **Runtime** | Go | 1.24+       | Primary programming language |
| **HTTP Framework** | Huma + Gin | Latest      | Schema-first REST API with auto-generated docs |
| **Database** | PostgreSQL | 16 | Primary data store |
| **Database Access** | SQLC + pgx/v5 | Latest      | Type-safe SQL with high-performance driver |
| **Migrations** | Atlas CLI | Latest      | Modern schema-as-code migrations |
| **Authorization** | Casbin | v2          | Multi-tenant RBAC engine |
| **Validation** | go-playground/validator | v10         | Comprehensive input validation |
| **Error Handling** | cockroach/errors | Latest      | Rich error context and stack traces |
| **UUID Generation** | google/uuid | Latest      | Standard UUID v4 generation |

### Development & Operations
| Component | Technology | Purpose |
|-----------|------------|---------|
| **Containerization** | Docker + docker-compose | Local development environment |
| **Code Generation** | SQLC + go generate | Type-safe database access |
| **Testing** | testify + mockery | Unit testing with generated mocks |
| **Documentation** | Auto-generated OpenAPI 3.1 | Living API documentation |

## Project Structure

```
class-backend-poc/
├── core/                           # Domain & Application Layer (Hexagon Center)
│   ├── app/                        # Application Services & Use Cases
│   │   ├── auth/                   # Authentication domain
│   │   │   ├── application/        # Auth use cases
│   │   │   │   └── use-cases/
│   │   │   │       └── signup-use-case/
│   │   │   │           ├── signup-use-case-input.go    # Input validation & structure
│   │   │   │           └── signup-use-case.go   # Business orchestration
│   │   │   └── domain/             # Auth domain entities (future)
│   │   ├── shared/                 # Cross-cutting application concerns
│   │   │   ├── errors/             # Application error types & handling
│   │   │   │   ├── application-errors.go
│   │   │   │   ├── constants.go
│   │   │   │   └── propagate-errors.go
│   │   │   └── utils/              # Application utilities
│   │   │       └── validator-readable-error-messages.go
│   │   └── user/                   # User management domain
│   │       └── domain/
│   │           ├── entities/       # Core business entities
│   │           │   └── user.go
│   │           ├── errors/         # Domain-specific errors
│   │           │   └── user-errors.go
│   │           └── ports/          # Repository interfaces (ports)
│   │               └── user-ports.go
│   └── tests/                      # Domain & Application tests
│       ├── app/                    # Use case tests
│       │   └── auth/
│       │       └── application/
│       │           └── use-cases/
│       │               └── signup_use_case_test.go
│       └── mocks/                  # Generated test mocks
│           └── mock_user_repository.go
├── infra/                          # Infrastructure Layer (Hexagon Adapters)
│   ├── main.go                     # Application entry point & dependency injection
│   ├── auth/                       # Auth infrastructure
│   │   ├── adapters/               # Auth repository implementations (empty)
│   │   ├── handlers/               # HTTP handlers (empty - future)
│   │   └── sql/                    # Database schema & queries
│   │       ├── queries/            # SQL queries (empty)
│   │       └── schema.sql          # Auth database schema
│   ├── configs/                    # Configuration files
│   │   └── rbac_model.conf         # Casbin RBAC model definition
│   ├── shared/                     # Shared infrastructure components
│   │   ├── authorization/          # Multi-tenant RBAC implementation
│   │   │   ├── casbin_service.go   # Authorization service wrapper
│   │   │   ├── policy_loader.go    # YAML policy loader
│   │   │   └── role_adapter.go     # Custom Casbin DB adapter
│   │   └── utils/                  # Infrastructure utilities
│   │       └── application-error-to-http.go  # Error translation
│   └── user/                       # User infrastructure
│       ├── adapters/               # Repository implementations
│       │   └── user-repository-adapter.go
│       ├── handlers/               # HTTP handlers (empty - future)
│       └── sql/                    # Database schema & queries
│           ├── queries/
│           │   └── users.sql       # User SQL queries
│           └── schema.sql          # User database schema
├── generated/                      # Generated code (git ignored)
│   └── sqlc/                       # SQLC generated database access
│       ├── db.go
│       ├── models.go
│       ├── querier.go
│       └── users.sql.go
├── migrations/                     # Database migrations (Atlas)
│   ├── 20250817205744.sql
│   ├── 20250818013945_add_casbin_table.sql
│   └── atlas.sum
├── docs/                           # Architecture & design documentation
│   ├── ADRs/                       # Architecture Decision Records
│   │   ├── ADR-001-api-framework-selection-huma-gin.md
│   │   └── ADR-002-centralized-sqlc-code-generation.md
│   ├── architecture-comparison.md
│   ├── authorization-architecture.md
│   └── domain-modeling-patterns.md
├── configs/                        # Application configuration
├── policies.yaml                   # RBAC policies definition
├── sqlc.yaml                       # SQLC configuration
├── atlas.hcl                       # Database migration configuration
├── docker-compose.yml              # Development environment
└── README.md                       # Project overview
```

## Hexagonal Architecture Implementation

### Core Principles

1. **Dependency Inversion**: Core business logic (center) depends only on abstractions
2. **Port & Adapter Pattern**: Clear interfaces between layers
3. **Business Logic Isolation**: Domain logic independent of infrastructure concerns
4. **Testability**: Easy to test business logic with mocked dependencies

### Layer Responsibilities

#### Domain Layer (`core/app/*/domain/`)
**Purpose**: Pure business logic with no external dependencies

```go
// Domain Entity with behavior and validation
type User struct {
    // Private fields - encapsulation enforced
    id        string    `validate:"required,uuid4"`
    name      string    `validate:"required,min=2,max=100"`
    email     string    `validate:"required,email"`
    isActive  bool
    createdAt time.Time `validate:"required"`
    updatedAt time.Time `validate:"required"`
}

// Rich domain model with explicit state change methods
func (u *User) Activate() error {
    if u.isActive {
        return errors.New("user already active")
    }
    u.isActive = true
    u.updatedAt = time.Now().UTC()
    return u.Validate() // Validation enforced on every state change
}
```

**Key Characteristics:**
- No external dependencies (HTTP, database, etc.)
- Rich entities with behavior, not anemic data holders
- Private fields with explicit state change methods
- Validation enforced on every state transition
- Repository interfaces (ports) defined by domain needs

#### Application Layer (`core/app/*/application/`)
**Purpose**: Use cases and business workflow orchestration

```go
// Use Case with explicit input validation
type SignupUseCaseInput struct {
    Name     string `validate:"required,min=2,max=100"`
    Email    string `validate:"required,email,max=255"`
    Password string `validate:"required,min=8,max=128"`
    TenantID string `validate:"required,uuid4"`
}

// Use Case implementation (Transaction Script pattern)
func (uc *CreateUserUseCase) Execute(cmd *SignupUseCaseInput) (*entities.User, error) {
    // 1. Business validation
    exists, err := uc.userRepo.ExistsByEmail(cmd.Email)
    if err != nil {
        return nil, errors.PropagateError(err)
    }
    if exists {
        return nil, userErrors.NewEmailAlreadyExistsError(cmd.Email)
    }

    // 2. Entity creation through constructor
    user, err := entities.NewUser(cmd.Name, cmd.Email)
    if err != nil {
        return nil, errors.PropagateError(err)
    }

    // 3. Persistence
    return uc.userRepo.Save(user, cmd.Password)
}
```

**Key Characteristics:**
- Transaction Script pattern by default (pragmatic choice)
- Domain Services only when complexity/reuse justifies it
- Input validation with comprehensive command objects
- Orchestrates domain objects and external services
- Thin layer focused on workflow coordination

#### Infrastructure Layer (`infra/`)
**Purpose**: External system adapters and technical concerns

```go
// Repository Adapter implementing domain port
type PostgresUserRepository struct {
    db      *pgxpool.Pool
    queries *db.Queries  // SQLC generated
}

func (r *PostgresUserRepository) Save(user *entities.User, password string) (*entities.User, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, appErrors.PropagateError(err)
    }

    dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
        ID:           pgtype.UUID{Bytes: uuid.MustParse(user.ID()), Valid: true},
        Name:         user.Name(),
        Email:        user.Email(),
        PasswordHash: string(hashedPassword),
    })

    if err != nil {
        return nil, appErrors.PropagateError(err)
    }

    // Load existing user from database data
    return entities.LoadUser(
        dbUser.ID.String(),
        dbUser.Name,
        dbUser.Email,
        dbUser.CreatedAt.Time,
        dbUser.UpdatedAt.Time,
    )
}
```

## Domain Modeling Patterns

### Entity Design Philosophy

**Rich Domain Models with Encapsulation:**
- Private fields with public getters
- Explicit state change methods with validation
- Constructor pattern: `NewUser()` for business creation, `LoadUser()` for DB reconstruction
- No anemic entities - all entities have behavior

### Repository Design

**Interface Segregation Principle:**
```go
// Segregated interfaces instead of monolithic ones
type UserReader interface {
    FindByID(ctx context.Context, id string) (*entities.User, error)
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type UserWriter interface {
    Save(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id string) error
}

// Composed interfaces for specific needs
type UserRepository interface {
    UserReader
    UserWriter
}
```

### Business Logic Placement

**Decision Criteria:**
- **Transaction Script** (default): Simple operations, 1-2 use cases, straightforward validation
- **Domain Services**: Complex logic spanning multiple entities, 3+ reuse locations, multiple repository calls

## API Architecture

### RESTful Design with Huma + Gin

**Schema-First Development:**
- Go structs define API contracts
- Automatic OpenAPI 3.1 generation
- Built-in request/response validation
- Interactive Swagger UI at `/docs`

**HTTP Layer Responsibilities:**
```go
// HTTP Request structure (infrastructure layer)
type CreateUserHTTPRequest struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
    TenantID string `header:"X-Tenant-Id" validate:"required,uuid4"`
}

// Application Use Case Input (application layer)  
type SignupUseCaseInput struct {
    Name     string `validate:"required,min=2,max=100"`
    Email    string `validate:"required,email,max=255"`
    Password string `validate:"required,min=8,max=128"`
    TenantID string `validate:"required,uuid4"`
}
```

### Validation Strategy

**Three-Layer Validation:**
1. **HTTP Layer**: Basic structural validation (required fields, format)
2. **Application Layer**: Business constraints and policy validation
3. **Domain Layer**: Entity invariants and business rule validation

## Authorization Architecture

### Multi-Tenant RBAC with Casbin

**Key Components:**
- **Casbin Service**: Authorization engine wrapper
- **Policy-as-Code**: RBAC rules defined in `policies.yaml` (might exist on DB only in future iterations)
- **Role Adapter**: Custom database adapter for user-role assignments
- **Tenant Isolation**: Each tenant operates in isolated authorization domain

**Authorization Flow:**
```
HTTP Request → Middleware → Casbin Check → Allow/Deny
```

**RBAC Model:**
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

**Tenant Isolation:**
- Every request requires `X-Tenant-Id` header
- Role assignments scoped per tenant
- No cross-tenant access possible

## Database Architecture

### Schema Management with Atlas

**Migration Strategy:**
- Schema-as-code approach with Atlas CLI
- Version-controlled migrations in `migrations/`
- Environment-specific configurations (`local`, `dev`, `prod`)

**Current Schema:**
```sql
-- User Management
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Authorization (Casbin)
CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100),
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100)
);
```

### Type-Safe Database Access with SQLC

**Centralized Code Generation:**
```yaml
# sqlc.yaml - Single configuration for all modules
version: "2"
sql:
  - engine: "postgresql"
    schema:
      - "infra/user/sql/schema.sql"
      - "infra/auth/sql/schema.sql"
    queries: "infra/*/sql/queries"
    gen:
      go:
        out: "generated/sqlc"
        package: "db"
        sql_package: "pgx/v5"
        emit_interface: true
```

**Benefits:**
- Compile-time type safety
- Zero runtime reflection
- Cross-module query support
- Single source of truth for database operations

## Error Handling Strategy

### Structured Error Architecture

```go
// Rich application errors with context
type ApplicationError interface {
    GetCode() string
    GetMessage() string
    GetContext() map[string]interface{}
    GetOccurredAt() time.Time
    Unwrap() error
}

// Domain-specific error example
func NewEmailAlreadyExistsError(email string) *EmailAlreadyExistsError {
    return &EmailAlreadyExistsError{
        BaseApplicationError: BaseApplicationError{
            Code:       "EMAIL_ALREADY_EXISTS",
            Message:    "A user with this email already exists",
            Context:    map[string]interface{}{"email": email},
            OccurredAt: time.Now(),
        },
        Email: email,
    }
}
```

### Error Translation Layers

- **Domain**: Business rule violations
- **Application**: Use case orchestration errors  
- **Infrastructure**: Technical errors (database, HTTP, etc.)
- **HTTP**: Translated to appropriate status codes with safe error messages

## Development Workflow

### Local Development Setup

```bash
# 1. Start dependencies
docker-compose up -d

# 2. Apply database migrations
atlas migrate apply --env local

# 3. Generate database access code
sqlc generate

# 4. Run development server
go run infra/main.go
```

### Code Generation Workflow

```bash
# Generate SQLC database access code
sqlc generate

# Generate test mocks
go generate ./...

# Update dependencies
go mod tidy
```

### Testing Strategy

**Unit Testing:**
- Domain logic with isolated business rules
- Use case testing with mocked repositories
- Value object validation testing

**Integration Testing:**
- Database integration with real PostgreSQL
- HTTP API testing with full request/response cycle
- Authorization flow testing

**Test Organization:**
```
core/tests/
├── app/                    # Application layer tests
│   └── auth/
│       └── application/
│           └── use-cases/
└── mocks/                  # Generated mocks
    └── mock_user_repository.go
```

## Configuration Management

### Environment-Based Configuration

```bash
# .env example
DATABASE_URL=postgres://postgres:postgres@localhost:5437/edoo_class?sslmode=disable
HTTP_PORT=8081
JWT_SECRET=your-secret-here
```

### Multi-Environment Support

- **Local**: Docker Compose development environment
- **Dev**: Development deployment configuration
- **Prod**: Production deployment settings

## Deployment Architecture

### Monolithic Deployment Benefits

**Advantages:**
- **Operational Simplicity**: Single deployment unit
- **ACID Transactions**: Cross-domain operations in single database
- **Fast Development**: No distributed system complexity
- **Resource Efficiency**: Shared resources and connections

### Container Strategy

```dockerfile
# Multi-stage build for production
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app ./infra/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app"]
```

## Future Architecture Considerations

## Conclusion

The Class Backend successfully demonstrates a production-ready monolithic architecture with clean separation of concerns, comprehensive security, and excellent maintainability. The hexagonal architecture provides a solid foundation for both current MVP needs and future scaling requirements.

Key architectural strengths:
- **Clean boundaries** between business logic and infrastructure
- **Pragmatic patterns** balancing complexity with development velocity
- **Production-ready infrastructure** with proper tooling and automation
- **Security-first design** with comprehensive multi-tenant RBAC
- **Evolution-friendly structure** supporting both monolithic growth and future microservices extraction

The architecture successfully balances the team's constraints (2 developers, tight timeline) with long-term architectural quality, providing a sustainable foundation for the educational platform's growth.