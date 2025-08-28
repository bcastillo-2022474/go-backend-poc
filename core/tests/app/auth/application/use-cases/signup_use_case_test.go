package use_cases

import (
	"github.com/nahualventure/class-backend/core/app/auth/application/use-cases/signup-use-case"
	errors2 "github.com/nahualventure/class-backend/core/app/shared/errors"
	"github.com/nahualventure/class-backend/core/app/user/domain/entities"
	userErrors "github.com/nahualventure/class-backend/core/app/user/domain/errors"
	"github.com/nahualventure/class-backend/core/tests/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUserUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockUserRepository{}
	useCase := signup_use_case.NewCreateUserUseCase(mockRepo)

	command, err := signup_use_case.NewCreateUserCommand("John Doe", "john@example.com", "password123")
	assert.NoError(t, err)

	expectedUser, err := entities.NewUser(
		uuid.New().String(),
		"John Doe",
		"john@example.com",
		time.Now(),
		time.Now(),
	)
	assert.NoError(t, err)

	// Mock expectations
	mockRepo.On("ExistsByEmail", "john@example.com").Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*entities.User"), "password123").Return(expectedUser, nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.NotEmpty(t, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestCreateUserUseCase_Execute_InvalidCommand(t *testing.T) {
	// Arrange
	command, err := signup_use_case.NewCreateUserCommand("John Doe", "invalid-email", "password123")

	// Test invalid email
	assert.Error(t, err)
	assert.Nil(t, command)

	// Test short password
	command, err = signup_use_case.NewCreateUserCommand("John Doe", "john@example.com", "short")
	assert.Error(t, err)
	assert.Nil(t, command)

	// Test empty name
	command, err = signup_use_case.NewCreateUserCommand("", "john@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, command)
}

func TestCreateUserUseCase_Execute_UserAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockUserRepository{}
	useCase := signup_use_case.NewCreateUserUseCase(mockRepo)

	command, err := signup_use_case.NewCreateUserCommand("John Doe", "john@example.com", "password123")
	assert.NoError(t, err)

	// Mock expectations - user already exists
	mockRepo.On("ExistsByEmail", "john@example.com").Return(true, nil)

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr errors2.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.True(t, appErr.IsDomainError())
	assert.Equal(t, string(userErrors.EmailAlreadyExistsError), appErr.GetCode())
	mockRepo.AssertExpectations(t)
}

func TestCreateUserUseCase_Execute_RepositoryExistsByEmailError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockUserRepository{}
	useCase := signup_use_case.NewCreateUserUseCase(mockRepo)

	command, err := signup_use_case.NewCreateUserCommand("John Doe", "john@example.com", "password123")
	assert.NoError(t, err)

	mockRepo.On("ExistsByEmail", "john@example.com").Return(false, errors2.NewInfrastructureError("database read failed", nil))

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr errors2.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.False(t, appErr.IsDomainError())
	assert.Equal(t, string(errors2.InternalError), appErr.GetCode())
	mockRepo.AssertExpectations(t)
}

func TestCreateUserUseCase_Execute_RepositoryCreateError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockUserRepository{}
	useCase := signup_use_case.NewCreateUserUseCase(mockRepo)

	command, err := signup_use_case.NewCreateUserCommand("John Doe", "john@example.com", "password123")
	assert.NoError(t, err)

	// Mock expectations
	mockRepo.On("ExistsByEmail", "john@example.com").Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*entities.User"), "password123").Return(nil, errors2.NewInfrastructureError("database write failed", nil))

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	var appErr errors2.ApplicationError
	assert.ErrorAs(t, err, &appErr)
	assert.False(t, appErr.IsDomainError())
	assert.Equal(t, string(errors2.InternalError), appErr.GetCode())
	mockRepo.AssertExpectations(t)
}
