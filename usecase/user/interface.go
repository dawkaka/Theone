package user

import (
	"time"

	"github.com/dawkaka/theone/entity"
)

//Reader interface
type Reader interface {
	Get(userName string) (*entity.User, error)
	Search(query string) ([]*entity.User, error)
	List(users []entity.ID) ([]entity.User, error)
	ConfirmCouple(userID, partnerID entity.ID) bool
}

//Writer user writer
type Writer interface {
	Create(e *entity.User) error
	Update(e *entity.User) error
	Delete(id entity.ID) error
	Request(from, to entity.ID) error
	Notify(userToNotify string, notification any) error
}

//Repository interface
type Repository interface {
	Reader
	Writer
}

//UseCase interface
type UseCase interface {
	GetUser(userName string) (*entity.User, error)
	SearchUsers(query string) ([]*entity.User, error)
	ListUsers([]entity.ID) ([]entity.User, error)
	CreateUser(email, password, firstName, lastName, userName string, dateOfBirth time.Time) error
	CreateRequest(from, to entity.ID) error
	UpdateUser(e *entity.User) error
	DeleteUser(id entity.ID) error
	ConfirmCouple(userID, partnerID string) (bool, error)
	NotifyUser(userToNotify string, notification any) error
}
