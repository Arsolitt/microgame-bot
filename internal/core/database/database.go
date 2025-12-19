package database

import (
	"fmt"
	"microgame-bot/internal/core"

	gormGameRepository "microgame-bot/internal/repo/game"
	gormSessionRepository "microgame-bot/internal/repo/session"
	gormUserRepository "microgame-bot/internal/repo/user"

	gormLogger "gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func MustInit(cfg core.Config) (*gorm.DB, error) {
	const operationName = "core::database::DBMustInit"
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	}
	var dialector gorm.Dialector
	switch cfg.App.GormDialector {
	case "sqlite":
		dialector = sqlite.Open(cfg.Sqlite.URL)
	case "postgres":
		dialector = postgres.Open(cfg.Postgres.URL)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database in %s: %w", operationName, err)
	}

	err = db.AutoMigrate(&gormUserRepository.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate user table in %s: %w", operationName, err)
	}
	err = db.AutoMigrate(&gormSessionRepository.Session{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate game session table in %s: %w", operationName, err)
	}
	err = db.AutoMigrate(&gormGameRepository.Game{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate game table in %s: %w", operationName, err)
	}
	return db, nil
}
