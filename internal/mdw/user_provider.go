package mdw

import (
	"errors"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/locker"
	repository "microgame-bot/internal/repo/user"
	"strconv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func UserProvider(
	locker locker.ILocker[domainUser.ID],
	userRepo repository.IUserRepository,
) func(ctx *th.Context, update telego.Update) error {
	const operationName = "middleware::user_provider"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, update telego.Update) error {
		l.DebugContext(ctx, "UserProvider middleware started")
		var userTelegramID int64
		var firstName string
		var lastName string
		var username string
		var privateChatID *int64
		var groupChatID *int64
		var chatInstance *int64
		if update.Message != nil {
			userTelegramID = update.Message.From.ID
			firstName = update.Message.From.FirstName
			lastName = update.Message.From.LastName
			username = update.Message.From.Username
			if update.Message.Chat.Type == "private" {
				privateChatID = &update.Message.Chat.ID
			} else {
				groupChatID = &update.Message.Chat.ID
			}
		} else if update.CallbackQuery != nil {
			userTelegramID = update.CallbackQuery.From.ID
			firstName = update.CallbackQuery.From.FirstName
			lastName = update.CallbackQuery.From.LastName
			username = update.CallbackQuery.From.Username
			if update.CallbackQuery.Message != nil && update.CallbackQuery.Message.IsAccessible() {
				chat := update.CallbackQuery.Message.GetChat()
				if chat.Type == "private" {
					privateChatID = &chat.ID
				} else {
					groupChatID = &chat.ID
				}
			}
			parsedChatInstance, err := strconv.ParseInt(update.CallbackQuery.ChatInstance, 10, 64)
			if err != nil {
				return err
			}
			chatInstance = &parsedChatInstance
		} else if update.InlineQuery != nil {
			userTelegramID = update.InlineQuery.From.ID
			firstName = update.InlineQuery.From.FirstName
			lastName = update.InlineQuery.From.LastName
			username = update.InlineQuery.From.Username
		} else {
			return core.ErrInvalidUpdate
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.UserTelegramIDField, userTelegramID)
		rawCtx = logger.WithLogValue(rawCtx, logger.UserFirstNameField, firstName)
		rawCtx = logger.WithLogValue(rawCtx, logger.UserLastNameField, lastName)
		rawCtx = logger.WithLogValue(rawCtx, logger.UserNameField, username)
		rawCtx = logger.WithLogValue(rawCtx, logger.PrivateChatIDField, privateChatID)
		rawCtx = logger.WithLogValue(rawCtx, logger.GroupChatIDField, groupChatID)
		rawCtx = logger.WithLogValue(rawCtx, logger.ChatInstanceField, chatInstance)

		ctx = ctx.WithContext(rawCtx)

		var user domainUser.User
		user, err := userRepo.UserByTelegramID(ctx, userTelegramID)
		if err != nil {
			if errors.Is(err, core.ErrUserNotFound) {
				l.InfoContext(ctx, "User not found, creating new")
				buildedUser, err := domainUser.New(
					domainUser.WithNewID(),
					domainUser.WithTelegramIDFromInt(userTelegramID),
					domainUser.WithFirstName(domainUser.FirstName(firstName)),
					domainUser.WithLastName(domainUser.LastName(lastName)),
					domainUser.WithUsername(domainUser.Username(username)),
					domainUser.WithChatIDFromPointer(privateChatID),
					domainUser.WithTokens(domain.StartBonusTokens),
				)
				if err != nil {
					return err
				}
				dbUser, err := userRepo.CreateUser(ctx, buildedUser)
				if err != nil {
					return err
				}
				l.InfoContext(ctx, "User created")
				user = dbUser
			} else {
				return err
			}
		}

		rawCtx = logger.WithLogValue(rawCtx, logger.UserIDField, user.ID().String())
		ctx = ctx.WithContext(rawCtx)
		ctx = ctx.WithValue(core.ContextKeyUser, user)

		err = locker.Lock(ctx, user.ID())
		if err != nil {
			return err
		}
		defer func() { _ = locker.Unlock(ctx, user.ID()) }()

		l.DebugContext(ctx, "UserProvider middleware finished")

		return ctx.Next(update)
	}
}
