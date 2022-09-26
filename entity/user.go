package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	PostID     string `json:"post_id,omitempty" bson:"post_id,omitempty"`
	CoupleName string `json:"couple_name,omitempty" bson:"couple_name,omitempty"`
}

//User data
type User struct {
	ID                   ID             `json:"id,omitempty" bson:"_id,omitempty"`
	Email                string         `json:"email,omitempty" bson:"email,omitempty"`
	UserName             string         `json:"user_name,omitempty" bson:"user_name,omitempty"`
	FirstName            string         `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName             string         `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Password             string         `json:"password,omitempty"`
	DateOfBirth          time.Time      `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	CoupleID             ID             `json:"couple_id,omitempty" bson:"couple_id,omitempty"`
	Bio                  string         `json:"bio,omitempty" bson:"bio"`
	Website              string         `json:"website,omitempty" bson:"website,omitempty"`
	OpenToRequests       bool           `json:"open_to_requests,omitempty" bson:"open_to_request,omitempty"`
	HasPartner           bool           `json:"has_partner,omitempty" bson:"has_partner"`
	PendingRequest       int8           `json:"pending_request,omitempty" bson:"pending_request"`
	CreatedAt            time.Time      `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at,omitempty" bson:"updated_at"`
	ProfilePicture       string         `json:"profile_picture,omitempty" bson:"profile_picture"`
	ShowPictures         []string       `json:"show_pictures,omitempty" bson:"show_pictures"`
	Likes                []ID           `json:"likes,omitempty"`
	LikesCount           int64          `json:"likes_count,omitempty" bson:"likes_count"`
	EmailVerified        bool           `json:"email_verified,omitempty" bson:"email_verified"`
	PartnerID            ID             `json:"partner_id,omitempty" bson:"partner_id,omitempty"`
	Following            []ID           `json:"following,omitempty"`
	FollowingCount       uint64         `json:"following_count,omitempty" bson:"following_count"`
	Notifications        []Notification `json:"notifications,omitempty"`
	Language             string         `json:"language,omitempty"`
	LastVisited          time.Time      `json:"last_visited,omitempty"`
	LoginIPs             []string       `json:"login_ips,omitempty" bson:"loging_ips"`
	ContentPriorityQueue []ID           `json:"content_priority_queue,omitempty" bson:"content_priority_queue"`
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
	ID             ID        `json:"id" bson:"_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HasPartner     bool      `json:"has_partner"`
	PartnerID      ID        `json:"partner_id"`
	CoupleID       ID        `json:"couple_id"`
	PendingRequest int8      `json:"pending_request"`
	FirstName      string    `json:"first_name" bson:"first_name"`
	LastName       string    `json:"last_name" bson:"last_name"`
	Lang           string    `json:"lang" bson:"lang"`
	DateOfBirth    time.Time `json:"date_of_birth" bson:"date_of_birth"`
	LastVisited    time.Time `json:"last_visited" bson:"last_visited"`
}

type Signup struct {
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name" bson:"first_name"`
	LastName       string    `json:"last_name" bson:"last_name"`
	UserName       string    `json:"user_name" bson:"user_name"`
	Password       string    `json:"password"`
	RepeatPassword string    `json:"repeat_password"`
	DateOfBirth    time.Time `json:"date_of_birth" bson:"date_of_birth"`
}

func (s Signup) Validate() []error {
	errs := []error{}
	if !validator.IsEmail(s.Email) {
		errs = append(errs, errors.New("WrongEmailFormat"))
	}
	if !validator.IsRealName(s.FirstName) {
		errs = append(errs, errors.New("WrongFirstNameFormat"))
	}
	if !validator.IsRealName(s.LastName) {
		errs = append(errs, errors.New("WrongLastNameFormat"))
	}
	if !validator.IsPassword(s.Password) {
		errs = append(errs, errors.New("WrongPasswordFormat"))
	}
	if s.Password != s.RepeatPassword {
		errs = append(errs, errors.New("PasswordsDontMatch"))
	}
	if !validator.IsUserName(s.UserName) {
		errs = append(errs, errors.New("WrongUserNameFormat"))
	}
	if ok, msg := validator.IsValidDateOfBirth(s.DateOfBirth); !ok {
		errs = append(errs, errors.New(msg))
	}
	return errs
}

