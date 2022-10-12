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
	Website        string    `json:"website"`
	FollowersCount uint64    `json:"followers_count" bson:"followers_count"`
	PostCount      uint64    `json:"post_count" bson:"post_count"`
	Posts          []string  `json:"posts" bson:"posts"`
	IsFollowing    bool      `json:"is_following" bson:"is_following, omitempty"`
}

type UpdateCouple struct {
	Bio           string    `json:"bio"`
	Website       string    `json:"website"`
	DateCommenced time.Time `json:"date_commenced"`
	UpdatedAt     time.Time
	Lang          string
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
	return errs
}

func (u *UpdateCouple) Sanitize() {
	u.Bio = strings.TrimSpace(u.Bio)
	u.Website = strings.TrimSpace(u.Website)
}

type ReportCouple struct {
	CoupleID  ID        `json:"couple_id" bson:"couple_id"`
	UserID    ID        `json:"user_id" bson:"user_id"`
	Report    []int     `json:"reports" bson:"reports"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Type      string    `json:"type" bson:"type"`
}
