package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const ROBOTS_FILE = "./robots.txt"

type TinyUrl struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ShortCode   string `gorm:"size:64;primaryKey"`
	TrueUrl     string
	AllowRobots bool `gorm:"default:false"`
}

// generate robots.txt after save
func (tu *TinyUrl) AfterSaveCommit(tx *gorm.DB) error {
	API_LOGGER.Info().Msg("Updating robots.txt after save.")
	return generateRobotsTxt()
}

func generateRobotsTxt() error {
	var tinyUrls []TinyUrl
	res := db.Where("allow_robots = ?", true).Find(&tinyUrls)
	if res.Error != nil {
		return res.Error
	}

	var strBuilder strings.Builder
	// disallow by default
	strBuilder.WriteString("User-Agent: *\nDisallow: /\n\n")

	for _, tinyUrl := range tinyUrls {
		strBuilder.WriteString(fmt.Sprintf("Allow: /%s\n", tinyUrl.ShortCode))
	}

	newContent := strBuilder.String()

	newHash := sha256.Sum256([]byte(newContent))

	if existingContent, err := os.ReadFile(ROBOTS_FILE); err == nil {
		existingHash := sha256.Sum256(existingContent)

		if bytes.Equal(newHash[:], existingHash[:]) {
			API_LOGGER.Info().Msg("robots.txt unchanged.")
			return nil
		}

		sizeDiff := len(newContent) - len(existingContent)
		API_LOGGER.Info().
			Int("old_size", len(existingContent)).
			Int("new_size", len(newContent)).
			Int("size_diff", sizeDiff).
			Msg("robots.txt content changed, updating file.")
	} else {
		API_LOGGER.Info().
			Int("new_size", len(newContent)).
			Msg("robots.txt doesn't exist, creating new file.")
	}

	if err := os.WriteFile(ROBOTS_FILE, []byte(newContent), 0644); err != nil {
		API_LOGGER.Error().Err(err).Msg("Failed to write robots.txt")
		return err
	}

	API_LOGGER.Info().Msg("robots.txt updated successfully.")
	return nil
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

func (u *User) AfterSaveCommit(tx *gorm.DB) {
	checkForAdmin()
}

func (u *User) AfterDeleteCommit(tx *gorm.DB) {
	checkForAdmin()
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
