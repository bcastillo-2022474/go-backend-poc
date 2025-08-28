package ports

import (
	"github.com/nahualventure/class-backend/core/app/user/domain/entities"
)

type UserRepository interface {
	Create(user *entities.User, password string) (*entities.User, error)
	ExistsByEmail(email string) (bool, error)
	FindByEmail(email string) (*entities.User, error)
}
