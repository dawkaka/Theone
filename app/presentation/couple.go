package presentation

import (
	"time"
)

type CoupleProfile struct {
	AcceptedAt     time.Time `json:"accepted_at"`
	CoupleName     string    `json:"couple_name"`
	ProfilePicture string    `json:"profile_picture"`
	CoverPicture   string    `json:"cover_picture"`
	Bio            string    `json:"bio"`
	FollowersCount uint64    `json:"followers_count"`
	PostCount      uint64    `json:"post_count"`
	Status         string    `json:"status"`
	Married        bool      `json:"married"`
	Verified       bool      `json:"verified"`
}

type CouplePreview struct {
	CoupleName     string `json:"couple_name" bson:"couple_name"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
	Status         string `json:"status"`
}
