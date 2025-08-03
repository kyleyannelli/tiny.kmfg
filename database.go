package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db.Callback().Create().After("gorm:commit_or_rollback_transaction").
		Register("after_create_commit", afterCreateCommitCallback)

	db.Callback().Update().After("gorm:commit_or_rollback_transaction").
		Register("after_update_commit", afterUpdateCommitCallback)

	db.Callback().Delete().After("gorm:commit_or_rollback_transaction").
		Register("after_delete_commit", afterDeleteCommitCallback)

	DB_LOGGER.Info().Msg("Database migrations completed")
}

func afterCreateCommitCallback(db *gorm.DB) {
	if db.Error == nil {
		callMethod(db, "AfterCreateCommit")
		callMethod(db, "AfterSaveCommit")
	}
}

func afterUpdateCommitCallback(db *gorm.DB) {
	if db.Error == nil {
		callMethod(db, "AfterUpdateCommit")
		callMethod(db, "AfterSaveCommit")
	}
}

func afterDeleteCommitCallback(db *gorm.DB) {
	if db.Error == nil {
		callMethod(db, "AfterDeleteCommit")
	}
}

func callMethod(db *gorm.DB, methodName string) {
	if db.Statement.Schema != nil {
		if db.Statement.ReflectValue.CanAddr() {
			if methodValue := db.Statement.ReflectValue.Addr().MethodByName(methodName); methodValue.IsValid() {
				methodValue.Call([]reflect.Value{reflect.ValueOf(db)})
			}
		}
	}
}

func visitUrl(tinyUrl *TinyUrl, c *fiber.Ctx) {
	tinyVisit := &TinyVisit{
		ShortCode: tinyUrl.ShortCode,
		IPAddress: c.IP(),
		Referer:   string(c.Request().Header.Referer()),
		UserAgent: string(c.Request().Header.UserAgent()),
	}
	db.Create(&tinyVisit)
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

	var gormLogger logger.Interface
	if os.Getenv("KMFG_TINY_DB_LOG") == "true" {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	DB_LOGGER.Info().Str("database", absPath).Msg("Opening database")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create or open the sqlite3 database at %s.", absPath))
	}

	DB_LOGGER.Info().Str("database", absPath).Msg("Database opened successfully")

	return db
}
