package domain

import "github.com/google/uuid"

type CreateTripParams struct {
	DriverID   int32     `json:"driver_id"`
	TotalCost  float64   `json:"total_cost"`
	ApprovedBy uuid.UUID `json:"approved_by"`
	ExpenseIDs []int32   `json:"expense_ids"`
}
