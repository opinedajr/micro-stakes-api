package auth

import (
	"time"

	"gorm.io/gorm"
)

type IdentityAdapter string

const (
	IdentityAdapterKeycloak IdentityAdapter = "keycloak"
)

type User struct {
	ID              uint            `gorm:"primaryKey;autoIncrement"`
	FullName        string          `gorm:"type:varchar(200);not null"`
	Email           string          `gorm:"type:varchar(255);uniqueIndex;not null"`
	IdentityID      string          `gorm:"type:varchar(255);not null"`
	IdentityAdapter IdentityAdapter `gorm:"type:varchar(50);not null"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt  `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
