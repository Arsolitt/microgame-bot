# Unit of Work Pattern with GORM

## TL;DR

**Ключевая идея:** Инкапсулируй транзакции и репозитории в `UnitOfWork`, добавь `*Locked` методы для `SELECT FOR UPDATE`.

**Использование:**
```go
// Simple read - no transaction
user, err := service.uow.UserRepo().UserByID(ctx, id)

// Complex operation - automatic transaction
err := service.uow.Do(ctx, func(uow repo.UnitOfWork) error {
    // SELECT FOR UPDATE
    game, err := uow.GameRepo().GameByIDLocked(ctx, gameID)
    
    // Business logic
    game.Update()
    
    // Save
    return uow.GameRepo().Update(ctx, game)
})
```

**Тестирование:**
```go
mockUow := mock.NewMockUnitOfWork()
mockRepo := mock.NewMockRepository()
mockUow.On("UserRepo").Return(mockRepo)
// No GORM dependency!
```

---

## Проблема

При работе с транзакциями в паттерне Repository возникает вопрос: как передать транзакцию между несколькими репозиториями, чтобы они работали в рамках одной транзакции, при этом сохраняя возможность легко писать юнит-тесты?

**Плохое решение:**
```go
// ❌ Сервис зависит от *gorm.DB напрямую
func (s *Service) UpdateUserAndGame(ctx context.Context) error {
    tx := s.db.Begin()
    defer tx.Rollback()
    
    if err := s.userRepo.Update(tx, user); err != nil {
        return err
    }
    
    if err := s.gameRepo.Update(tx, game); err != nil {
        return err
    }
    
    return tx.Commit().Error
}
```

Проблемы:
- Сервис знает о GORM
- Сложно тестировать
- Дублирование логики управления транзакцией

## Решение: Unit of Work

Unit of Work (UoW) — паттерн, который:
1. Инкапсулирует начало/завершение транзакции
2. Предоставляет доступ ко всем репозиториям в рамках одной транзакции
3. Легко мокается для тестов

## Реализация

### 1. Интерфейс Unit of Work

```go
// internal/repo/unit_of_work.go
package repo

import (
	"context"
	"gorm.io/gorm"
	userRepo "microgame-bot/internal/repo/user"
	tttRepo "microgame-bot/internal/repo/ttt"
)

// UnitOfWork provides transactional operations over multiple repositories
type UnitOfWork interface {
	// Do executes function within a transaction
	Do(ctx context.Context, fn func(uow UnitOfWork) error) error
	
	// Repository accessors
	UserRepo() userRepo.IUserRepository
	TTTRepo() tttRepo.ITTTRepository
}

type unitOfWork struct {
	db       *gorm.DB
	userRepo userRepo.IUserRepository
	tttRepo  tttRepo.ITTTRepository
}

// NewUnitOfWork creates a new unit of work instance
func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &unitOfWork{
		db:       db,
		userRepo: userRepo.NewGormRepository(db),
		tttRepo:  tttRepo.NewGormRepository(db),
	}
}

// Do executes function within a transaction
func (u *unitOfWork) Do(ctx context.Context, fn func(uow UnitOfWork) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txUow := &unitOfWork{
			db:       tx,
			userRepo: userRepo.NewGormRepository(tx),
			tttRepo:  tttRepo.NewGormRepository(tx),
		}
		return fn(txUow)
	})
}

func (u *unitOfWork) UserRepo() userRepo.IUserRepository {
	return u.userRepo
}

func (u *unitOfWork) TTTRepo() tttRepo.ITTTRepository {
	return u.tttRepo
}
```

### 2. Обновление интерфейсов репозиториев

Добавляем методы с `Locked` суффиксом для SELECT FOR UPDATE:

```go
// internal/repo/user/repository.go
package repository

import (
	"context"
	domainUser "microgame-bot/internal/domain/user"
)

type IUserGetter interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error)
	
	// Locked versions - work only within transaction
	UserByIDLocked(ctx context.Context, id domainUser.ID) (domainUser.User, error)
	UserByTelegramIDLocked(ctx context.Context, telegramID int64) (domainUser.User, error)
}

type IUserCreator interface {
	CreateUser(ctx context.Context, user domainUser.User) error
}

type IUserUpdater interface {
	UpdateUser(ctx context.Context, user domainUser.User) error
}

type IUserRepository interface {
	IUserGetter
	IUserCreator
	IUserUpdater
}
```

