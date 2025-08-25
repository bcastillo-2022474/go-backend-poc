package adapters

import (
	userdb "class-backend/class/user/generated/sqlc"
	appErrors "class-backend/core/app/shared/errors"
	"class-backend/core/app/user/domain/entities"
	"class-backend/core/app/user/domain/ports"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PostgresUserRepository struct {
	db      *pgxpool.Pool
	queries *userdb.Queries
}

func NewPostgresUserRepository(db *pgxpool.Pool) ports.UserRepository {
	return &PostgresUserRepository{
		db:      db,
		queries: userdb.New(db),
	}
}

func (p PostgresUserRepository) Create(user *entities.User, password string) (*entities.User, error) {
	ctx := context.Background()

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(user.ID); err != nil {
		return nil, appErrors.PropagateError(err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErrors.PropagateError(err)
	}

	dbUser, err := p.queries.CreateUser(ctx, userdb.CreateUserParams{
		ID:           pgUUID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: string(hashedPassword),
	})

	if err != nil {
		return nil, appErrors.PropagateError(err)
	}

	return entities.NewUser(
		dbUser.ID.String(),
		dbUser.Name,
		dbUser.Email,
		dbUser.CreatedAt.Time,
		dbUser.UpdatedAt.Time,
	)
}

func (p PostgresUserRepository) ExistsByEmail(email string) (bool, error) {
	ctx := context.Background()
	exists, err := p.queries.ExistsByEmail(ctx, email)

	if err != nil {
		return false, appErrors.PropagateError(err)
	}

	return exists, nil
}

func (p PostgresUserRepository) FindByEmail(email string) (*entities.User, error) {
	ctx := context.Background()
	dbUser, err := p.queries.FindByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, appErrors.PropagateError(err)
	}

	return entities.NewUser(
		dbUser.ID.String(),
		dbUser.Name,
		dbUser.Email,
		dbUser.CreatedAt.Time,
		dbUser.UpdatedAt.Time,
	)
}
