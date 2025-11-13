package domain

type UserLogin struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type User struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	// Password string `json:"password"`
	RoleID int32 `json:"role_id"`
}

type Role struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}
