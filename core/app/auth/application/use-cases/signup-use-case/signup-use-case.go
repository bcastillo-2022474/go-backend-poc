package signup_use_case

import (
	"class-backend/core/app/shared/errors"
	"class-backend/core/app/user/domain/entities"
	userErrors "class-backend/core/app/user/domain/errors"
	"class-backend/core/app/user/domain/ports"
	"time"

	"github.com/google/uuid"
)

type CreateUserUseCase struct {
	userRepo ports.UserRepository
}

func NewCreateUserUseCase(userRepo ports.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
	}
}

func (uc *CreateUserUseCase) Execute(cmd *CreateUserCommand) (*entities.User, error) {
	// Check if email already exists
	exists, err := uc.userRepo.ExistsByEmail(cmd.Email)
	if err != nil {
		return nil, errors.PropagateError(err)
	}

	if exists {
		return nil, userErrors.NewEmailAlreadyExistsError(cmd.Email)
	}

	// Create user entity
	user, err := entities.NewUser(uuid.NewString(), cmd.Name, cmd.Email, time.Now(), time.Now())
	if err != nil {
		return nil, errors.PropagateError(err)
	}

	// Persist user
	createdUser, err := uc.userRepo.Create(user, cmd.Password)
	if err != nil {
		return nil, errors.PropagateError(err)
	}

	return createdUser, nil
}
