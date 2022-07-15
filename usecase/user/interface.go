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
	Following(userName string, skip int) ([]entity.Following, error)
}

//Writer user writer
type Writer interface {
	Create(e *entity.User) (entity.ID, error)
	Update(e *entity.User) error
	Delete(id entity.ID) error
	Request(from, to entity.ID) error
	Follow(coupleId, userID entity.ID) error
	Notify(userToNotify string, notification any) error
	NotifyCouple(c [2]entity.ID, notif entity.Notification) error
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
	UserFollowing(userName string, skip int) ([]entity.Following, error)
	CreateUser(email, password, firstName, lastName, userName string, dateOfBirth time.Time) (entity.ID, error)
	CreateRequest(from, to entity.ID) error
	UpdateUser(e *entity.User) error
	DeleteUser(id entity.ID) error
	ConfirmCouple(userID, partnerID string) (bool, error)
	Follow(coupleID, userID entity.ID) error
	NotifyUser(userToNotify string, notification any) error
	NotifyCouple(c [2]entity.ID, notif entity.Notification) error
}
