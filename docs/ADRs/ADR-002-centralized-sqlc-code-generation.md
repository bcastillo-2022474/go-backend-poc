---
status: "proposed"
date: 2025-01-27
decision-makers: [Joao]
consulted: []
informed: []
---

# SQLC Configuration Strategy: Centralized Code Generation Architecture

## Context and Problem Statement

We are building a Go backend with clean/hexagonal architecture that requires type-safe database access code. SQLC provides excellent SQL-to-Go code generation capabilities, but we must decide on the optimal configuration strategy for organizing generated code across multiple domain modules (user, auth, etc.).

Two primary approaches are possible:

1. **Modular Approach**: Each domain module has its own `sqlc.yaml` configuration with generated code in module-specific directories
2. **Centralized Approach**: Single root `sqlc.yaml` configuration with all generated code in a shared directory

**Key Challenge**: Select an SQLC configuration strategy that supports cross-module database queries, minimizes maintenance overhead, and scales effectively as the application grows.

## Decision Drivers

* **Cross-Module Query Support** - Must enable complex queries spanning multiple database tables across domains
* **Schema Dependency Management** - Minimize configuration overhead when modules reference each other's schemas  
* **CI/CD Simplicity** - Code generation should be simple and fast in build pipelines
* **Maintenance Burden** - Configuration should not require constant updates as modules are added
* **Development Experience** - Developers should have consistent, predictable database access patterns
* **Generated Code Management** - Code generation artifacts should follow industry best practices
* **Scalability** - Architecture should handle growth from 2-3 modules to dozens of modules

## Considered Options

* **Centralized SQLC Configuration** - Single root configuration with shared generated code
* **Modular SQLC Configuration** - Individual module configurations with domain-specific generated code  
* **Hybrid Approach** - Centralized schema, modular generation targets
* **External Schema Management** - Separate schema repository with multiple generation targets

## Decision Outcome

Chosen option: **"Centralized SQLC Configuration"**, because it provides the best balance of cross-module query support, maintenance simplicity, and development experience.

Our implementation uses the following architecture:

```
project/
├── sqlc.yaml              # Single configuration at root
├── generated/
│   └── sqlc/              # All generated code centralized  
├── infra/
│   ├── user/sql/          # SQL queries (modularized by domain)
│   └── auth/sql/
└── .gitignore             # Excludes generated/ directory
```

### Architecture Details

1. **Single Configuration**: One `sqlc.yaml` at project root containing all module schemas and queries
2. **Centralized Output**: All generated code in `generated/sqlc/` with single package name
3. **Modular Input**: SQL queries remain organized by domain in respective modules
4. **Git Exclusion**: Generated code ignored via `.gitignore`
5. **Simple CI**: Single `sqlc generate` command

### Consequences

* **Good**, because cross-module database queries work naturally without complex schema imports
* **Good**, because single `sqlc generate` command handles all code generation in CI/CD
* **Good**, because one configuration file eliminates maintenance overhead of multiple configs
* **Good**, because all database types share consistent patterns and naming conventions
* **Good**, because generated code treated as build artifacts follows industry best practices
* **Good**, because adding new modules only requires updating single schema list
* **Neutral**, because all modules regenerate when any SQL changes (acceptable with modern build speeds)
* **Neutral**, because configuration file grows with each new module (manageable complexity)
* **Bad**, because single package namespace could become large as system scales
* **Bad**, because no module-level isolation of generated database types

### Confirmation

Implementation success will be validated through:
- **Cross-Module Queries**: Successful implementation of queries spanning user and auth domains
- **CI/CD Simplicity**: Single command generation with fast build times
- **Developer Experience**: Team feedback on ease of database access development
- **Configuration Maintenance**: Adding new modules requires only schema list updates
- **Generated Code Quality**: All database types follow consistent patterns and conventions

## Pros and Cons of the Options

### Centralized SQLC Configuration (Chosen Option)

Single root configuration with shared generated code.