```go
// internal/repo/ttt/repository.go
package ttt

import (
	"context"
	"microgame-bot/internal/domain/ttt"
)

type ITTTGetter interface {
	GameByMessageID(ctx context.Context, id ttt.InlineMessageID) (ttt.TTT, error)
	GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error)
	
	// Locked version - works only within transaction
	GameByIDLocked(ctx context.Context, id ttt.ID) (ttt.TTT, error)
}

type ITTTCreator interface {
	CreateGame(ctx context.Context, game ttt.TTT) error
}

type ITTTUpdater interface {
	UpdateGame(ctx context.Context, game ttt.TTT) error
}

type ITTTRepository interface {
	ITTTCreator
	ITTTUpdater
	ITTTGetter
}
```

### 3. Реализация в GORM репозитории

```go
// internal/repo/user/gorm_repository.go
package user

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	domainUser "microgame-bot/internal/domain/user"
	repository "microgame-bot/internal/repo/user"
)

var ErrNotInTransaction = errors.New("locked methods can only be called within a transaction")

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) repository.IUserRepository {
	return &gormRepository{db: db}
}

// isInTransaction checks if current db instance is in transaction
func (r *gormRepository) isInTransaction() bool {
	committer, ok := r.db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}

func (r *gormRepository) UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error) {
	model, err := gorm.G[domainUser.GUserModel](r.db).
		Where("telegram_id = ?", telegramID).
		First(ctx)
	if err != nil {
		return domainUser.User{}, err
	}
	return model.ToDomain()
}

func (r *gormRepository) UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error) {
	model, err := gorm.G[domainUser.GUserModel](r.db).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		return domainUser.User{}, err
	}
	return model.ToDomain()
}

// UserByIDLocked - SELECT FOR UPDATE, works only in transaction
func (r *gormRepository) UserByIDLocked(ctx context.Context, id domainUser.ID) (domainUser.User, error) {
	if !r.isInTransaction() {
		return domainUser.User{}, ErrNotInTransaction
	}
	
	var model domainUser.GUserModel
	err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id.String()).
		First(&model).Error
	if err != nil {
		return domainUser.User{}, err
	}
	
	return model.ToDomain()
}

// UserByTelegramIDLocked - SELECT FOR UPDATE, works only in transaction
func (r *gormRepository) UserByTelegramIDLocked(ctx context.Context, telegramID int64) (domainUser.User, error) {
	if !r.isInTransaction() {
		return domainUser.User{}, ErrNotInTransaction
	}
	
	var model domainUser.GUserModel
	err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("telegram_id = ?", telegramID).
		First(&model).Error
	if err != nil {
		return domainUser.User{}, err
	}
	
	return model.ToDomain()
}

func (r *gormRepository) CreateUser(ctx context.Context, user domainUser.User) error {
	model := user.ToModel()
	return gorm.G[domainUser.GUserModel](r.db).Create(ctx, &model)
}

func (r *gormRepository) UpdateUser(ctx context.Context, user domainUser.User) error {
	model := user.ToModel()
	return gorm.G[domainUser.GUserModel](r.db).
		Where("id = ?", model.ID).
		Updates(ctx, model)
}
```

```go
// internal/repo/ttt/gorm_repository.go
package ttt

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	domainTTT "microgame-bot/internal/domain/ttt"
	repository "microgame-bot/internal/repo/ttt"
)

var ErrNotInTransaction = errors.New("locked methods can only be called within a transaction")

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) repository.ITTTRepository {
	return &gormRepository{db: db}
}

func (r *gormRepository) isInTransaction() bool {
	committer, ok := r.db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}

func (r *gormRepository) GameByID(ctx context.Context, id domainTTT.ID) (domainTTT.TTT, error) {
	model, err := gorm.G[domainTTT.GGameModel](r.db).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		return domainTTT.TTT{}, err
	}
	return model.ToDomain()
}

// GameByIDLocked - SELECT FOR UPDATE, works only in transaction
func (r *gormRepository) GameByIDLocked(ctx context.Context, id domainTTT.ID) (domainTTT.TTT, error) {
	if !r.isInTransaction() {
		return domainTTT.TTT{}, ErrNotInTransaction
	}
	
	var model domainTTT.GGameModel
	err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id.String()).
		First(&model).Error
	if err != nil {
		return domainTTT.TTT{}, err
	}
	
	return model.ToDomain()
}

func (r *gormRepository) GameByMessageID(ctx context.Context, id domainTTT.InlineMessageID) (domainTTT.TTT, error) {
	model, err := gorm.G[domainTTT.GGameModel](r.db).
		Where("inline_message_id = ?", string(id)).
		First(ctx)
	if err != nil {
		return domainTTT.TTT{}, err
	}
	return model.ToDomain()
}

func (r *gormRepository) CreateGame(ctx context.Context, game domainTTT.TTT) error {
	model := game.ToModel()
	return gorm.G[domainTTT.GGameModel](r.db).Create(ctx, &model)
}

func (r *gormRepository) UpdateGame(ctx context.Context, game domainTTT.TTT) error {
	model := game.ToModel()
	return gorm.G[domainTTT.GGameModel](r.db).
		Where("id = ?", model.ID).
		Updates(ctx, model)
}
```

