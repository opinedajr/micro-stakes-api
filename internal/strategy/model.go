package strategy

import (
	"time"

	"gorm.io/gorm"
)

type StrategyType string

const (
	StrategyTypeBack StrategyType = "Back"
	StrategyTypeLay  StrategyType = "Lay"
)

type Strategy struct {
	ID           uint           `gorm:"primaryKey;autoIncrement"`
	UserID       uint           `gorm:"not null;index"`
	Name         string         `gorm:"type:varchar(100);not null"`
	Description  *string        `gorm:"type:text"`
	DefaultStake float64        `gorm:"type:numeric(15,2);not null"`
	Type         StrategyType   `gorm:"type:varchar(4);not null"`
	Active       bool           `gorm:"not null;default:true"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (Strategy) TableName() string {
	return "strategies"
}
