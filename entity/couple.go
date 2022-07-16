package entity

import "time"

//Couple data

type Couple struct {
	ID             ID        `json:"id"`
	Initiated      ID        `json:"iniated"`
	Accepted       ID        `json:"accepted"`
	AcceptedAt     time.Time `json:"accepted_at" bson:"accepted_at"`
	CoupleName     string    `json:"couple_name" bson:"couple_name"`
	Married        bool      `json:"married"`
	Verified       bool      `json:"verified"`
	ProfilePicture string    `json:"profile_picture" bson:"profile_picture"`
	CoverPicture   string    `json:"cover_picture" bson:"cover_picture"`
	Bio            string    `json:"bio"`
	Followers      []ID      `json:"followers"`
	FollowersCount uint64    `json:"followers_count" bson:"followers_count"`
	PostCount      uint64    `json:"post_count" bson:"post_count"`
	Status         string    `json:"status"`
}