### 3. Использование в сервисе

```go
// internal/service/game_service.go
package service

import (
	"context"
	"errors"
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/repo"
)

type GameService struct {
	uow repo.UnitOfWork
}

func NewGameService(uow repo.UnitOfWork) *GameService {
	return &GameService{uow: uow}
}

// Simple operation without transaction (reads only)
func (s *GameService) GetUserGame(ctx context.Context, telegramID int64, gameID ttt.ID) (user.User, ttt.TTT, error) {
	usr, err := s.uow.UserRepo().UserByTelegramID(ctx, telegramID)
	if err != nil {
		return user.User{}, ttt.TTT{}, err
	}
	
	game, err := s.uow.TTTRepo().GameByID(ctx, gameID)
	if err != nil {
		return user.User{}, ttt.TTT{}, err
	}
	
	return usr, game, nil
}

// Complex operation with transaction
func (s *GameService) CreateGameForUser(ctx context.Context, telegramID int64, game ttt.TTT) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		// Find or create user
		usr, err := uow.UserRepo().UserByTelegramID(ctx, telegramID)
		if err != nil {
			// User not found, create new
			usr, err = user.NewBuilder().
				TelegramIDFromInt(telegramID).
				// ... other fields
				Build()
			if err != nil {
				return err
			}
			
			if err := uow.UserRepo().CreateUser(ctx, usr); err != nil {
				return err
			}
		}
		
		// Create game
		if err := uow.TTTRepo().CreateGame(ctx, game); err != nil {
			return err
		}
		
		return nil
	})
}

// Transaction with SELECT FOR UPDATE
func (s *GameService) MakeMove(ctx context.Context, gameID ttt.ID, playerID user.ID, move int) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		// SELECT FOR UPDATE through repository
		game, err := uow.TTTRepo().GameByIDLocked(ctx, gameID)
		if err != nil {
			return err
		}
		
		// Business logic
		if game.IsFinished() {
			return errors.New("game already finished")
		}
		
		// Make move
		game, err = game.MakeMove(playerID, move)
		if err != nil {
			return err
		}
		
		// Save
		return uow.TTTRepo().UpdateGame(ctx, game)
	})
}

// Multiple entities update in single transaction
func (s *GameService) FinishGameAndUpdateStats(ctx context.Context, gameID ttt.ID) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		game, err := uow.TTTRepo().GameByID(ctx, gameID)
		if err != nil {
			return err
		}
		
		// Get players
		player1, err := uow.UserRepo().UserByID(ctx, game.Player1ID())
		if err != nil {
			return err
		}
		
		player2, err := uow.UserRepo().UserByID(ctx, game.Player2ID())
		if err != nil {
			return err
		}
		
		// Update game status
		game = game.Finish()
		if err := uow.TTTRepo().UpdateGame(ctx, game); err != nil {
			return err
		}
		
		// Update users stats (example)
		player1 = player1.UpdateTimestamp()
		player2 = player2.UpdateTimestamp()
		
		if err := uow.UserRepo().UpdateUser(ctx, player1); err != nil {
			return err
		}
		
		if err := uow.UserRepo().UpdateUser(ctx, player2); err != nil {
			return err
		}
		
		return nil
	})
}
```

## Юнит-тесты

### 1. Mock для Unit of Work

