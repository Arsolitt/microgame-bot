package main

import (
	"context"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/mdw"
	"os"
	"os/signal"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	memoryFSM "minigame-bot/internal/fsm/memory"
	"minigame-bot/internal/handlers"
	memoryLocker "minigame-bot/internal/locker/memory"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"
	memoryUserRepository "minigame-bot/internal/repo/user/memory"
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

	_ = memoryLocker.New()
	_ = memoryFSM.New()

	bot, err := telego.NewBot(string(cfg.Telegram.Token))
	if err != nil {
		return err
	}

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return err
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return err
	}
	defer bh.StopWithContext(ctx)

	userRepo := memoryUserRepository.New()
	gameRepo := memoryTTTRepository.New()

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.UserProvider(memoryLocker.New(), userRepo),
	)

	bh.HandleChosenInlineResult(handlers.WrapChosenInlineResult(handlers.GameCreator(gameRepo)), th.AnyChosenInlineResult())

	bh.HandleInlineQuery(handlers.WrapInlineQuery(handlers.GameSelector()), th.AnyInlineQuery())

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTJoin(gameRepo, userRepo)), th.CallbackDataPrefix("ttt::join::"))

	bh.HandleCallbackQuery(handlers.WrapCallbackQuery(handlers.TTTMove(gameRepo, userRepo)), th.CallbackDataPrefix("ttt::move::"))

	// Start handling updates
	return bh.Start()
}

func main() {
	if err := startup(); err != nil {
		slog.Error("Failed to startup application", logger.ErrorField, err.Error())
		os.Exit(1)
	}
}
