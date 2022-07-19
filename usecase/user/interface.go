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
	Update(userID entity.ID, update entity.UpdateUser) error
	Delete(id entity.ID) error
	Request(from, to entity.ID) error
	Follow(coupleId, userID entity.ID) error
	Unfollow(coupleId, userId entity.ID) error
	Notify(userToNotify string, notification any) error
	NotifyCouple(c [2]entity.ID, notif entity.Notification) error
	NewCouple(c [2]entity.ID, coupleID entity.ID) error
	UpdateProfilePic(fileName string, userID entity.ID) error
	UpdateShowPicture(userID entity.ID, index int, fileName string) error
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
	UpdateUser(userID entity.ID, update entity.UpdateUser) error
	DeleteUser(id entity.ID) error
	ConfirmCouple(userID, partnerID string) (bool, error)
	Follow(coupleID, userID entity.ID) error
	Unfollow(coupleID, userID entity.ID) error
	NotifyUser(userToNotify string, notification any) error
	NotifyCouple(c [2]entity.ID, notif entity.Notification) error
	NewCouple(c [2]entity.ID, coupleID entity.ID) error
	UpdateUserProfilePic(fileName string, userID entity.ID) error
	UpdateShowPicture(userID entity.ID, index int, fileName string) error
}
