# Unit of Work Pattern with GORM

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
	userRepo "minigame-bot/internal/repo/user"
	tttRepo "minigame-bot/internal/repo/ttt"
)

// UnitOfWork provides transactional operations over multiple repositories
type UnitOfWork interface {
	// Do executes function within a transaction
	Do(ctx context.Context, fn func(uow UnitOfWork) error) error
	
	// Repository accessors
	UserRepo() userRepo.IUserRepository
	TTTRepo() tttRepo.ITTTRepository
	
	// Raw DB access for complex queries
	DB() *gorm.DB
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

func (u *unitOfWork) DB() *gorm.DB {
	return u.db
}
```

### 2. Обновление репозиториев

```go
// internal/repo/user/gorm_repository.go
package user

import (
	"context"
	"gorm.io/gorm"
	domainUser "minigame-bot/internal/domain/user"
	repository "minigame-bot/internal/repo/user"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) repository.IUserRepository {
	return &gormRepository{db: db}
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

### 3. Использование в сервисе

```go
// internal/service/game_service.go
package service

import (
	"context"
	"errors"
	"minigame-bot/internal/domain/ttt"
	"minigame-bot/internal/domain/user"
	"minigame-bot/internal/repo"
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
		// SELECT FOR UPDATE - lock row
		var gameModel ttt.GGameModel
		err := uow.DB().
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", gameID.String()).
			First(&gameModel).Error
		if err != nil {
			return err
		}
		
		game, err := gameModel.ToDomain()
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
	"minigame-bot/internal/repo"
	userRepo "minigame-bot/internal/repo/user"
	tttRepo "minigame-bot/internal/repo/ttt"
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

func (m *MockUnitOfWork) DB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}
```

### 2. Mock для репозиториев

```go
// internal/repo/user/mock/user_repository_mock.go
package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"minigame-bot/internal/domain/user"
	repository "minigame-bot/internal/repo/user"
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
	"minigame-bot/internal/domain/ttt"
	repository "minigame-bot/internal/repo/ttt"
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
	
	"minigame-bot/internal/domain/ttt"
	"minigame-bot/internal/domain/user"
	repoMock "minigame-bot/internal/repo/mock"
	userRepoMock "minigame-bot/internal/repo/user/mock"
	tttRepoMock "minigame-bot/internal/repo/ttt/mock"
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
	
	"minigame-bot/internal/domain/ttt"
	"minigame-bot/internal/domain/user"
	"minigame-bot/internal/repo"
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
```

## Преимущества паттерна

1. **Инкапсуляция транзакций** - логика управления транзакциями в одном месте
2. **Простота тестирования** - легко мокировать через интерфейс
3. **Consistency** - все операции в рамках одной транзакции или все откатываются
4. **Clean Architecture** - сервис не зависит от деталей реализации БД
5. **Переиспользование** - один UoW для всех сервисов

## Best Practices

1. **Всегда используй транзакции** для операций, затрагивающих несколько сущностей
2. **SELECT FOR UPDATE** для конкурентных операций (через `uow.DB()`)
3. **Короткие транзакции** - минимизируй время блокировок
4. **Не смешивай** чтение и запись в одном методе сервиса (если не нужна транзакция)
5. **Моки в юнит-тестах**, реальная БД в интеграционных
