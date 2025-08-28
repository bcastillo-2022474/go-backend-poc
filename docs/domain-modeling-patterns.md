# Domain Modeling Patterns & Design Decisions

## Overview & Context

This document outlines our pragmatic approach to domain modeling in the class-backend-poc, balancing architectural cleanliness with development velocity and team constraints.

**Constraints:**
- Team Size: 2 developers
- Timeline: MVP with tight deadlines
- Domain Complexity: Educational platform with moderate business rules
- Priority: Ship fast while maintaining clean architecture boundaries

## Core Architecture Principles

1. **Transaction Script by Default** - Simple, clear, fast to implement
2. **Rich Domain Entities** - Never anemic, always with behavior and lifecycle control
3. **Explicit Validation** - Enforce validation on every state transition through encapsulation
4. **Interface Segregation** - Focused repository interfaces over monoliths
5. **Pragmatic Value Objects** - Use go tags first, upgrade when business complexity demands it

## Entity Design: Rich Domain Models with Encapsulation

### Private Fields + Public Getters + Explicit State Change Methods

Our entities are **not** data holders (POCOs). They have behavior, control their own lifecycle, and enforce validation on every state change through encapsulation.

```go
type User struct {
    // Private fields - cannot be accessed directly from outside
    id        string    `validate:"required,uuid4"`
    name      string    `validate:"required,min=2,max=100"`
    email     string    `validate:"required,email"`
    isActive  bool
    createdAt time.Time `validate:"required"`
    updatedAt time.Time `validate:"required"`
}

// Constructor for NEW users (business operation)
func NewUser(name, email string) (*User, error) {
    user := &User{
        id:        uuid.NewString(),
        name:      name,
        email:     email,
        isActive:  true,
        createdAt: time.Now().UTC(),
        updatedAt: time.Now().UTC(),
    }
    
    if err := user.Validate(); err != nil {
        return nil, err
    }
    
    return user, nil
}

// Constructor for EXISTING users (from database)
func LoadUser(id, name, email string, isActive bool, createdAt, updatedAt time.Time) (*User, error) {
    user := &User{
        id:        id,
        name:      name,
        email:     email,
        isActive:  isActive,
        createdAt: createdAt,
        updatedAt: updatedAt,
    }
    
    if err := user.Validate(); err != nil {
        return nil, err
    }
    
    return user, nil
}

// Public getters for external access
func (u *User) ID() string { return u.id }
func (u *User) Name() string { return u.name }
func (u *User) Email() string { return u.email }
func (u *User) IsActive() bool { return u.isActive }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Explicit state change methods - validation guaranteed
func (u *User) ChangeName(newName string) error {
    u.name = newName
    u.updatedAt = time.Now().UTC()
    return u.Validate()
}

func (u *User) ChangeEmail(newEmail string) error {
    u.email = newEmail
    u.updatedAt = time.Now().UTC()
    return u.Validate()
}

func (u *User) Activate() error {
    if u.isActive {
        return errors.New("user already active")
    }
    
    u.isActive = true
    u.updatedAt = time.Now().UTC()
    
    return u.Validate()
}

func (u *User) Deactivate() error {
    if !u.isActive {
        return errors.New("user already inactive")
    }
    
    u.isActive = false
    u.updatedAt = time.Now().UTC()
    
    return u.Validate()
}

// Business rules about the entity itself
func (u *User) CanBeDeactivated() bool {
    return u.isActive && time.Since(u.createdAt) > 24*time.Hour
}
```

### Why Private Fields Are Critical

**The Problem with Public Fields:**
```go
// ❌ With public fields, validation can be bypassed
type User struct {
    Name string `validate:"required,min=2"`
}

// Somewhere in domain service or use case
user.Name = "" // ❌ Creates invalid state - no validation run!
```

**The Solution with Private Fields:**
```go
// ✅ With private fields, validation is enforced
type User struct {
    name string `validate:"required,min=2"`
}

func (u *User) ChangeName(newName string) error {
    u.name = newName
    return u.Validate() // ✅ Impossible to bypass validation
}

// user.name = "" // ❌ Compile error - cannot access private field
```

### Constructor Pattern: NewUser vs LoadUser

**Why Two Constructors:**
- `NewUser()`: Business operation with default values and generated IDs
- `LoadUser()`: Reconstruction from stored data with existing values
- Clear separation between creating new vs loading existing

**Usage Examples:**
```go
// In Use Cases (New Users)
func (uc *CreateUserUseCase) Execute(cmd CreateUserCommand) (*User, error) {
    user, err := NewUser(cmd.Name, cmd.Email)
    if err != nil {
        return nil, err
    }
    return uc.userRepo.Save(user)
}

// In Repository Adapters (Existing Users)
func (r *PostgresUserRepository) FindByID(id string) (*User, error) {
    dbUser, err := r.queries.FindUserByID(id)
    if err != nil {
        return nil, err
    }
    
    return LoadUser(
        dbUser.ID,
        dbUser.Name, 
        dbUser.Email,
        dbUser.IsActive,
        dbUser.CreatedAt,
        dbUser.UpdatedAt,
    )
}
```

