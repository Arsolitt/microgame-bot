package main

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/bot"
	"microgame-bot/internal/core/database"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/mdw"
	"os"
	"os/signal"

	th "github.com/mymmrac/telego/telegohandler"

	"microgame-bot/internal/handlers"
	memoryLocker "microgame-bot/internal/locker/memory"
	gormRPSRepository "microgame-bot/internal/repo/game/rps"
	gormTTTRepository "microgame-bot/internal/repo/game/ttt"
	gormSessionRepository "microgame-bot/internal/repo/session"
	gormUserRepository "microgame-bot/internal/repo/user"
	uowGorm "microgame-bot/internal/uow"
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

	db, err := database.MustInit(cfg)
	if err != nil {
		return err
	}

	userLocker := memoryLocker.New[user.ID]()

	bh, err := bot.MustInit(ctx, cfg)
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

	// Game selector
	bh.HandleInlineQuery(handlers.WrapInlineQuery(handlers.GameSelector(cfg.App)), th.AnyInlineQuery())

	// RPS GAME HANDLERS
	rpsCreateUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	bh.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.RPSCreate(rpsCreateUnit, cfg.App)),
		th.CallbackDataPrefix("create::rps"),
	)

	rpsG := bh.Group(th.CallbackDataPrefix("g::rps::"))

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

	// TTT GAME HANDLERS
	tttCreateUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
	)
	bh.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTCreate(tttCreateUnit, cfg.App)),
		th.CallbackDataPrefix("create::ttt"),
	)

	tttG := bh.Group(th.CallbackDataPrefix("g::ttt::"))

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

	// Empty callback handler
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
