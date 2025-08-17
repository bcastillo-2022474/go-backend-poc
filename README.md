# Class Backend POC

## TLDR

A Go microservice built with **Clean Architecture** + **Hexagonal Architecture** patterns. Exposes user signup via both gRPC and HTTP REST APIs. Uses PostgreSQL with type-safe queries (SQLC), schema migrations (Atlas), and comprehensive error handling with full stack traces.

**Tech Stack:** Go, gRPC, HTTP REST, PostgreSQL, SQLC, Atlas, Protocol Buffers, Docker, Testify

---

## Architecture Overview

This project demonstrates a modern Go microservice using **Clean Architecture** (Domain-Driven Design) combined with **Hexagonal Architecture** principles. The codebase is organized around bounded contexts with clear separation of concerns.

### Core Architectural Principles

- **Domain-First Design**: Business logic lives in the core domain layer
- **Dependency Inversion**: Core domain has no dependencies on infrastructure
- **Interface Segregation**: Clean port/adapter pattern for external dependencies
- **Error Handling**: Rich error types with stack traces and proper boundary handling
- **Type Safety**: Generated code for database queries and API contracts

## Project Structure

```
├── class/                           # Infrastructure Layer (Hexagonal Architecture "adapters")
│   ├── main.go                     # Application entry point with dual gRPC+HTTP servers
│   ├── auth/handlers/              # gRPC handlers (interface adapters)
│   ├── user/adapters/              # Database adapters (PostgreSQL implementation)
│   └── shared/utils/               # Infrastructure utilities (error conversion, etc.)
├── core/                           # Domain + Application Layer (Clean Architecture core)
│   ├── shared/errors/              # Domain error types and propagation
│   ├── auth/application/           # Auth use cases and commands
│   │   └── use-cases/signup-use-case/  # Signup business logic + tests
│   └── user/domain/                # User domain entities, ports, and business rules
│       ├── entities/               # Domain entities (User)
│       ├── ports/                  # Repository interfaces (dependency inversion)
│       └── errors/                 # Domain-specific errors
├── proto/                          # Schema-First API Design
│   ├── auth/v1/                    # Auth service definitions
│   ├── common/v1/                  # Shared proto messages
│   └── generated/                  # Generated gRPC + HTTP gateway code
└── migrations/                     # Database schema migrations (Atlas)
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

### Signup Endpoint

**gRPC:**
```protobuf
service AuthService {
  rpc Signup(SignupRequest) returns (SignupResponse);
}
```

**HTTP REST:**
```
POST /api/v1/auth/signup
{
  "name": "John Doe",
  "email": "john@example.com", 
  "password": "password123"
}
```

## Database Schema

Managed via Atlas migrations with type-safe SQLC queries:

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

4. **Run Server:**
   ```bash
   go run class/main.go
   ```

5. **Test API:**
   ```bash
   # gRPC
   grpcurl -plaintext -d '{"name":"John","email":"john@example.com","password":"password123"}' localhost:8080 auth.v1.AuthService/Signup
   
   # HTTP REST
   curl -X POST http://localhost:8081/api/v1/auth/signup \
     -H "Content-Type: application/json" \
     -d '{"name":"John","email":"john@example.com","password":"password123"}'
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
```

## Monitoring & Observability

- **Structured Logging** - Request/response logging with timing
- **Error Tracking** - Full stack traces for debugging
- **Health Checks** - Database connectivity verification
- **Graceful Shutdown** - Clean server termination

## Future Enhancements

- [ ] Authentication middleware (JWT)
- [ ] Request tracing (OpenTelemetry)
- [ ] Metrics collection (Prometheus)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] API rate limiting
- [ ] Caching layer (Redis)
- [ ] Event sourcing/CQRS patterns

## Contributing

This project follows Clean Architecture principles. When adding new features:

1. Start with domain entities and business rules
2. Define repository ports (interfaces) 
3. Implement use cases (application layer)
4. Add infrastructure adapters
5. Write comprehensive tests

## License

MIT License - see LICENSE file for details.