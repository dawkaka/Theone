package presentation

type UserProfile struct {
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	UserName       string    `json:"user_name"`
	ProfilePicture string    `json:"profile_picture"`
	Bio            string    `json:"bio"`
	FollowingCount uint64    `json:"following_count"`
	ShowPictures   [6]string `json:"show_pictures"`
}
