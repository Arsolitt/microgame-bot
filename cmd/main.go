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

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	memoryFSM "microgame-bot/internal/fsm/memory"
	"microgame-bot/internal/handlers"
	memoryLocker "microgame-bot/internal/locker/memory"
	gormGameRepository "microgame-bot/internal/repo/game"
	gormRPSRepository "microgame-bot/internal/repo/game/rps"
	gormTTTRepository "microgame-bot/internal/repo/game/ttt"
	gormSessionRepository "microgame-bot/internal/repo/session"
	gormUserRepository "microgame-bot/internal/repo/user"
	uowGorm "microgame-bot/internal/uow"

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
		return err
	}

	err = db.AutoMigrate(&gormUserRepository.User{})
	if err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}
	err = db.AutoMigrate(&gormSessionRepository.Session{})
	if err != nil {
		return fmt.Errorf("failed to migrate game session table: %w", err)
	}
	err = db.AutoMigrate(&gormGameRepository.Game{})
	if err != nil {
		return fmt.Errorf("failed to migrate game table: %w", err)
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

	bh, err := th.NewBotHandler(
		bot,
		updates,
		th.WithErrorHandler(func(ctx *th.Context, update telego.Update, err error) {
			slog.ErrorContext(ctx, "Handler error occurred", logger.ErrorField, err.Error())
		}),
	)
	if err != nil {
		return err
	}
	defer bh.StopWithContext(ctx)

	userRepo := gormUserRepository.New(db)
	tttRepo := gormTTTRepository.New(db)
	rpsRepo := gormRPSRepository.New(db)
	sessionRepo := gormSessionRepository.New(db)

	inlineMsgLocker := memoryLocker.New[domain.InlineMessageID]()

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.InlineMsgProvider(inlineMsgLocker),
		mdw.UserProvider(userLocker, userRepo),
	)

	bh.HandleInlineQuery(handlers.WrapInlineQuery(handlers.GameSelector(cfg.App)), th.AnyInlineQuery())

	tttG := bh.Group(th.CallbackDataPrefix("g::ttt::"))
	// tttG.Use(mdw.GameProvider(memoryLocker.New[ttt.ID](), tttRepo, gsRepo, "ttt"))

	tttJoinUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
	)
	tttG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTJoin(userRepo, tttJoinUnit)),
		th.CallbackDataPrefix("g::ttt::join::"),
	)

	tttMoveUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
	)
	tttG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTMove(userRepo, tttMoveUnit)),
		th.CallbackDataPrefix("g::ttt::move::"),
	)

	tttG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTRebuild(userRepo, tttRepo)),
		th.CallbackDataPrefix("g::ttt::rebuild::"),
	)

	rpsG := bh.Group(th.CallbackDataPrefix("g::rps::"))
	// rpsG.Use(mdw.GameProvider(memoryLocker.New[rps.ID](), rpsRepo, gsRepo, "rps"))

	rpsJoinUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	rpsG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.RPSJoin(userRepo, rpsJoinUnit)),
		th.CallbackDataPrefix("g::rps::join::"),
	)
	rpsChoiceUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	rpsG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.RPSChoice(userRepo, rpsChoiceUnit)),
		th.CallbackDataPrefix("g::rps::choice::"),
	)

	tttUow := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
	)
	bh.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTCreate(tttUow, cfg.App)),
		th.CallbackDataPrefix("create::ttt"),
	)

	rpsCreateUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	bh.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.RPSCreate(rpsCreateUnit, cfg.App)),
		th.CallbackDataPrefix("create::rps"),
	)

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
