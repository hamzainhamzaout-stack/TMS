package domain

import "github.com/google/uuid"

type CreateTripParams struct {
	DriverID int32 `json:"driver_id"`
	// total cost indicates the total amount of money allocated by mahi
	// for this specific trip
	TotalCost  float64   `json:"total_cost"`
	ApprovedBy uuid.UUID `json:"approved_by"`
	ExpenseIDs []int32   `json:"expense_ids"`
	ClientName string    `json:"client_name"`
	TotalQuote float64   `json:"total_quote"`
	PaidAmount float64   `json:"paid_amount"`
}

type TripExpense struct {
	TripID    int     `json:"trip_id"`
	ExpenseID int     `json:"expense_id"`
	Cost      float64 `json:"cost"`
	ImageURL  string  `json:"image_url"`
	Remark    string  `json:"remark"`
}

type Transaction struct {
	ID           int     `json:"id"`
	TripID       int     `json:"trip_id"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	Method       string  `json:"method"`
	ReceiverID   string  `json:"receiver_id"`
	ReceiverName string  `json:"receiver_name"`
	Notes        string  `json:"notes"`
}

type MaintenanceAddition struct {
	CarID       int32   `json:"car_id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
}

type FinancialSummary struct {
	Revenue      float64        `json:"revenue"`
	Expenses     ExpenseSummary `json:"expenses"`
	NetProfit    float64        `json:"net_profit"`
	ProfitMargin float64        `json:"profit_margin"`
}

type ExpenseSummary struct {
	Maintenance   float64 `json:"maintenance"`
	DriverPayroll float64 `json:"driver_payroll"`
	TripOverheads float64 `json:"trip_overheads"`
}

type TripProfitability struct {
	TripID        int32   `json:"trip_id"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	NetProfit     float64 `json:"net_profit"`
}

type DriverPayment struct {
	DriverID     int     `json:"driver_id"`
	Amount       float64 `json:"amount"`
	Method       string  `json:"method"`
	ReceiverName string  `json:"receiver_name"`
	Notes        string  `json:"notes"`
}

type TripOverhead struct {
	TripID       int     `json:"trip_id"`
	Amount       float64 `json:"amount"`
	Method       string  `json:"method"`
	ReceiverName string  `json:"receiver_name"`
	Notes        string  `json:"notes"`
}
