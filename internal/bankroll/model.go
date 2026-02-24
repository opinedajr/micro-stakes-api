package bankroll

import (
	"time"

	"gorm.io/gorm"
)

type Currency string

const (
	CurrencyBRL Currency = "BRL"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyBTC Currency = "BTC"
)

type Bankroll struct {
	ID                   uint           `gorm:"primaryKey;autoIncrement"`
	UserID               uint           `gorm:"not null;index"`
	Name                 string         `gorm:"type:varchar(100);not null"`
	Currency             Currency       `gorm:"type:varchar(4);not null"`
	InitialBalance       float64        `gorm:"type:decimal(19,4);not null"`
	CurrentBalance       float64        `gorm:"type:decimal(19,4);not null"`
	StartDate            time.Time      `gorm:"type:date;not null"`
	CommissionPercentage float64        `gorm:"type:decimal(5,2);not null"`
	CreatedAt            time.Time      `gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`
}

func (Bankroll) TableName() string {
	return "bankrolls"
}
