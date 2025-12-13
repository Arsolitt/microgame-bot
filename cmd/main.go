package main

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/mdw"
	"os"
	"os/signal"

	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	memoryFSM "microgame-bot/internal/fsm/memory"
	"microgame-bot/internal/handlers"
	memoryLocker "microgame-bot/internal/locker/memory"
	gormTTTRepository "microgame-bot/internal/repo/ttt/gorm"
	gormUserRepository "microgame-bot/internal/repo/user/gorm"

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

	userLocker := memoryLocker.New[user.ID]()
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
	tttRepo := gormTTTRepository.New(db)

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.UserProvider(userLocker, userRepo),
	)

	bh.HandleInlineQuery(handlers.WrapInlineQuery(handlers.GameSelector()), th.AnyInlineQuery())

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTCreate(tttRepo)), th.CallbackDataEqual("create::ttt"))

	tttG := bh.Group(th.CallbackDataPrefix("g::ttt::"))
	tttG.Use(mdw.GameProvider(memoryLocker.New[ttt.ID](), tttRepo, "ttt"))

	tttG.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTJoin(tttRepo, userRepo)), th.CallbackDataPrefix("g::ttt::join::"))
	tttG.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTMove(tttRepo, userRepo)), th.CallbackDataPrefix("g::ttt::move::"))

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
