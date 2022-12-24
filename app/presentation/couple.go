package presentation

import (
	"time"

	"github.com/dawkaka/theone/entity"
)

type CoupleProfile struct {
	AcceptedAt     time.Time `json:"accepted_at"`
	CoupleName     string    `json:"couple_name"`
	ProfilePicture string    `json:"profile_picture"`
	CoverPicture   string    `json:"cover_picture"`
	Bio            string    `json:"bio"`
	FollowersCount uint64    `json:"followers_count"`
	PostCount      uint64    `json:"post_count"`
	Married        bool      `json:"married"`
	UserNames      []string  `json:"user_names"`
	Website        string    `json:"website"`
	DateCommenced  time.Time `json:"date_commenced,omitempty" bson:"date_commenced,omitempty"`
	Verified       bool      `json:"verified"`
	IsThisCouple   bool      `json:"is_this_couple"`
	IsFollowing    bool      `json:"is_following"`
}

type CouplePreview struct {
	ID             entity.ID `json:"id" bson:"_id"`
	CoupleName     string    `json:"couple_name" bson:"couple_name"`
	ProfilePicture string    `json:"profile_picture" bson:"profile_picture"`
	Married        bool      `json:"married"`
	Verified       bool      `json:"verified"`
	IsFollowing    bool      `json:"is_following" bson:"is_following"`
}
