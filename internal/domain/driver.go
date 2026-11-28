package domain

type Driver struct {
	ID                int64   `json:"id"`
	FirstName         string  `json:"first_name,omitempty"`
	LastName          string  `json:"last_name,omitempty"`
	PhoneNumber       string  `json:"phone_number,omitempty"`
	Licenses          string  `json:"license_number,omitempty"`
	Status            string  `json:"status"`
	ProfilePictureUrl *string `json:"profile_picture_url,omitempty"`
}

type CarStatus string

const (
	Available   CarStatus = "available"
	InUse       CarStatus = "in_use"
	Maintenance CarStatus = "maintenance"
	Inactive    CarStatus = "inactive"
)

type UpdateDriverRequest struct {
	FirstName     *string `json:"first_name,omitempty"`
	LastName      *string `json:"last_name,omitempty"`
	PhoneNumber   *string `json:"phone_number,omitempty"`
	LicenseNumber *string `json:"license_number,omitempty"`
}

type DriverLedger struct {
	DriverID  int32   `json:"driver_id"`
	EntryType string  `json:"entry_type"`
	Amount    float64 `json:"amount"`
	Notes     string  `json:"notes"`
}
