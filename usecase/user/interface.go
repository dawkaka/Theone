package user

import (
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
)

//Reader interface
type Reader interface {
	Get(userName string) (entity.User, error)
	Search(query string) ([]presentation.UserPreview, error)
	List(users []entity.ID) ([]presentation.UserPreview, error)
	ListFollowers(flws []entity.ID) ([]entity.Follower, error)
	ConfirmCouple(userID, partnerID entity.ID) bool
	Following(userName string, skip int) ([]entity.ID, error)
	Login(param string) (entity.User, error)
	CheckSignup(userName, email string) (entity.User, error)
	Startup(userID entity.ID) (presentation.StartupInfo, error)
}

//Writer user writer
type Writer interface {
	Create(e *entity.User) (entity.ID, error)
	Update(userID entity.ID, update entity.UpdateUser) error
	Delete(id entity.ID) error
	SendRequest(from, to entity.ID) error
	RecieveRequest(from, to entity.ID) error
	Follow(coupleId, userID entity.ID) error
	Unfollow(coupleId, userId entity.ID) error
	Notify(userToNotify string, notification any) error
	NotifyUsers(users []string, notif any) error
	NotifyCouple(c [2]entity.ID, notif any) error
	NewCouple(c [2]entity.ID, coupleID entity.ID) error
	UpdateProfilePic(fileName string, userID entity.ID) error
	UpdateShowPicture(userID entity.ID, index int, fileName string) error
	ChangeRequestStatus(userId entity.ID, status string) error
	ChangeName(userID entity.ID, userName string) error
	ChangeSettings(userID entity.ID, setting, value string) error
	NullifyRequest([2]entity.ID) error
	Notifications(userName string, page int) ([]entity.Notification, error)
	BreakedUp(couple [2]entity.ID) error
}

//Repository interface
type Repository interface {
	Reader
	Writer
}

//UseCase interface
type UseCase interface {
	GetUser(userName string) (entity.User, error)
	SearchUsers(query string) ([]presentation.UserPreview, error)
	ListUsers([]entity.ID) ([]presentation.UserPreview, error)
	ListFollowers(flws []entity.ID) ([]entity.Follower, error)
	UserFollowing(userName string, skip int) ([]entity.ID, error)
	CreateUser(email, password, firstName, lastName, userName string, dateOfBirth time.Time, lang string) (entity.ID, error)
	SendRequest(from, to entity.ID) error
	RecieveRequest(from, to entity.ID) error
	UpdateUser(userID entity.ID, update entity.UpdateUser) error
	DeleteUser(id entity.ID) error
	ConfirmCouple(userID, partnerID string) (bool, error)
	Follow(coupleID, userID entity.ID) error
	Unfollow(coupleID, userID entity.ID) error
	NotifyUser(userToNotify string, notification any) error
	NotifyCouple(c [2]entity.ID, notif any) error
	NotifyMultipleUsers(users []string, notif any) error
	NewCouple(c [2]entity.ID, coupleID entity.ID) error
	UpdateUserProfilePic(fileName string, userID entity.ID) error
	UpdateShowPicture(userID entity.ID, index int, fileName string) error
	ChangeUserRequestStatus(userID entity.ID, status string) error
	ChangeUserName(userID entity.ID, userName string) error
	ChangeSettings(userID entity.ID, setting, value string) error
	Login(param string) (entity.User, error)
	CheckSignup(userName, email string) (entity.User, error)
	NullifyRequest([2]entity.ID) error
	GetNotifications(userName string, page int) ([]entity.Notification, error)
	BreakedUp(couple [2]entity.ID) error
	StartupInfo(userID entity.ID) (presentation.StartupInfo, error)
}