### Entity Validation Strategy

We use **go-playground/validator tags + explicit validation calls** on every state transition.

```go
var validate = validator.New()

func (u *User) Validate() error {
    return validate.Struct(u)
}

// Template for all state changes
func (u *User) SomeStateChange(...) error {
    // 1. Business rule checks (entity-specific invariants)
    if !u.canMakeThisChange() {
        return errors.New("business rule violation")
    }
    
    // 2. Apply changes to private fields
    u.someField = newValue
    u.updatedAt = time.Now().UTC()
    
    // 3. Always validate - this is why private fields are crucial
    return u.Validate()
}
```

## Business Logic Placement: Transaction Script vs Domain Services

### Default Pattern: Transaction Script

We use **Transaction Script** as our default pattern for simple business operations.

#### When to Use Transaction Script (Default)

```go
func (uc *CreateUserUseCase) Execute(cmd *CreateUserCommand) (*entities.User, error) {
    // Direct business logic in use case
    exists, err := uc.userRepo.ExistsByEmail(cmd.Email)
    if err != nil {
        return nil, errors.PropagateError(err)
    }

    if exists {
        return nil, userErrors.NewEmailAlreadyExistsError(cmd.Email)
    }

    // Entity handles its own creation and validation
    user, err := entities.NewUser(cmd.Name, cmd.Email)
    if err != nil {
        return nil, errors.PropagateError(err)
    }

    return uc.userRepo.Save(user, cmd.Password)
}
```

**Use Transaction Script for:**
- ✅ Simple CRUD operations
- ✅ Business logic used in 1-2 places only
- ✅ Straightforward validation
- ✅ Time-sensitive features

#### When to Upgrade to Domain Services

```go
type UserDomainService struct {
    userReader   ports.UserReader
    tenantReader ports.TenantReader
}

func (s *UserDomainService) ProcessUserRegistration(name, email, tenantID string) (*User, error) {
    // Multiple repository calls
    exists, err := s.userReader.ExistsByEmail(email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("email already taken")
    }
    
    // Complex tenant validation
    tenant, err := s.tenantReader.FindByID(tenantID)
    if err != nil {
        return nil, err
    }
    
    userCount, err := s.userReader.CountByTenant(tenantID)
    if err != nil {
        return nil, err
    }
    
    if userCount >= tenant.UserLimit() {
        return nil, errors.New("tenant user limit exceeded")
    }
    
    // Multiple business rules
    if s.isEmailDomainBanned(email, tenant.AllowedDomains()) {
        return nil, errors.New("email domain not allowed for this tenant")
    }
    
    if s.isRegistrationPeriodClosed(tenantID) {
        return nil, errors.New("registration period closed")
    }
    
    return NewUser(name, email, tenantID)
}
```

**Upgrade to Domain Service when:**
- ❌ Business logic spans multiple entities
- ❌ Logic needs reuse across 3+ use cases
- ❌ Complex validation requiring multiple repository calls
- ❌ Core domain concepts that business experts discuss frequently

### Logic Placement Rules

#### Entity Methods: Single Entity, No External Dependencies
```go
// ✅ Lives in Entity - no external dependencies
func (u *User) Activate() error {
    if u.isActive {
        return errors.New("already active")
    }
    u.isActive = true
    return u.Validate()
}

// ❌ Don't do this - requires external dependency
func (u *User) ChangeEmail(newEmail string, userRepo UserRepository) error {
    // Repository dependency doesn't belong in entity
}
```

#### Use Cases: Application Orchestration
Use cases orchestrate workflow but prefer **Transaction Script** for simple operations.

```go
// ✅ Simple use case - keep logic here (Transaction Script)
func (uc *CreateUserUseCase) Execute(cmd CreateUserCommand) (*User, error) {
    exists, err := uc.userRepo.ExistsByEmail(cmd.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("email already exists")
    }
    
    user, err := NewUser(cmd.Name, cmd.Email)
    if err != nil {
        return nil, err
    }
    
    return uc.userRepo.Save(user)
}

// ✅ Complex use case - delegate to domain service
func (uc *RegisterUserUseCase) Execute(cmd RegisterUserCommand) (*User, error) {
    user, err := uc.userDomainService.ProcessUserRegistration(cmd.Name, cmd.Email, cmd.TenantID)
    if err != nil {
        return nil, err
    }
    
    return uc.userRepo.Save(user)
}
```

#### Responsibility Chain
```
Use Case → Domain Service → Entity Methods → Private Fields
```

- **Use Cases**: Never access entity fields directly
- **Domain Services**: Call entity state change methods, handle business logic requiring external dependencies  
- **Entity Methods**: Control internal state, ensure validation on every change
- **Private Fields**: Protected from external modification

