# Demo Implementation Plan

## Overview
Create a focused POC with essential endpoints to showcase:
- **Authentication**: Login/Signup workflow
- **Authorization**: Role-based access control
- **Architecture**: Clean Architecture + DDD + Hexagonal Architecture
- **Multi-tenancy**: Tenant isolation

## Implementation Steps

### 1. Login Use Case & Endpoint
**Files to create**:
- `core/app/auth/application/use-cases/login-use-case/login-command.go`
- `core/app/auth/application/use-cases/login-use-case/login-use-case.go`
- `class/auth/handlers/login-handler.go`
- Add Login RPC to `proto/auth/v1/auth.proto`

**Features**:
- Email/password validation
- JWT token generation (compatible with NextAuth)
- Password hashing verification
- Proper error handling

### 2. ListUsersInTenant (Admin Only)
**Files to create**:
- `core/app/user/application/use-cases/list-users-use-case/`
- `class/user/handlers/user-handler.go`
- Add User service to proto definitions

**Authorization**: Admin role only
**Features**: 
- List all users in requesting user's tenant
- Pagination support
- Tenant isolation enforcement

### 3. AssignRoleInTenant (Admin Only)
**Files to create**:
- `core/app/auth/application/use-cases/assign-role-use-case/`
- Add role assignment endpoint to auth service

**Authorization**: Admin role only
**Features**:
- Assign roles (admin/instructor/student) to users
- Validate role exists and user exists in tenant
- Update Casbin role assignments

### 4. ListStudents (Instructor Only)
**Files to create**:
- `core/app/user/application/use-cases/list-students-use-case/`
- Add to user service

**Authorization**: Instructor role only
**Features**:
- List only users with student role in tenant
- Demonstrate role-filtered queries

### 5. Authorization Middleware Updates
**Updates needed**:
- Add endpoint mappings for new endpoints
- Update JWT token extraction (replace header-based auth)
- Add role validation helpers

### 6. Testing Suite
**Test files to create**:
- Integration tests for each endpoint
- Authorization test scenarios
- JWT token generation/validation tests
- Multi-tenant isolation tests

### 7. Documentation & Demo Scripts
**Files to create**:
- API documentation with examples
- Demo data seeder scripts
- Postman collection for manual testing
- README updates with demo instructions

## Expected Demo Flow
1. **Signup**: Create admin and instructor users
2. **Login**: Authenticate and get JWT tokens
3. **Admin operations**: List users, assign roles
4. **Instructor operations**: List students
5. **Authorization failures**: Show denied access for unauthorized operations

This focused scope provides a complete demonstration of all architectural components while remaining manageable for a POC iteration.