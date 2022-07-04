package entity

import "time"

type Notification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

//User data
type User struct {
	ID                  ID             `json:"id"`
	Email               string         `json:"email"`
	UserName            string         `json:"user_name"`
	FirstName           string         `json:"first_name"`
	LastName            string         `json:"last_name"`
	Password            string         `json:"password"`
	Bio                 string         `json:"bio"`
	HasPartner          bool           `json:"has_partner"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	ProfilePicture      string         `json:"profile_picture"`
	CoverPicture        string         `json:"cover_picture"`
	ShowPictures        []string       `json:"show_pictures"`
	Likes               []string       `json:"likes"`
	Partner             ID             `json:"partner"`
	Following           []ID           `json:"following"`
	Notifications       []Notification `json:"notifications"`
	LastVisited         time.Time      `json:"last_visited"`
	LastLoginIP         []string       `json:"last_login_ip"`
	ContentPriorityQuee []ID           `json:"content_priority_queue"`
}