```go
// internal/repo/mock/unit_of_work_mock.go
package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"microgame-bot/internal/repo"
	userRepo "microgame-bot/internal/repo/user"
	tttRepo "microgame-bot/internal/repo/ttt"
)

type MockUnitOfWork struct {
	mock.Mock
}

func NewMockUnitOfWork() *MockUnitOfWork {
	return &MockUnitOfWork{}
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(uow repo.UnitOfWork) error) error {
	// Call the function with self to allow testing transaction logic
	return fn(m)
}

func (m *MockUnitOfWork) UserRepo() userRepo.IUserRepository {
	args := m.Called()
	return args.Get(0).(userRepo.IUserRepository)
}

func (m *MockUnitOfWork) TTTRepo() tttRepo.ITTTRepository {
	args := m.Called()
	return args.Get(0).(tttRepo.ITTTRepository)
}
```

### 2. Mock для репозиториев

```go
// internal/repo/user/mock/user_repository_mock.go
package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"microgame-bot/internal/domain/user"
	repository "microgame-bot/internal/repo/user"
)

type MockUserRepository struct {
	mock.Mock
}

var _ repository.IUserRepository = (*MockUserRepository)(nil)

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) UserByTelegramID(ctx context.Context, telegramID int64) (user.User, error) {
	args := m.Called(ctx, telegramID)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) UserByID(ctx context.Context, id user.ID) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) UserByIDLocked(ctx context.Context, id user.ID) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) UserByTelegramIDLocked(ctx context.Context, telegramID int64) (user.User, error) {
	args := m.Called(ctx, telegramID)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
```

```go
// internal/repo/ttt/mock/ttt_repository_mock.go
package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"microgame-bot/internal/domain/ttt"
	repository "microgame-bot/internal/repo/ttt"
)

type MockTTTRepository struct {
	mock.Mock
}

var _ repository.ITTTRepository = (*MockTTTRepository)(nil)

func NewMockTTTRepository() *MockTTTRepository {
	return &MockTTTRepository{}
}

func (m *MockTTTRepository) GameByMessageID(ctx context.Context, id ttt.InlineMessageID) (ttt.TTT, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(ttt.TTT), args.Error(1)
}

func (m *MockTTTRepository) GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(ttt.TTT), args.Error(1)
}

func (m *MockTTTRepository) GameByIDLocked(ctx context.Context, id ttt.ID) (ttt.TTT, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(ttt.TTT), args.Error(1)
}

func (m *MockTTTRepository) CreateGame(ctx context.Context, game ttt.TTT) error {
	args := m.Called(ctx, game)
	return args.Error(0)
}

func (m *MockTTTRepository) UpdateGame(ctx context.Context, game ttt.TTT) error {
	args := m.Called(ctx, game)
	return args.Error(0)
}
```

### 3. Примеры тестов

