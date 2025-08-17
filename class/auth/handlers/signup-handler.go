package handlers

import (
	"class-backend/core/app/auth/application/use-cases/signup-use-case"
	"context"

	"class-backend/class/shared/utils"
	"class-backend/class/user/adapters"
	authv1 "class-backend/proto/generated/go/auth/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ah *AuthHandler) Signup(ctx context.Context, req *authv1.SignupRequest) (*authv1.SignupResponse, error) {
	userAdapter := adapters.NewPostgresUserRepository(ah.db)
	useCase := signup_use_case.NewCreateUserUseCase(userAdapter)

	// Create command with validation
	cmd, err := signup_use_case.NewCreateUserCommand(req.Name, req.Email, req.Password)
	if err != nil {
		return nil, utils.ApplicationErrorToGrpcStatus(err)
	}

	// Execute use case
	user, err := useCase.Execute(cmd)
	if err != nil {
		return nil, utils.ApplicationErrorToGrpcStatus(err)
	}

	// Convert domain entity to proto response
	protoUser := &authv1.User{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return &authv1.SignupResponse{
		User: protoUser,
	}, nil

}
