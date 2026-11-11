package domain

type UserLogin struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type User struct {
	UserName    string `json:"user_name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	// Password    string `json:"password"`
	RoleID int32 `json:"role_id"`
}

type UserCreate struct {
	UserName    string `json:"user_name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	RoleID      int32  `json:"role_id"`
}

type UserUpdateRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}

type Role struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AdminResetPasswordRequest struct {
	UserID            string `json:"user_id"`
	TemporaryPassword string `json:"temporary_password"`
}