func (s *Signup) Sanitize() {
	s.FirstName = strings.TrimSpace(s.FirstName)
	s.LastName = strings.TrimSpace(s.LastName)
	s.UserName = strings.TrimSpace(s.UserName)
	s.Password = strings.TrimSpace(s.Password)
	s.Email = strings.ToLower(strings.TrimSpace(s.Password))
}

type Login struct {
	UserNameOrEmail string `json:"user_name_or_email"`
	Password        string `json:"password"`
}

type NotifyRequest struct {
	UserName string `json:"user_name" bons:"user_name"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

type ChangePassword struct {
	Current string `json:"current"`
	New     string `json:"new"`
	Repeat  string `json:"repeat"`
}

func (c ChangePassword) Validate() string {
	if !validator.IsPassword(c.New) {
		return "WrongPasswordFormat"
	}
	if c.New != c.Repeat {
		return "PasswordsDontMatch"
	}
	return ""
}

type UpdateUser struct {
	FirstName   string    `json:"first_name" bson:"first_name"`
	LastName    string    `json:"last_name" bson:"last_name"`
	Bio         string    `json:"bio"`
	Pronouns    string    `json:"pronouns"`
	UpdatedAt   time.Time `bson:"updated_at"`
	Website     string    `json:"website"`
	Lang        string    `json:"lang"` //for error messages
	DateOfBirth time.Time `json:"date_of_birth" bson:"date_of_birth"`
}

func (u UpdateUser) Validate() []string {
	errs := []string{}
	if !validator.IsRealName(u.FirstName) {
		errs = append(errs, inter.Localize(u.Lang, "WrongFirstNameFormat"))
	}
	if !validator.IsRealName(u.LastName) {
		errs = append(errs, inter.Localize(u.Lang, "WrongLastNameFormat"))
	}
	if u.Bio != "" && !validator.IsBio(u.Bio) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidBio"))
	}
	if u.Pronouns != "" && !validator.IsPronouns(u.Pronouns) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidPronouns"))
	}
	if !validator.IsWebsite(u.Website) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidWebsite"))
	}
	return errs
}

func (u *UpdateUser) Sanitize() {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)
	u.Bio = strings.TrimSpace(u.Bio)
	u.Pronouns = strings.TrimSpace(u.Pronouns)
	u.Website = strings.TrimSpace(u.Website)
}

func NewUser(email, password, firstName, lastName, userName string, dateOfBirth time.Time, lang string) *User {
	return &User{
		ID:                   primitive.NewObjectID(),
		Email:                email,
		UserName:             userName,
		FirstName:            firstName,
		LastName:             lastName,
		Password:             password,
		DateOfBirth:          dateOfBirth,
		Bio:                  "-",
		OpenToRequests:       true,
		HasPartner:           false,
		PendingRequest:       NO_REQUEST,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		ProfilePicture:       "defaultProfile.jpg",
		ShowPictures:         []string{"defaultshow1.jpg", "defaultshow2.jpg", "defaultshow3.jpg", "defaultshow4.jpg", "defaultshow5.jpg", "defaultshow6.jpg"},
		Likes:                []primitive.ObjectID{},
		LikesCount:           0,
		EmailVerified:        false,
		Following:            []primitive.ObjectID{},
		FollowingCount:       0,
		Notifications:        []Notification{},
		LastVisited:          time.Now(),
		LoginIPs:             []string{},
		ContentPriorityQueue: []primitive.ObjectID{},
		Language:             lang,
	}
}
