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
	Type    string `json:"type"`
	Message string `json:"message"`
}

type MentionedNotif struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	PostID     string `json:"post_id" bson:"post_id"`
	CoupleName string `json:"couple_name" bson:"couple_name"`
}

//User data
type User struct {
	ID                   ID        `json:"id,omitempty" bson:"_id,,omitempty"`
	Email                string    `json:"email,omitempty" bson:"required,omitempty"`
	UserName             string    `json:"user_name,omitempty" bson:"user_name,omitempty"`
	FirstName            string    `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName             string    `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Password             string    `json:"password,omitempty"`
	DateOfBirth          time.Time `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	CoupleID             ID        `json:"couple_id,omitempty" bson:"couple_id,omitempty"`
	Bio                  string    `json:"bio,omitempty" bson:"bio"`
	Website              string    `json:"website,omitempty" bson:"website,omitempty"`
	OpenToRequests       bool      `json:"open_to_requests,omitempty" bson:"open_to_request,omitempty"`
	HasPartner           bool      `json:"has_partner,omitempty" bson:"has_partner,omitempty"`
	HasPendingRequest    bool      `json:"has_pending_request,omitempty" bson:"has_pending_request,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	ProfilePicture       string    `json:"profile_picture,omitempty" bson:"profile_picture,omitempty"`
	ShowPictures         [6]string `json:"show_pictures,omitempty" bson:"show_pictures,omitempty"`
	Likes                []ID      `json:"likes,omitempty"`
	LikesCount           int64     `json:"likes_count,omitempty" bson:"likes_count,omitempty"`
	EmailVerified        bool      `json:"email_verified,omitempty" bson:"email_verified,omitempty"`
	PartnerID            ID        `json:"partner_id,omitempty" bson:"partner_id,omitempty"`
	Following            []ID      `json:"following,omitempty"`
	FollowingCount       uint64    `json:"following_count,omitempty" bson:"following_count,omitempty"`
	Notifications        []any     `json:"notifications,omitempty"`
	Lang                 string    `json:"lang,omitempty"`
	LastVisited          time.Time `json:"last_visited,omitempty"`
	LoginIPs             []string  `json:"login_ips,omitempty" bson:"loging_ips",omitempty`
	ContentPriorityQueue []ID      `json:"content_priority_queue,omitempty" bson:"content_priority_queue,omitempty"`
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
	CoupleID          ID        `json:"couple_id"`
	HasPendingRequest bool      `json:"has_pending_request"`
	FirstName         string    `json:"first_name" bson:"first_name"`
	LastName          string    `json:"last_name" bson:"last_name"`
	Lang              string    `json:"lang" bson:"lang"`
	DateOfBirth       time.Time `json:"date_of_birth" bson:"date_of_birth"`
	LastVisited       time.Time `json:"last_visited" bson:"last_visited"`
}

type Signup struct {
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name" bson:"first_name"`
	LastName    string    `json:"last_name" bson:"last_name"`
	UserName    string    `json:"user_name" bson:"user_name"`
	Password    string    `json:"password"`
	DateOfBirth time.Time `json:"date_of_birth" bson:"date_of_birth"`
}

func (s Signup) Validate() []error {
	errs := []error{}
	if !validator.IsEmail(s.Email) {
		errs = append(errs, errors.New("WrongEmailFormat"))
	}
	if !validator.IsRealName(s.FirstName) {
		errs = append(errs, errors.New("WrongFirstNameFormat"))
	}
	if !validator.IsRealName(s.FirstName) {
		errs = append(errs, errors.New("WrongLastNameFormat"))
	}
	if !validator.IsPassword(s.Password) {
		errs = append(errs, errors.New("WrongPasswordFormat"))
	}
	if !validator.IsUserName(s.UserName) {
		errs = append(errs, errors.New("WrongUserNameFormat"))
	}
	return errs
}

func (s *Signup) Sanitize() {
	s.FirstName = strings.TrimSpace(s.FirstName)
	s.LastName = strings.TrimSpace(s.LastName)
	s.UserName = strings.TrimSpace(s.UserName)
	s.Password = strings.TrimSpace(s.Password)
	s.Email = strings.TrimSpace(s.Password)
}

type Login struct {
	Email    string `json:"email"`
	UserName string `json:"user_name" bson:"user_name"`
	Password string `json:"password"`
}

type NotifyRequest struct {
	UserName string `json:"user_name" bons:"user_name"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

type UpdateUser struct {
	FirstName string    `json:"first_name" bson:"first_name"`
	LastName  string    `json:"last_name" bson:"last_name"`
	Bio       string    `json:"bio"`
	Pronouns  string    `json:"pronouns"`
	UpdatedAt time.Time `bson:"updated_at"`
	Website   string    `json:"website"`
	Lang      string    `json:"lang"`
}

func (u UpdateUser) Validate() []string {
	errs := []string{}
	if !validator.IsRealName(u.FirstName) || !validator.IsRealName(u.LastName) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidFirstNameOrLastName"))
	}
	if !validator.IsBio(u.Bio) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidBio"))
	}
	if !validator.IsPronouns(u.Pronouns) {
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
		HasPendingRequest:    false,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		ProfilePicture:       "defaultProfile.jpg",
		ShowPictures:         [6]string{"defaultshow.jpg", "defaultshow.jpg", "defaultshow.jpg", "defaultshow.jpg", "defaultshow.jpg", "defaultshow.jpg"},
		Likes:                []primitive.ObjectID{},
		LikesCount:           0,
		EmailVerified:        false,
		Following:            []primitive.ObjectID{},
		FollowingCount:       0,
		Notifications:        []any{},
		LastVisited:          time.Now(),
		LoginIPs:             []string{},
		ContentPriorityQueue: []primitive.ObjectID{},
		Lang:                 lang,
	}
}
