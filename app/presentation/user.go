package presentation

import (
	"time"

	"github.com/dawkaka/theone/entity"
)

type UserProfile struct {
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	UserName       string    `json:"user_name"`
	ProfilePicture string    `json:"profile_picture"`
	Bio            string    `json:"bio"`
	FollowingCount uint64    `json:"following_count"`
	ShowPictures   []string  `json:"show_pictures"`
	HasPartener    bool      `json:"has_partner"`
	IsThisUser     bool      `json:"is_this_user"`
	Website        string    `json:"website"`
	DateOfBirth    time.Time `json:"date_of_birth"`
}

type UserPreview struct {
	ID             entity.ID `json:"id" bson:"_id"`
	FirstName      string    `json:"first_name" bson:"first_name"`
	LastName       string    `json:"last_name" bson:"last_name"`
	UserName       string    `json:"user_name" bson:"user_name"`
	HasPartner     bool      `json:"has_partner" bson:"has_partner"`
	ProfilePicture string    `json:"profile_picture" bson:"profile_picture"`
	PendingRequest int8      `json:"pending_request" bson:"pending_request"`
	PartnerID      entity.ID `json:"partner_id" bson:"partner_id"`
	Lang           string    `json:"lang"`
}
