package domain

type Driver struct {
	ID          int64  `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Licenses    string `json:"licenses"`
	Status      string `json:"status"`
}

type CarStatus string

const (
	Available   CarStatus = "available"
	InUse       CarStatus = "in_use"
	Maintenance CarStatus = "maintenance"
	Inactive    CarStatus = "inactive"
)