## Repository Design: Interface Segregation

We use **Interface Segregation Principle** instead of monolithic repository interfaces.

```go
// Query operations
type UserReader interface {
    FindByID(ctx context.Context, id string) (*entities.User, error)
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// Command operations
type UserWriter interface {
    Save(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id string) error
}

// Authentication-specific operations
type UserAuthRepository interface {
    CreateWithCredentials(ctx context.Context, user *entities.User, hashedPassword string) error
    ValidateCredentials(ctx context.Context, email string, password string) (*entities.User, error)
}

// Composed interfaces for specific use cases
type UserRepository interface {
    UserReader
    UserWriter
}
```

**Benefits:**
- ✅ Use cases depend only on methods they need
- ✅ Easier to mock specific behaviors in tests
- ✅ Clear separation of concerns
- ✅ Flexible composition

## Validation Layers

We have **three distinct validation layers** with clear responsibilities:

### 1. HTTP/Infrastructure Layer
```go
type CreateUserHTTPRequest struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
    TenantID string `header:"X-Tenant-Id" validate:"required,uuid4"`
}
```
**Purpose**: "Can I extract valid data from this HTTP request?"

### 2. Application Layer (Use Case Input)
```go
type CreateUserCommand struct {
    Name     string `validate:"required,min=2,max=100"`        // Business constraints
    Email    string `validate:"required,email,max=255"`        // Application limits
    Password string `validate:"required,min=8,max=128"`        // Security policy
    TenantID string `validate:"required,uuid4"`                // Business rule
}
```
**Purpose**: "Does this data meet our application's business requirements?"

### 3. Domain Layer
```go
func (u *User) Validate() error {
    return validate.Struct(u) // Entity validation
}

func (s *UserDomainService) CanCreateUser(...) error {
    // Business rules requiring external state
}
```
**Purpose**: "Does this entity/operation violate business rules?"

## Value Objects Strategy

### Current Approach: Go Tags + Validation

We chose **go tags over value objects** for pragmatic reasons:

**Go Tags + Validation:**
- ✅ Faster development (less boilerplate)
- ✅ Familiar to team members
- ✅ Easy to iterate and modify
- ✅ Good enough for MVP validation needs

**Value Objects:**
- ❌ More verbose for simple validation
- ❌ Overkill for basic format/length rules
- ❌ Slower to implement under time constraints

### Migration Strategy

Start with go tags, upgrade to value objects when business rules become complex:

```go
// Current: go tags
type User struct {
    email string `validate:"required,email"`
}

// Future: value object when business rules get complex
type Email string

func NewEmail(email string) (Email, error) {
    // Complex business rules
    if strings.Contains(email, "@tempmail.com") {
        return "", errors.New("temporary emails not allowed")
    }
    return Email(email), nil
}
```

### When to Consider Migration

Move to value objects when you experience:
- ❌ Duplicated validation logic across entities
- ❌ Complex business rules beyond format/length validation
- ❌ Domain experts frequently discuss the concept
- ❌ Invalid states causing production bugs

## Future-Proofing Strategy

Our current approach supports easy migration to value objects:

```go
// Phase 1: Private fields + getters (current)
type User struct {
    email string `validate:"required,email"`
}

func (u *User) Email() string { return u.email }

// Phase 2: Internal value object (future)
type User struct {
    email Email // Changed type, same public API
}

func (u *User) Email() string { return string(u.email) }
```

## Implementation Templates

### 1. Entity State Changes
```go
func (entity *Entity) StateChangeMethod(...) error {
    // 1. Pre-condition checks
    if !entity.canMakeChange() {
        return errors.New("invalid state transition")
    }
    
    // 2. Apply changes
    entity.field = newValue
    entity.updatedAt = time.Now().UTC()
    
    // 3. Always validate
    return entity.Validate()
}
```

### 2. Constructor Pattern
```go
func NewEntity(...) (*Entity, error) {
    entity := &Entity{
        // Initialize fields with business defaults
    }
    
    if err := entity.Validate(); err != nil {
        return nil, err
    }
    
    return entity, nil
}
```

### 3. Testing Approach
```go
func TestUser_Activate_Success(t *testing.T) {
    user := createInactiveUser()
    
    err := user.Activate()
    
    assert.NoError(t, err)
    assert.True(t, user.IsActive())
}

func TestUser_Activate_AlreadyActive(t *testing.T) {
    user := createActiveUser()
    
    err := user.Activate()
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already active")
}
```

## Summary

This approach balances architectural cleanliness with development velocity, ensuring we can ship an MVP quickly while maintaining the ability to evolve our domain model as business complexity grows. The key is starting simple with Transaction Script and Rich Entities, then upgrading to Domain Services and Value Objects only when complexity justifies the additional abstraction.