package entity

import (
	"errors"
	"time"

	"github.com/dawkaka/theone/pkg/password/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

//User data
type User struct {
	ID                   ID             `json:"id" bson:"_id"`
	Email                string         `json:"email"`
	UserName             string         `json:"user_name" bson:"user_name"`
	FirstName            string         `json:"first_name" bson:"first_name"`
	LastName             string         `json:"last_name" bson:"last_name"`
	Password             string         `json:"password"`
	DateOfBirth          time.Time      `json:"date_of_birth" bson:"date_of_birth"`
	Bio                  string         `json:"bio"`
	HasPartner           bool           `json:"has_partner" bson:"has_partner"`
	CreatedAt            time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" bson:"updated_at"`
	ProfilePicture       string         `json:"profile_picture" bson:"profile_picture"`
	CoverPicture         string         `json:"cover_picture" bson:"cover_picture"`
	ShowPictures         []string       `json:"show_pictures" bson:"show_pictures"`
	Likes                []string       `json:"likes"`
	EmailVerified        bool           `json:"email_verified" bson:"email_verified"`
	Partner              ID             `json:"partner"`
	Following            []ID           `json:"following"`
	Notifications        []Notification `json:"notifications"`
	LastVisited          time.Time      `json:"last_visited"`
	LoginIPs             []string       `json:"login_ips" bson:"loging_ips"`
	ContentPriorityQueue []ID           `json:"content_priority_queue" bson:"content_priority_queue"`
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
		errs = append(errs, errors.New("Wrong password format"))
	}
	return errs
}

type Login struct {
	Email    string `json:"email"`
	UserName string `json:"user_name" bson:"user_name"`
	Password string `json:"password"`
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
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Time{},
		ProfilePicture:       "defaultProfile.jpg",
		CoverPicture:         "defaultCover.jpg",
		ShowPictures:         []string{},
		Likes:                []string{},
		EmailVerified:        false,
		Partner:              [12]byte{},
		Following:            []primitive.ObjectID{},
		Notifications:        []Notification{},
		LastVisited:          time.Time{},
		LoginIPs:             []string{},
		ContentPriorityQueue: []primitive.ObjectID{},
	}
}
