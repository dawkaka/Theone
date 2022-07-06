package entity

import (
	"fmt"
	"time"
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
	Partner              ID             `json:"partner"`
	Following            []ID           `json:"following"`
	Notifications        []Notification `json:"notifications"`
	LastVisited          time.Time      `json:"last_visited"`
	LoginIPs             []string       `json:"login_ips" bson:"loging_ips"`
	ContentPriorityQueue []ID           `json:"content_priority_queue" bson:"content_priority_queue"`
}

func NewUser(firstName, lastName, email string) *User {
	return &User{
		Email:      email,
		LastName:   lastName,
		FirstName:  firstName,
		UserName:   fmt.Sprintf("%s_%s_%d", firstName, lastName, time.Now()),
		CreatedAt:  time.Now(),
		HasPartner: false,
		Bio:        "-",
	}
}

func (u *User) Validate() error {
	if u.Email == "" || u.LastName == "" || u.FirstName == "" {
		return ErrInvalidEntity
	}
	return nil
}
