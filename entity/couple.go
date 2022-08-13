package entity

import (
	"strings"
	"time"

	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/validator"
)

//Couple data

type Couple struct {
	ID             ID        `json:"id" bson:"_id"`
	Initiated      ID        `json:"iniated"`
	Accepted       ID        `json:"accepted"`
	AcceptedAt     time.Time `json:"accepted_at" bson:"accepted_at"`
	CoupleName     string    `json:"couple_name" bson:"couple_name"`
	DateCommenced  time.Time `json:"date_commenced" bson:"date_commenced"`
	Married        bool      `json:"married"`
	Verified       bool      `json:"verified"`
	ProfilePicture string    `json:"profile_picture" bson:"profile_picture"`
	CoverPicture   string    `json:"cover_picture" bson:"cover_picture"`
	Bio            string    `json:"bio"`
	Followers      []ID      `json:"followers" bson:"followers"`
	Separated      bool      `json:"separated"`
	FollowersCount uint64    `json:"followers_count" bson:"followers_count"`
	PostCount      uint64    `json:"post_count" bson:"post_count"`
	Status         string    `json:"status"`
}

type UpdateCouple struct {
	Bio           string    `json:"bio"`
	Website       string    `json:"website"`
	DateCommenced time.Time `json:"date_commenced"`
	Status        string    `json:"status"`
	UpdatedAt     time.Time `bson:"updated_at"`
	Lang          string    `json:"lang"`
}

func (u UpdateCouple) Validate() []string {
	errs := []string{}
	if !validator.IsBio(u.Bio) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidBio"))
	}
	if !validator.IsWebsite(u.Website) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidWebsite"))
	}
	if !validator.IsValidPastDate(u.DateCommenced) {
		errs = append(errs, inter.Localize(u.Lang, "InvalidCommencedDate"))
	}
	if len(u.Status) < 4 || len(u.Status) > 30 {
		errs = append(errs, inter.Localize(u.Lang, "InvalidStatus"))
	}
	return errs
}

func (u *UpdateCouple) Sanitize() {
	u.Bio = strings.TrimSpace(u.Bio)
	u.Website = strings.TrimSpace(u.Website)
	u.Status = strings.TrimSpace(u.Status)
}