```go
// internal/service/game_service_test.go
package service

import (
	"context"
	"errors"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
	repoMock "microgame-bot/internal/repo/mock"
	userRepoMock "microgame-bot/internal/repo/user/mock"
	tttRepoMock "microgame-bot/internal/repo/ttt/mock"
)

func TestGameService_GetUserGame_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	telegramID := int64(12345)
	gameID := ttt.ID{} // Mock ID
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockUserRepo := userRepoMock.NewMockUserRepository()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	expectedUser, _ := user.NewBuilder().
		TelegramIDFromInt(telegramID).
		UsernameFromString("testuser").
		Build()
	
	expectedGame := ttt.TTT{} // Mock game
	
	mockUow.On("UserRepo").Return(mockUserRepo)
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	mockUserRepo.On("UserByTelegramID", ctx, telegramID).Return(expectedUser, nil)
	mockTTTRepo.On("GameByID", ctx, gameID).Return(expectedGame, nil)
	
	service := NewGameService(mockUow)
	
	// Act
	usr, game, err := service.GetUserGame(ctx, telegramID, gameID)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, usr)
	assert.Equal(t, expectedGame, game)
	mockUserRepo.AssertExpectations(t)
	mockTTTRepo.AssertExpectations(t)
}

func TestGameService_GetUserGame_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	telegramID := int64(12345)
	gameID := ttt.ID{}
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockUserRepo := userRepoMock.NewMockUserRepository()
	
	mockUow.On("UserRepo").Return(mockUserRepo)
	mockUserRepo.On("UserByTelegramID", ctx, telegramID).
		Return(user.User{}, errors.New("user not found"))
	
	service := NewGameService(mockUow)
	
	// Act
	_, _, err := service.GetUserGame(ctx, telegramID, gameID)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestGameService_CreateGameForUser_NewUser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	telegramID := int64(12345)
	game := ttt.TTT{} // Mock game
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockUserRepo := userRepoMock.NewMockUserRepository()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	mockUow.On("UserRepo").Return(mockUserRepo)
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	
	// User not found
	mockUserRepo.On("UserByTelegramID", ctx, telegramID).
		Return(user.User{}, errors.New("not found"))
	
	// Create user called with any user
	mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("user.User")).
		Return(nil)
	
	// Create game called
	mockTTTRepo.On("CreateGame", ctx, game).Return(nil)
	
	service := NewGameService(mockUow)
	
	// Act
	err := service.CreateGameForUser(ctx, telegramID, game)
	
	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockTTTRepo.AssertExpectations(t)
	mockUserRepo.AssertCalled(t, "CreateUser", ctx, mock.AnythingOfType("user.User"))
}

func TestGameService_CreateGameForUser_ExistingUser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	telegramID := int64(12345)
	game := ttt.TTT{}
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockUserRepo := userRepoMock.NewMockUserRepository()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	existingUser, _ := user.NewBuilder().
		TelegramIDFromInt(telegramID).
		Build()
	
	mockUow.On("UserRepo").Return(mockUserRepo)
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	mockUserRepo.On("UserByTelegramID", ctx, telegramID).Return(existingUser, nil)
	mockTTTRepo.On("CreateGame", ctx, game).Return(nil)
	
	service := NewGameService(mockUow)
	
	// Act
	err := service.CreateGameForUser(ctx, telegramID, game)
	
	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockTTTRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}

func TestGameService_CreateGameForUser_TransactionRollback(t *testing.T) {
	// Arrange
	ctx := context.Background()
	telegramID := int64(12345)
	game := ttt.TTT{}
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockUserRepo := userRepoMock.NewMockUserRepository()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	existingUser, _ := user.NewBuilder().
		TelegramIDFromInt(telegramID).
		Build()
	
	mockUow.On("UserRepo").Return(mockUserRepo)
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	mockUserRepo.On("UserByTelegramID", ctx, telegramID).Return(existingUser, nil)
	mockTTTRepo.On("CreateGame", ctx, game).Return(errors.New("db error"))
	
	service := NewGameService(mockUow)
	
	// Act
	err := service.CreateGameForUser(ctx, telegramID, game)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockTTTRepo.AssertExpectations(t)
}

func TestGameService_MakeMove_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := ttt.ID{}
	playerID := user.ID{}
	move := 0
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	game := ttt.TTT{} // Mock game in progress
	
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	mockTTTRepo.On("GameByIDLocked", ctx, gameID).Return(game, nil)
	mockTTTRepo.On("UpdateGame", ctx, mock.AnythingOfType("ttt.TTT")).Return(nil)
	
	service := NewGameService(mockUow)
	
	// Act
	err := service.MakeMove(ctx, gameID, playerID, move)
	
	// Assert
	assert.NoError(t, err)
	mockTTTRepo.AssertExpectations(t)
	mockTTTRepo.AssertCalled(t, "GameByIDLocked", ctx, gameID)
}

func TestGameService_MakeMove_GameAlreadyFinished(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := ttt.ID{}
	playerID := user.ID{}
	move := 0
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	finishedGame := ttt.TTT{} // Mock finished game
	
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	mockTTTRepo.On("GameByIDLocked", ctx, gameID).Return(finishedGame, nil)
	
	service := NewGameService(mockUow)
	
	// Act
	err := service.MakeMove(ctx, gameID, playerID, move)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, "game already finished", err.Error())
	mockTTTRepo.AssertExpectations(t)
	mockTTTRepo.AssertNotCalled(t, "UpdateGame", mock.Anything, mock.Anything)
}

func TestGameService_FinishGameAndUpdateStats_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := ttt.ID{}
	player1ID := user.ID{}
	player2ID := user.ID{}
	
	mockUow := repoMock.NewMockUnitOfWork()
	mockUserRepo := userRepoMock.NewMockUserRepository()
	mockTTTRepo := tttRepoMock.NewMockTTTRepository()
	
	game := ttt.TTT{} // Mock with player IDs
	player1, _ := user.NewBuilder().TelegramIDFromInt(1).Build()
	player2, _ := user.NewBuilder().TelegramIDFromInt(2).Build()
	
	mockUow.On("TTTRepo").Return(mockTTTRepo)
	mockUow.On("UserRepo").Return(mockUserRepo)
	
	mockTTTRepo.On("GameByID", ctx, gameID).Return(game, nil)
	mockUserRepo.On("UserByID", ctx, player1ID).Return(player1, nil)
	mockUserRepo.On("UserByID", ctx, player2ID).Return(player2, nil)
	mockTTTRepo.On("UpdateGame", ctx, mock.AnythingOfType("ttt.TTT")).Return(nil)
	mockUserRepo.On("UpdateUser", ctx, mock.AnythingOfType("user.User")).Return(nil).Twice()
	
	service := NewGameService(mockUow)
	
	// Act
	err := service.FinishGameAndUpdateStats(ctx, gameID)
	
	// Assert
	assert.NoError(t, err)
	mockTTTRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNumberOfCalls(t, "UpdateUser", 2)
}
```

