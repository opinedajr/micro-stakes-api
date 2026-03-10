package strategy

import (
	"time"
)

type CreateStrategyInput struct {
	Name         string       `json:"name"          binding:"required,min=1,max=100"`
	Description  *string      `json:"description"`
	DefaultStake float64      `json:"default_stake" binding:"required,gte=0.01"`
	Type         StrategyType `json:"type"          binding:"required,oneof=Back Lay"`
	Active       *bool        `json:"active"        binding:"required"`
}

type UpdateStrategyInput struct {
	Name         string       `json:"name"          binding:"required,min=1,max=100"`
	Description  *string      `json:"description"`
	DefaultStake float64      `json:"default_stake" binding:"required,gte=0.01"`
	Type         StrategyType `json:"type"          binding:"required,oneof=Back Lay"`
	Active       *bool        `json:"active"        binding:"required"`
}

type UpdateStrategyStatusInput struct {
	Active bool `json:"active"`
}

type StrategyOutput struct {
	ID           uint         `json:"id"`
	UserID       uint         `json:"user_id"`
	Name         string       `json:"name"`
	Description  *string      `json:"description"`
	DefaultStake float64      `json:"default_stake"`
	Type         StrategyType `json:"type"`
	Active       bool         `json:"active"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type StrategyListOutput struct {
	Data       []*StrategyOutput `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type ErrorOutput struct {
	Error   string              `json:"error"`
	Code    string              `json:"code"`
	Details map[string][]string `json:"details,omitempty"`
}
