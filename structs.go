package main

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TinyUrl struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ShortCode string `gorm:"size:64;primaryKey"`
	TrueUrl   string
}

type TinyVisit struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	ShortCode string  `gorm:"size:64"`
	TinyUrl   TinyUrl `gorm:"foreignKey:ShortCode;references:ShortCode"`

	IPAddress string
	UserAgent string
	Referer   string
}

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	IsAdmin  bool   `gorm:"not null;default:false"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(u.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Password") && u.Password != "" {
		if !isAlreadyHashed(u.Password) {
			hashedPassword, err := bcrypt.GenerateFromPassword(
				[]byte(u.Password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				return err
			}
			u.Password = string(hashedPassword)
		}
	}
	return nil
}

func isAlreadyHashed(password string) bool {
	return len(password) == 60 && password[:4] == "$2a$" ||
		password[:4] == "$2b$" || password[:4] == "$2y$"
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

type UserAudit struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID uint `gorm:"not null"`
	User   User `gorm:"foreignKey:UserID"`

	Action    string
	IPAddress string
	Details   string
}
