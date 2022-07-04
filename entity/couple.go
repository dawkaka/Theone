package entity

import "time"

//Couple data

type Couple struct {
	ID             ID        `json:"id"`
	Iniated        ID        `json:"iniated"`
	Accepted       ID        `json:"accepted"`
	IniatedAt      time.Time `json:"iniated_at"`
	AcceptedAt     time.Time `json:"accepted_at"`
	CoupleName     string    `json:"couple_name"`
	ProfilePicture string    `json:"profile_picture"`
	CoverPicture   string    `json:"cover_picture"`
	Bio            string    `json:"bio"`
	Followers      []ID      `json:"followers"`
	PostCount      uint64    `json:"post_count"`
	Status         string    `json:"status"`
}
