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
	FollowersCount uint64    `json:"followers"`
	PostCount      uint64    `json:"post_count"`
	Status         string    `json:"status"`
}