* **Good**, because enables natural cross-module queries without schema import complexity
* **Good**, because single `sqlc generate` command handles all code generation requirements
* **Good**, because one configuration file minimizes maintenance as modules are added
* **Good**, because consistent import patterns across all database operations (`db.Queries`, `db.CreateUserParams`)
* **Good**, because generated code follows industry practice of not being tracked in version control
* **Good**, because all database types share naming conventions and patterns
* **Neutral**, because adding modules requires updating single schema list (predictable maintenance)
* **Neutral**, because single package could grow large but mitigated by clear naming conventions
* **Bad**, because all modules regenerate on any SQL change (minimal impact with fast builds)
* **Bad**, because no module-level isolation if different teams own different database domains

### Modular SQLC Configuration

Individual module configurations with domain-specific generated code.

* **Good**, because each module maintains independent configuration and generated code
* **Good**, because module changes only regenerate code for that specific module  
* **Good**, because clear ownership boundaries for different teams working on different domains
* **Good**, because smaller generated packages per module with focused imports
* **Neutral**, because familiar pattern that mirrors application module structure
* **Bad**, because cross-module queries require complex schema dependency management
* **Bad**, because adding schema references requires updating multiple configuration files
* **Bad**, because CI/CD requires complex directory traversal: `find infra -name "sqlc.yaml" -execdir sqlc generate \;`
* **Bad**, because schema imports create tight coupling between supposedly independent modules
* **Bad**, because configuration maintenance scales poorly as module count increases

### Hybrid Approach  

Centralized schema definitions with modular generation targets.

* **Good**, because provides single schema source of truth for all modules
* **Good**, because each module can generate only types it needs
* **Good**, because enables cross-module queries through shared schema access
* **Neutral**, because balances centralized schema benefits with modular generation
* **Bad**, because requires complex configuration mapping schema to multiple targets
* **Bad**, because still requires multiple configuration files and CI generation commands
* **Bad**, because adds complexity without clear benefits over full centralization
* **Bad**, because developers must understand both centralized and modular patterns

### External Schema Management

Separate schema repository with multiple generation targets.

* **Good**, because schema versioning independent of application code
* **Good**, because multiple applications could share same database schema definitions
* **Good**, because enables database-first development approach
* **Neutral**, because clear separation of concerns between schema and application logic
* **Bad**, because adds significant complexity for single-application use case  
* **Bad**, because requires separate repository maintenance and synchronization
* **Bad**, because schema changes require coordination across multiple repositories
* **Bad**, because over-engineered solution for monolithic backend architecture

## More Information

### Architecture Integration Strategy

Our implementation maintains hexagonal architecture principles through clear layer responsibilities:

**Database Access Layer (Generated)**: SQLC generates type-safe database interfaces (`Queries`, parameter structs, model types) that handle all SQL execution and row mapping.

**Repository Adapters**: Thin adapter implementations that use generated database code to fulfill domain repository interfaces. Adapters handle connection management and translate between database types and domain entities.

**Domain Layer**: Continues to define repository interfaces and domain entities without any knowledge of SQLC or database implementation details.

This approach ensures single source of truth for database operations while maintaining clean architecture boundaries.

### Implementation Details

**Root Configuration** (`sqlc.yaml`):
```yaml
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
```

**Usage Pattern**:
```go
import db "class-backend/generated/sqlc"

type PostgresUserRepository struct {
    queries *db.Queries
}

func (r *PostgresUserRepository) Create(user *entities.User) error {
    return r.queries.CreateUser(ctx, db.CreateUserParams{
        Name:  user.Name,
        Email: user.Email,
    })
}
```

### Risk Assessment and Mitigation

**Single Package Growth Risk**: Generated package could become large
- *Mitigation*: Clear naming conventions and Go's compile-time dead code elimination minimize impact

**Configuration Maintenance Risk**: Schema list requires updates for new modules  
- *Mitigation*: Predictable, single-file maintenance with clear documentation

**Cross-Team Coordination Risk**: Multiple teams working on same generated package
- *Mitigation*: Generated code is deterministic and merge conflicts are rare; teams own their SQL inputs

### Generated Code Management

Following industry best practices, all SQLC generated code is treated as build artifacts:
- Excluded from version control via `.gitignore`
- Generated fresh in CI/CD pipelines
- Developers run `sqlc generate` locally during development
- No manual editing of generated files