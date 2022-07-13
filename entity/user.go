package entity

import (
	"errors"
	"time"

	"github.com/dawkaka/theone/pkg/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

//User data
type User struct {
	ID                   ID        `json:"id" bson:"_id"`
	Email                string    `json:"email"`
	UserName             string    `json:"user_name" bson:"user_name"`
	FirstName            string    `json:"first_name" bson:"first_name"`
	LastName             string    `json:"last_name" bson:"last_name"`
	Password             string    `json:"password"`
	DateOfBirth          time.Time `json:"date_of_birth" bson:"date_of_birth"`
	Bio                  string    `json:"bio"`
	HasPartner           bool      `json:"has_partner" bson:"has_partner"`
	HasPendingRequest    bool      `json:"has_pending_request" bson:"has_pending_request"`
	CreatedAt            time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" bson:"updated_at"`
	ProfilePicture       string    `json:"profile_picture" bson:"profile_picture"`
	CoverPicture         string    `json:"cover_picture" bson:"cover_picture"`
	ShowPictures         []string  `json:"show_pictures" bson:"show_pictures"`
	Likes                []string  `json:"likes"`
	EmailVerified        bool      `json:"email_verified" bson:"email_verified"`
	PartnerID            ID        `json:"partner_id" bson:"partner_id"`
	Following            []ID      `json:"following"`
	FollowingCount       uint64    `json:"following_count" bson:"following_count"`
	Notifications        []any     `json:"notifications"`
	LastVisited          time.Time `json:"last_visited"`
	LoginIPs             []string  `json:"login_ips" bson:"loging_ips"`
	ContentPriorityQueue []ID      `json:"content_priority_queue" bson:"content_priority_queue"`
}

type Follower struct {
	FirstName      string `json:"first_name" bson:"first_name"`
	LastName       string `json:"last_name" bson:"last_name"`
	UserName       string `json:"user_name" bson:"user_name"`
	HasPartner     bool   `json:"has_partner" bson:"has_partner"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
}

type Following struct {
	CoupleName     string `json:"couple_name" bson:"couple_name"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
	Married        string `json:"married"`
	Verified       string `json:"verified"`
}
type UserSession struct {
	ID                ID        `json:"id" bson:"_id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	HasPartner        bool      `json:"has_partner"`
	PartnerID         ID        `json:"partner_id"`
	HasPendingRequest bool      `json:"has_pending_request"`
	FirstName         string    `json:"first_name" bson:"first_name"`
	LastName          string    `json:"last_name" bson:"last_name"`
	DateOfBirth       time.Time `json:"date_of_birth" bson:"date_of_birth"`
}

type Signup struct {
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name" bson:"first_name"`
	LastName    string    `json:"last_name" bson:"last_name"`
	UserName    string    `json:"user_name" bson:"user_name"`
	Password    string    `json:"password"`
	DateOfBirth time.Time `json:"date_of_birth" bson:"date_of_birth"`
}

func (s *Signup) Validate() []error {
	errs := []error{}
	if !validator.IsEmail(s.Email) {
		errs = append(errs, errors.New("wrong email format"))
	}
	if !validator.IsRealName(s.FirstName) {
		errs = append(errs, errors.New("wrong first name format"))
	}
	if !validator.IsRealName(s.FirstName) {
		errs = append(errs, errors.New("wrong last name format"))
	}
	if !validator.IsPassword(s.Password) {
		errs = append(errs, errors.New("wrong password format"))
	}
	return errs
}

type Login struct {
	Email    string `json:"email"`
	UserName string `json:"user_name" bson:"user_name"`
	Password string `json:"password"`
}

type NotifyRequest struct {
	UserName string `json:"user_name"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

func NewUser(email, password, firstName, lastName, userName string, dateOfBirth time.Time) *User {
	return &User{
		Email:                email,
		UserName:             userName,
		FirstName:            firstName,
		LastName:             lastName,
		Password:             password,
		DateOfBirth:          dateOfBirth,
		Bio:                  "-",
		HasPartner:           false,
		HasPendingRequest:    false,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		ProfilePicture:       "defaultProfile.jpg",
		CoverPicture:         "defaultCover.jpg",
		ShowPictures:         []string{"defaultPic.jpg", "defaultPic.jpg", "defaultPic.jpg", "defaultPic.jpg", "defaultPic.jpg", "defaultPic.jpg"},
		Likes:                []string{},
		EmailVerified:        false,
		PartnerID:            [12]byte{},
		Following:            []primitive.ObjectID{},
		FollowingCount:       0,
		Notifications:        []any{},
		LastVisited:          time.Now(),
		LoginIPs:             []string{},
		ContentPriorityQueue: []primitive.ObjectID{},
	}
}
