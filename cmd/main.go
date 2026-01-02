package main

import (
	"context"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/bot"
	"microgame-bot/internal/core/database"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/mdw"
	"microgame-bot/internal/queue"
	"microgame-bot/internal/scheduler"
	"os"
	"os/signal"
	"time"

	th "github.com/mymmrac/telego/telegohandler"

	"microgame-bot/internal/handlers"
	memoryLocker "microgame-bot/internal/locker/memory"
	qHandlers "microgame-bot/internal/queue/handlers"
	gormBetRepository "microgame-bot/internal/repo/bet"
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
	betRepo := gormBetRepository.New(db)

	inlineMsgLocker := memoryLocker.New[domain.InlineMessageID]()

	q := queue.New(db, 10)
	q.Register("queue.cleanup", func(ctx context.Context, _ []byte) error {
		return q.CleanupStuckTasks(ctx)
	})

	// Register bet payout handler
	betPayoutUnit := uowGorm.New(db,
		uowGorm.WithBetRepo(betRepo),
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithUserRepo(userRepo),
		uowGorm.WithTTTRepo(tttRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	q.Register("bets.payout", qHandlers.BetPayoutHandler(betPayoutUnit))

	// Register game timeout handler
	gameTimeoutUnit := uowGorm.New(db,
		uowGorm.WithBetRepo(betRepo),
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithUserRepo(userRepo),
		uowGorm.WithTTTRepo(tttRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	q.Register("games.timeout", qHandlers.GameTimeoutHandler(gameTimeoutUnit, q))

	defer func() { _ = q.Stop(ctx) }()
	q.Start(ctx)

	cronJobs := []scheduler.CronJob{
		{
			Name:       "queue-cleanup",
			Expression: "0 */15 * * * *",
			Status:     scheduler.CronJobStatusActive,
			Subject:    "queue.cleanup",
			Payload:    queue.EmptyPayload,
		},
		{
			Name:       "bets-payout",
			Expression: "0 */5 * * * *",
			Status:     scheduler.CronJobStatusActive,
			Subject:    "bets.payout",
			Payload:    queue.EmptyPayload,
		},
		{
			Name:       "games-timeout",
			Expression: "0 */5 * * * *",
			Status:     scheduler.CronJobStatusActive,
			Subject:    "games.timeout",
			Payload:    queue.EmptyPayload,
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
		mdw.DailyBonusMiddleware(dbmUow),
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
		uowGorm.WithUserRepo(userRepo),
		uowGorm.WithBetRepo(betRepo),
	)
	rpsG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.RPSJoin(userRepo, rpsJoinUnit)),
		th.CallbackDataPrefix("g::rps::join::"),
	)
	rpsChoiceUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
		uowGorm.WithBetRepo(betRepo),
	)
	rpsG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.RPSChoice(userRepo, rpsChoiceUnit, q)),
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
		uowGorm.WithUserRepo(userRepo),
		uowGorm.WithBetRepo(betRepo),
	)
	tttG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTJoin(userRepo, tttJoinUnit)),
		th.CallbackDataPrefix("g::ttt::join::"),
	)

	tttMoveUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
		uowGorm.WithBetRepo(betRepo),
	)
	tttG.HandleCallbackQuery(
		handlers.WrapCallbackQuery(handlers.TTTMove(userRepo, tttMoveUnit, q)),
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
