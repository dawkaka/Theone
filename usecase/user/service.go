package user

import (
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
)

//Service  interface
type Service struct {
	repo Repository
}

//NewService create new use case
func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

//CreateUser Create an user
func (s *Service) CreateUser(email, password, firstName, lastName, userName string,
	dateOfBirth time.Time, lang string) (entity.ID, error) {
	e := entity.NewUser(email, password, firstName, lastName, userName, dateOfBirth, lang)
	return s.repo.Create(e)
}

//Send
func (s *Service) SendRequest(from, to entity.ID) error {
	return s.repo.SendRequest(from, to)
}

func (s *Service) RecieveRequest(from, to entity.ID) error {
	return s.repo.RecieveRequest(from, to)
}

//GetUser Get an user
func (s *Service) GetUser(userName string) (entity.User, error) {
	return s.repo.Get(userName)
}

//SearchUsers Search users
func (s *Service) SearchUsers(query string) ([]presentation.UserPreview, error) {
	return s.repo.Search(query)
}

//ListUsers List users
func (s *Service) ListUsers(users []entity.ID) ([]presentation.UserPreview, error) {
	return s.repo.List(users)
}

func (s *Service) ListFollowers(flw []entity.ID) ([]entity.Follower, error) {
	return s.repo.ListFollowers(flw)
}

func (s *Service) UserFollowing(userName string, skip int) ([]entity.ID, error) {
	return s.repo.Following(userName, skip)
}

//DeleteUser Delete an user
func (s *Service) DeleteUser(id entity.ID) error {
	return s.repo.Delete(id)
}

//UpdateUser Update an user
func (s *Service) UpdateUser(userID entity.ID, update entity.UpdateUser) error {
	update.UpdatedAt = time.Now()
	return s.repo.Update(userID, update)
}

func (s *Service) Follow(coupleID, userID entity.ID) error {
	return s.repo.Follow(coupleID, userID)
}

func (s *Service) Unfollow(coupleID, userID entity.ID) error {
	return s.repo.Unfollow(coupleID, userID)
}

func (s *Service) ConfirmCouple(userID, partnerID string) (bool, error) {
	user, err := entity.StringToID(userID)
	if err != nil {
		return false, err
	}
	parnter, err := entity.StringToID(partnerID)

	if err != nil {
		return false, err
	}
	return s.repo.ConfirmCouple(user, parnter), nil
}

func (s *Service) NotifyUser(user string, notif any) error {
	return s.repo.Notify(user, notif)
}

func (s *Service) NotifyCouple(couple [2]entity.ID, notif any) error {
	return s.repo.NotifyCouple(couple, notif)
}

func (s *Service) NotifyMultipleUsers(users []string, notif any) error {
	return s.repo.NotifyUsers(users, notif)
}

func (s *Service) NewCouple(couple [2]entity.ID, coupleID entity.ID) error {
	return s.repo.NewCouple(couple, coupleID)
}

func (s *Service) UpdateUserProfilePic(fileName string, userID entity.ID) error {
	return s.repo.UpdateProfilePic(fileName, userID)
}

func (s *Service) UpdateShowPicture(userID entity.ID, index int, fileName string) error {
	return s.repo.UpdateShowPicture(userID, index, fileName)
}

func (s *Service) ChangeUserRequestStatus(userID entity.ID, status string) error {
	return s.repo.ChangeRequestStatus(userID, status)
}

func (s *Service) ChangeUserName(userID entity.ID, userName string) error {
	return s.repo.ChangeName(userID, userName)
}

func (s *Service) ChangeSettings(userID entity.ID, setting, value string) error {
	return s.repo.ChangeSettings(userID, setting, value)
}

func (s *Service) Login(param string) (entity.User, error) {
	return s.repo.Login(param)
}

func (s *Service) CheckSignup(userName, email string) (entity.User, error) {
	return s.repo.CheckSignup(userName, email)
}

func (s *Service) NullifyRequest(userIDs [2]entity.ID) error {
	return s.repo.NullifyRequest(userIDs)
}

func (s *Service) GetNotifications(userName string, skip int) (presentation.Notification, error) {
	return s.repo.Notifications(userName, skip)
}

func (s *Service) BreakedUp(couple [2]entity.ID) error {
	return s.repo.BreakedUp(couple)
}

func (s *Service) StartupInfo(userID entity.ID) (presentation.StartupInfo, error) {
	return s.repo.Startup(userID)
}

func (s *Service) ClearNotifsCount(userID entity.ID) error {
	return s.repo.ClearNotifsCount(userID)
}

func (s *Service) UsageMonitoring(userID entity.ID) error {
	return s.repo.UsageMonitoring(userID)
}
