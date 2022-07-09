package presentation

type UserProfile struct {
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	UserName       string   `json:"user_name"`
	ProfilePicture string   `json:"profile_picture"`
	CoverPicture   string   `json:"cover_picture"`
	Bio            string   `json:"bio"`
	ShowPictures   []string `json:"show_pictures"`
}
