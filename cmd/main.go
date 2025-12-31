package main

import (
	"context"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/bot"
	"microgame-bot/internal/core/database"
	"microgame-bot/internal/core/kv"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/mdw"
	"microgame-bot/internal/queue"
	"microgame-bot/internal/scheduler"
	"microgame-bot/internal/utils"
	"os"
	"os/signal"
	"time"

	th "github.com/mymmrac/telego/telegohandler"

	"microgame-bot/internal/handlers"
	memoryLocker "microgame-bot/internal/locker/memory"
	natsMetastore "microgame-bot/internal/metastore/nats"
	gormClaimRepository "microgame-bot/internal/repo/claim"
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

	natsConn, err := kv.MustInit(ctx, cfg.Nats)
	if err != nil {
		return err
	}
	defer natsConn.Close()

	claimMetastore, err := natsMetastore.New(ctx, natsConn, "claims_cache", 3)
	if err != nil {
		return err
	}

	userLocker := memoryLocker.New[user.ID]()

	bh, err := bot.MustInit(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = bh.StopWithContext(ctx) }()

	userRepo := gormUserRepository.New(db)
	tttRepo := gormTTTRepository.New(db)
	rpsRepo := gormRPSRepository.New(db)
	sessionRepo := gormSessionRepository.New(db)
	claimRepo := gormClaimRepository.New(db)

	inlineMsgLocker := memoryLocker.New[domain.InlineMessageID]()

	q := queue.New(db, 10)
	q.Register("test", func(ctx context.Context, data []byte) error {
		slog.InfoContext(ctx, "Test message received", "data", string(data))
		// if utils.RandInt(2) == 0 {
		// 	return fmt.Errorf("test error")
		// }
		time.Sleep(10 * time.Second)
		return nil
	})
	q.Register("queue:cleanup", func(ctx context.Context, data []byte) error {
		return q.CleanupStuckTasks(ctx, 15*time.Minute)
	})
	defer func() { _ = q.Stop(ctx) }()
	q.Start(ctx)

	cronJobs := []scheduler.CronJob{
		{
			ID:         utils.NewUniqueID(),
			Name:       "queue-cleanup",
			Expression: "0 */15 * * * *",
			Status:     scheduler.CronJobStatusActive,
			Subject:    "queue:cleanup",
			Payload:    utils.EmptyPayload,
		},
	}
	sc := scheduler.New(db, 10, q, 1*time.Second)

	err = sc.CreateOrUpdateCronJobs(ctx, cronJobs)
	if err != nil {
		return fmt.Errorf("failed to create or update cron jobs: %w", err)
	}

	defer func() { _ = sc.Stop(ctx) }()
	err = sc.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	dbmUow := uowGorm.New(db,
		uowGorm.WithUserRepo(userRepo),
		uowGorm.WithClaimRepo(claimRepo),
	)

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.InlineMsgProvider(inlineMsgLocker),
		mdw.UserProvider(userLocker, userRepo),
		mdw.DailyBonusMiddleware(dbmUow, claimMetastore),
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