## Интеграционные тесты

```go
// internal/service/game_service_integration_test.go
//go:build integration
package service

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/repo"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Migrate schemas
	err = db.AutoMigrate(&user.GUserModel{}, &ttt.GGameModel{})
	require.NoError(t, err)
	
	return db
}

func TestGameService_CreateGameForUser_Integration(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	uow := repo.NewUnitOfWork(db)
	service := NewGameService(uow)
	
	ctx := context.Background()
	telegramID := int64(12345)
	game := ttt.TTT{} // Create valid game
	
	// Act
	err := service.CreateGameForUser(ctx, telegramID, game)
	
	// Assert
	assert.NoError(t, err)
	
	// Verify user was created
	var userCount int64
	db.Model(&user.GUserModel{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)
	
	// Verify game was created
	var gameCount int64
	db.Model(&ttt.GGameModel{}).Count(&gameCount)
	assert.Equal(t, int64(1), gameCount)
}

func TestGameService_TransactionRollback_Integration(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	uow := repo.NewUnitOfWork(db)
	
	ctx := context.Background()
	
	// Act - try to create with invalid data to trigger rollback
	err := uow.Do(ctx, func(uow repo.UnitOfWork) error {
		usr, _ := user.NewBuilder().
			TelegramIDFromInt(12345).
			Build()
		
		if err := uow.UserRepo().CreateUser(ctx, usr); err != nil {
			return err
		}
		
		// Force error
		return errors.New("forced error")
	})
	
	// Assert
	assert.Error(t, err)
	
	// Verify nothing was committed
	var userCount int64
	db.Model(&user.GUserModel{}).Count(&userCount)
	assert.Equal(t, int64(0), userCount)
}

func TestRepository_LockedMethodsOutsideTransaction_Integration(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	userRepo := userRepo.NewGormRepository(db)
	
	ctx := context.Background()
	userID := user.ID{}
	
	// Act - try to call locked method outside transaction
	_, err := userRepo.UserByIDLocked(ctx, userID)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, userRepo.ErrNotInTransaction, err)
}

func TestRepository_LockedMethodsInsideTransaction_Integration(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	uow := repo.NewUnitOfWork(db)
	
	ctx := context.Background()
	
	// Create user first
	usr, _ := user.NewBuilder().
		TelegramIDFromInt(12345).
		UsernameFromString("testuser").
		Build()
	
	err := uow.UserRepo().CreateUser(ctx, usr)
	require.NoError(t, err)
	
	// Act - call locked method inside transaction
	err = uow.Do(ctx, func(uow repo.UnitOfWork) error {
		lockedUser, err := uow.UserRepo().UserByIDLocked(ctx, usr.ID())
		if err != nil {
			return err
		}
		
		assert.Equal(t, usr.TelegramID(), lockedUser.TelegramID())
		return nil
	})
	
	// Assert
	assert.NoError(t, err)
}
```

## Преимущества паттерна

### 1. Инкапсуляция транзакций
Логика управления транзакциями в одном месте - в `UnitOfWork.Do()`. Сервисы не знают про `Begin/Commit/Rollback`.

### 2. Полная абстракция от GORM
Сервисы не зависят от деталей реализации БД. Даже `SELECT FOR UPDATE` инкапсулирован в репозиториях через `*Locked` методы.

```go
// Сервис не знает про GORM, clause.Locking, и т.д.
game, err := uow.GameRepo().GameByIDLocked(ctx, id)
```

### 3. Простота тестирования
Легко мокировать через интерфейсы. Не нужна реальная БД для юнит-тестов.

