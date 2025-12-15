package main

import (
	"context"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
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
	gormGSRepository "microgame-bot/internal/repo/gs/gorm"
	gormRPSRepository "microgame-bot/internal/repo/rps/gorm"
	gormTTTRepository "microgame-bot/internal/repo/ttt/gorm"
	gormUserRepository "microgame-bot/internal/repo/user/gorm"
	uowGorm "microgame-bot/internal/uow/gorm"

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

	err = db.AutoMigrate(&gormUserRepository.User{})
	if err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}
	err = db.AutoMigrate(&gormGSRepository.GameSession{})
	if err != nil {
		return fmt.Errorf("failed to migrate game session table: %w", err)
	}
	err = db.AutoMigrate(&gormTTTRepository.TTT{})
	if err != nil {
		return fmt.Errorf("failed to migrate ttt table: %w", err)
	}
	err = db.AutoMigrate(&gormRPSRepository.RPS{})
	if err != nil {
		return fmt.Errorf("failed to migrate rps table: %w", err)
	}

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
	_ = gormTTTRepository.New(db)
	_ = gormRPSRepository.New(db)
	_ = gormGSRepository.New(db)

	inlineMsgLocker := memoryLocker.New[domain.InlineMessageID]()

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.InlineMsgProvider(inlineMsgLocker),
		mdw.UserProvider(userLocker, userRepo),
	)

	bh.HandleInlineQuery(handlers.WrapInlineQuery(handlers.GameSelector()), th.AnyInlineQuery())

	// tttG := bh.Group(th.CallbackDataPrefix("g::ttt::"))
	// tttG.Use(mdw.GameProvider(memoryLocker.New[ttt.ID](), tttRepo, gsRepo, "ttt"))

	// tttG.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTJoin(tttRepo, userRepo)), th.CallbackDataPrefix("g::ttt::join::"))
	// tttG.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTMove(tttRepo, userRepo)), th.CallbackDataPrefix("g::ttt::move::"))

	rpsG := bh.Group(th.CallbackDataPrefix("g::rps::"))
	// rpsG.Use(mdw.GameProvider(memoryLocker.New[rps.ID](), rpsRepo, gsRepo, "rps"))

	rpsJoinUnit := uowGorm.New(db,
		uowGorm.WithGSRepo(gormGSRepository.New(db)),
		uowGorm.WithRPSRepo(gormRPSRepository.New(db)),
	)
	rpsG.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.RPSJoin(userRepo, rpsJoinUnit)), th.CallbackDataPrefix("g::rps::join::"))
	rpsChoiceUnit := uowGorm.New(db,
		uowGorm.WithGSRepo(gormGSRepository.New(db)),
		uowGorm.WithRPSRepo(gormRPSRepository.New(db)),
	)
	rpsG.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.RPSChoice(userRepo, rpsChoiceUnit)), th.CallbackDataPrefix("g::rps::choice::"))

	// tttUow := uowGorm.New(db,
	// 	uowGorm.WithGSRepo(gormGSRepository.New(db)),
	// 	uowGorm.WithTTTRepo(gormTTTRepository.New(db)),
	// )
	// bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTCreate(tttUow)), th.CallbackDataEqual("create::ttt"))

	rpsCreateUnit := uowGorm.New(db,
		uowGorm.WithGSRepo(gormGSRepository.New(db)),
		uowGorm.WithRPSRepo(gormRPSRepository.New(db)),
	)
	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.RPSCreate(rpsCreateUnit)), th.CallbackDataEqual("create::rps"))

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
