package bankroll

import (
	"time"
)

type CreateBankrollInput struct {
	Name                 string   `json:"name" binding:"required,min=1,max=100"`
	Currency             Currency `json:"currency" binding:"required"`
	InitialBalance       float64  `json:"initial_balance" binding:"required,gte=0"`
	StartDate            string   `json:"start_date" binding:"required"`
	CommissionPercentage float64  `json:"commission_percentage" binding:"required,gte=0,lte=100"`
}

type UpdateBankrollInput struct {
	Name                 string   `json:"name" binding:"required,min=1,max=100"`
	Currency             Currency `json:"currency" binding:"required"`
	StartDate            string   `json:"start_date" binding:"required"`
	CommissionPercentage float64  `json:"commission_percentage" binding:"required,gte=0,lte=100"`
}

type BankrollOutput struct {
	ID                   uint      `json:"id"`
	Name                 string    `json:"name"`
	Currency             Currency  `json:"currency"`
	InitialBalance       float64   `json:"initial_balance"`
	CurrentBalance       float64   `json:"current_balance"`
	StartDate            string    `json:"start_date"`
	CommissionPercentage float64   `json:"commission_percentage"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ErrorOutput struct {
	Error   string              `json:"error"`
	Code    string              `json:"code"`
	Details map[string][]string `json:"details,omitempty"`
}
