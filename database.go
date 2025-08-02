package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DEFAULT_DB = "kmfg.tiny.db"
)

var db *gorm.DB

func setupDb() {
	db = createOrOpenDb()

	db.AutoMigrate(&TinyUrl{})
	db.AutoMigrate(&TinyVisit{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&UserAudit{})

	DB_LOGGER.Info().Msg("Database migrations completed")
}

func createOrOpenDb() *gorm.DB {
	dbPath := os.Getenv("KMFG_TINY_DB")
	if dbPath == "" {
		dbPath = DEFAULT_DB
	}

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		DB_LOGGER.Fatal().Err(err).Str("path", dbPath).Msg("Failed to resolve absolute path")
	}

	DB_LOGGER.Info().Str("database", absPath).Msg("Opening database")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create or open the sqlite3 database at %s.", absPath))
	}

	DB_LOGGER.Info().Str("database", absPath).Msg("Database opened successfully")

	return db
}