### 4. Consistency
Все операции в рамках одной транзакции или все откатываются. Невозможно забыть `Commit()` или `Rollback()`.

### 5. Безопасность
Методы `*Locked` проверяют наличие транзакции - нельзя случайно вызвать `SELECT FOR UPDATE` вне транзакции.

### 6. Переиспользование
Один UoW для всех сервисов. Легко добавлять новые репозитории - просто добавь метод в интерфейс.

## Best Practices

### 1. Используй транзакции для связанных операций
Любые операции, затрагивающие несколько сущностей, должны быть в транзакции:

```go
// ✅ Good
func (s *Service) UpdateRelated(ctx context.Context) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		// Multiple operations in single transaction
		return nil
	})
}

// ❌ Bad - no transaction, inconsistent state possible
func (s *Service) UpdateRelated(ctx context.Context) error {
	s.uow.UserRepo().Update(ctx, user)
	s.uow.GameRepo().Update(ctx, game)
	return nil
}
```

### 2. Методы *Locked только для конкурентных операций
Используй `*Locked` методы когда несколько процессов могут изменять одни данные:

```go
// ✅ Good - concurrent updates need locking
func (s *Service) DecrementStock(ctx context.Context, productID string) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		product, err := uow.ProductRepo().ProductByIDLocked(ctx, productID)
		if err != nil {
			return err
		}
		
		if product.Stock <= 0 {
			return errors.New("out of stock")
		}
		
		product.Stock--
		return uow.ProductRepo().Update(ctx, product)
	})
}

// ❌ Bad - race condition, two requests can decrement below 0
func (s *Service) DecrementStock(ctx context.Context, productID string) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		product, err := uow.ProductRepo().ProductByID(ctx, productID) // No lock!
		// ... same logic
	})
}
```

### 3. GORM не должен протекать в бизнес-логику
Все специфичные для БД операции инкапсулируй в репозиториях:

```go
// ✅ Good - business logic doesn't know about GORM
func (s *Service) UpdateGame(ctx context.Context) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		game, err := uow.GameRepo().GameByIDLocked(ctx, id)
		// ...
	})
}

// ❌ Bad - GORM leaks to service layer
func (s *Service) UpdateGame(ctx context.Context) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		var model GameModel
		err := uow.DB().Clauses(clause.Locking{Strength: "UPDATE"}).First(&model)
		// ...
	})
}
```

### 4. Методы *Locked должны проверять наличие транзакции
Всегда проверяй `isInTransaction()` в `*Locked` методах:

```go
// ✅ Good - prevents misuse
func (r *repository) UserByIDLocked(ctx context.Context, id ID) (User, error) {
	if !r.isInTransaction() {
		return User{}, ErrNotInTransaction
	}
	// ... SELECT FOR UPDATE
}

// ❌ Bad - silent failure or unexpected behavior
func (r *repository) UserByIDLocked(ctx context.Context, id ID) (User, error) {
	// SELECT FOR UPDATE without transaction - may not lock!
}
```

### 5. Короткие транзакции
Минимизируй время удержания блокировок:

```go
// ✅ Good - short transaction
func (s *Service) ProcessOrder(ctx context.Context) error {
	// Heavy computation OUTSIDE transaction
	processed := s.heavyProcessing(order)
	
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		// Only DB operations inside
		return uow.OrderRepo().Update(ctx, processed)
	})
}

// ❌ Bad - long transaction blocks other requests
func (s *Service) ProcessOrder(ctx context.Context) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		order, _ := uow.OrderRepo().OrderByIDLocked(ctx, id)
		
		// Heavy computation INSIDE transaction - BAD!
		time.Sleep(5 * time.Second)
		processed := s.heavyProcessing(order)
		
		return uow.OrderRepo().Update(ctx, processed)
	})
}
```

### 6. Читай без транзакций где возможно
Если операция только читает данные, транзакция не нужна:

```go
// ✅ Good - simple read without transaction
func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
	return s.uow.UserRepo().UserByID(ctx, id)
}

// ❌ Bad - unnecessary transaction overhead
func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
	var user User
	s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		user, _ = uow.UserRepo().UserByID(ctx, id)
		return nil
	})
	return user, nil
}
```

### 7. Моки в юнит-тестах, реальная БД в интеграционных

