package main

import (
	"context"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/mdw"
	"os"
	"os/signal"

	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	memoryFSM "minigame-bot/internal/fsm/memory"
	"minigame-bot/internal/handlers"
	memoryLocker "minigame-bot/internal/locker/memory"
	gormTTTRepository "minigame-bot/internal/repo/ttt/gorm"
	gormUserRepository "minigame-bot/internal/repo/user/gorm"

	gormLogger "gorm.io/gorm/logger"
)

func startup() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}

	logger.InitLogger(cfg.Logs)
	slog.Info("Logger initialized successfully")

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return err
	}

	db.AutoMigrate(&gormUserRepository.User{})
	db.AutoMigrate(&gormTTTRepository.TTT{})

	_ = memoryLocker.New()
	_ = memoryFSM.New()

	bot, err := telego.NewBot(string(cfg.Telegram.Token), telego.WithDiscardLogger())
	if err != nil {
		return err
	}

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return err
	}

	bh, err := th.NewBotHandler(bot, updates, th.WithErrorHandler(func(ctx *th.Context, update telego.Update, err error) {
		slog.ErrorContext(ctx, "Handler error occurred", logger.ErrorField, err.Error())
	}))
	if err != nil {
		return err
	}
	defer bh.StopWithContext(ctx)

	// userRepo := memoryUserRepository.New()
	userRepo := gormUserRepository.New(db)
	// gameRepo := memoryTTTRepository.New()
	gameRepo := gormTTTRepository.New(db)

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.UserProvider(memoryLocker.New(), userRepo),
	)

	bh.HandleInlineQuery(handlers.WrapInlineQuery(handlers.GameSelector()), th.AnyInlineQuery())

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTCreate(gameRepo)), th.CallbackDataEqual("ttt::create"))

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTJoin(gameRepo, userRepo)), th.CallbackDataPrefix("ttt::join::"))

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTMove(gameRepo, userRepo)), th.CallbackDataPrefix("ttt::move::"))

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.Empty()), th.CallbackDataEqual("empty"))

	// Start handling updates
	return bh.Start()
}

func main() {
	if err := startup(); err != nil {
		slog.Error("Failed to startup application", logger.ErrorField, err.Error())
		os.Exit(1)
	}
}
