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
	"microgame-bot/internal/health"
	"microgame-bot/internal/locker"
	"microgame-bot/internal/mdw"
	"microgame-bot/internal/queue"
	"microgame-bot/internal/scheduler"
	"net/http"
	"os"
	"os/signal"
	"time"

	th "github.com/mymmrac/telego/telegohandler"

	"microgame-bot/internal/handlers"
	gormLocker "microgame-bot/internal/locker/gorm"
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
	var inlineMsgLocker locker.ILocker[domain.InlineMessageID]
	var userLocker locker.ILocker[user.ID]
	if cfg.App.LockerDriver == "gorm" {
		userLocker, err = gormLocker.New(db, func(id user.ID) string {
			return id.String()
		})
		if err != nil {
			return fmt.Errorf("failed to create user locker: %w", err)
		}
		inlineMsgLocker, err = gormLocker.New(db, func(id domain.InlineMessageID) string {
			return id.String()
		})
		if err != nil {
			return fmt.Errorf("failed to create inline message locker: %w", err)
		}
	} else {
		userLocker = memoryLocker.New[user.ID]()
		inlineMsgLocker = memoryLocker.New[domain.InlineMessageID]()
	}
	slog.Info("Locker initialized successfully", "driver", cfg.App.LockerDriver)

	healthHandler := health.NewHandler(5 * time.Second)
	healthHandler.RegisterChecker("database", health.NewDatabaseChecker(db))
	healthHandler.RegisterChecker("queue", health.NewQueueChecker(db))
	healthHandler.RegisterChecker("scheduler", health.NewSchedulerChecker(db))
	slog.Info("Health check initialized successfully")

	bot, bh, webhookSrv, err := bot.MustInit(ctx, cfg, &bot.InitOptions{
		HealthHandler: healthHandler,
	})
	if err != nil {
		return err
	}
	defer func() { _ = bh.StopWithContext(ctx) }()

	if webhookSrv != nil {
		go func() {
			slog.Info("Starting webhook server", "addr", webhookSrv.Addr)
			if err := webhookSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Webhook server error", logger.ErrorField, err.Error())
			}
		}()
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := webhookSrv.Shutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown webhook server", logger.ErrorField, err.Error())
			}
		}()
	}

	bufferedHandler := handlers.NewBufferedHandler(
		//nolint:mnd // requests per second (Telegram limit)
		30,
		//nolint:mnd // burst capacity
		10,
		//nolint:mnd // queue size
		10000,
		//nolint:mnd // worker goroutines
		5,
		//nolint:mnd // max retries
		3,
	)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := bufferedHandler.Shutdown(shutdownCtx); err != nil {
			slog.Error("Failed to shutdown buffered handler gracefully", logger.ErrorField, err.Error())
		}
	}()

	wrap := handlers.NewHandlerWrapper(bufferedHandler)

	userRepo := gormUserRepository.New(db)
	tttRepo := gormTTTRepository.New(db)
	rpsRepo := gormRPSRepository.New(db)
	sessionRepo := gormSessionRepository.New(db)
	claimRepo := gormClaimRepository.New(db)
	betRepo := gormBetRepository.New(db)

	q := queue.New(db, 10)
	q.Register("queue.cleanup", func(ctx context.Context, _ []byte) error {
		return q.CleanupStuckTasks(ctx)
	})

	// Register profile load handler
	profileLoadUnit := uowGorm.New(db,
		uowGorm.WithUserRepo(userRepo),
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
		uowGorm.WithTTTRepo(tttRepo),
	)
	q.Register("profile.load", qHandlers.ProfileLoadHandler(profileLoadUnit, bot))

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
	q.Register("locks.cleanup", qHandlers.LockCleanupHandler(userLocker, cfg.App.LockerTTL))

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
		{
			Name:       "locks-cleanup",
			Expression: "0 33 0 * * *",
			Status:     scheduler.CronJobStatusActive,
			Subject:    "locks.cleanup",
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

	// Selector
	bh.HandleInlineQuery(wrap.WrapInlineQuery(handlers.GameSelector(cfg.App)), th.AnyInlineQuery())

	// Profile handler
	bh.HandleChosenInlineResult(
		wrap.WrapChosenInlineResult(handlers.ProfileChosen(q)),
		handlers.ChosenInlineResultID("profile"),
	)

	// RPS GAME HANDLERS
	rpsCreateUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
	)
	bh.HandleCallbackQuery(
		wrap.WrapCallbackQuery(handlers.RPSCreate(rpsCreateUnit, cfg.App)),
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
		wrap.WrapCallbackQuery(handlers.RPSJoin(userRepo, rpsJoinUnit)),
		th.CallbackDataPrefix("g::rps::join::"),
	)
	rpsChoiceUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithRPSRepo(rpsRepo),
		uowGorm.WithBetRepo(betRepo),
	)
	rpsG.HandleCallbackQuery(
		wrap.WrapCallbackQuery(handlers.RPSChoice(userRepo, rpsChoiceUnit, q)),
		th.CallbackDataPrefix("g::rps::choice::"),
	)

	// TTT GAME HANDLERS
	tttCreateUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
	)
	bh.HandleCallbackQuery(
		wrap.WrapCallbackQuery(handlers.TTTCreate(tttCreateUnit, cfg.App)),
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
		wrap.WrapCallbackQuery(handlers.TTTJoin(userRepo, tttJoinUnit)),
		th.CallbackDataPrefix("g::ttt::join::"),
	)

	tttMoveUnit := uowGorm.New(db,
		uowGorm.WithSessionRepo(sessionRepo),
		uowGorm.WithTTTRepo(tttRepo),
		uowGorm.WithBetRepo(betRepo),
	)
	tttG.HandleCallbackQuery(
		wrap.WrapCallbackQuery(handlers.TTTMove(userRepo, tttMoveUnit, q)),
		th.CallbackDataPrefix("g::ttt::move::"),
	)

	tttG.HandleCallbackQuery(
		wrap.WrapCallbackQuery(handlers.TTTRebuild(userRepo, tttRepo)),
		th.CallbackDataPrefix("g::ttt::rebuild::"),
	)

	// Empty callback handler
	bh.HandleCallbackQuery(wrap.WrapCallbackQuery(handlers.Empty()), th.CallbackDataEqual("empty"))

	// Start handling updates
	return bh.Start()
}

func main() {
	if err := startup(); err != nil {
		slog.Error("Failed to startup application", logger.ErrorField, err.Error())
		os.Exit(1)
	}
}