```go
// Unit test - fast, isolated
func TestService_UpdateUser(t *testing.T) {
	mockUow := mock.NewMockUnitOfWork()
	mockRepo := mock.NewMockUserRepository()
	// ...
}

// Integration test - slow, but tests real DB behavior
//go:build integration
func TestService_UpdateUser_Integration(t *testing.T) {
	db := setupTestDB(t)
	uow := repo.NewUnitOfWork(db)
	// ...
}
```

## Альтернативные подходы

### Подход 1: Передача *gorm.DB в репозитории

```go
// ❌ GORM протекает в сервис
func (s *Service) Update(ctx context.Context) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	
	if err := s.userRepo.Update(tx, user); err != nil {
		return err
	}
	
	return tx.Commit().Error
}
```

**Проблемы:**
- Сервис зависит от `*gorm.DB`
- Сложно тестировать
- Дублирование логики управления транзакциями
- Легко забыть `Rollback` или `Commit`

### Подход 2: Context-based транзакции

```go
type txKey struct{}

func InjectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func (r *repository) UserByID(ctx context.Context, id ID) (User, error) {
	db := ExtractTx(ctx, r.db) // Магия!
	// ...
}
```

**Проблемы:**
- Неявная зависимость через context
- Сложнее отлаживать (где началась транзакция?)
- Легко забыть извлечь tx из context
- `*Locked` методы сложнее реализовать

### Подход 3: WithTx() методы в репозиториях

```go
type Repository interface {
	WithTx(tx *gorm.DB) Repository
	UserByID(ctx context.Context, id ID) (User, error)
}

func (s *Service) Update(ctx context.Context) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		userRepo := s.userRepo.WithTx(tx)
		// ...
	})
}
```

**Проблемы:**
- Сервис всё равно зависит от `*gorm.DB`
- Нужно помнить вызывать `WithTx()` для каждого репозитория
- Легко ошибиться и использовать старый репозиторий без tx

### Подход 4: Unit of Work (текущий) ✅

```go
func (s *Service) Update(ctx context.Context) error {
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		user, err := uow.UserRepo().UserByIDLocked(ctx, id)
		// ... business logic
		return uow.UserRepo().Update(ctx, user)
	})
}
```

**Преимущества:**
- Полная абстракция от БД
- Явное управление транзакциями через `Do()`
- Автоматический rollback при ошибке
- Легко тестировать
- Невозможно забыть включить транзакцию для `*Locked` методов

## Ограничения и компромиссы

### 1. Boilerplate для новых репозиториев
При добавлении нового репозитория нужно:
- Добавить метод в интерфейс `UnitOfWork`
- Добавить поле в структуру `unitOfWork`
- Инициализировать в `NewUnitOfWork()` и в `Do()`

**Решение:** Это одноразовая операция, зато получаешь чистую архитектуру.

### 2. Сложные запросы
Для очень специфичных запросов (JOIN, подзапросы, агрегация) может потребоваться добавлять новые методы в репозиторий.

**Решение:** Это правильно! Специфичные для БД операции должны быть в репозитории, а не в сервисе.

### 3. Вложенные транзакции
GORM поддерживает savepoints, но через UoW это не так очевидно.

**Решение:** Используй GORM savepoints только если действительно нужны. Обычно достаточно одной транзакции на операцию.

### 4. Deadlocks
`*Locked` методы могут привести к deadlock при неправильном порядке блокировок.

**Решение:** 
- Всегда блокируй сущности в одном порядке (например, сначала User, потом Game)
- Держи транзакции короткими
- Используй database query timeout

```go
// ✅ Good - consistent lock order
func (s *Service) Transfer(ctx context.Context, fromID, toID string) error {
	// Sort IDs to ensure consistent lock order
	ids := []string{fromID, toID}
	sort.Strings(ids)
	
	return s.uow.Do(ctx, func(uow repo.UnitOfWork) error {
		user1, _ := uow.UserRepo().UserByIDLocked(ctx, ids[0])
		user2, _ := uow.UserRepo().UserByIDLocked(ctx, ids[1])
		// ...
	})
}
```

## Заключение

Unit of Work с методами `*Locked` - оптимальный баланс между чистотой архитектуры и удобством использования. Паттерн решает все основные проблемы работы с транзакциями в Go + GORM:

✅ Полная изоляция от деталей БД  
✅ Явное управление транзакциями  
✅ Безопасные конкурентные операции  
✅ Легко тестировать  
✅ SOLID принципы  

При этом остаётся простым и понятным для команды.
