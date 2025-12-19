package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type TelegramToken string

type Config struct {
	Postgres PostgresConfig `env-prefix:"POSTGRES__"`
	Sqlite   SqliteConfig   `env-prefix:"SQLITE__"`
	Nats     NatsConfig     `env-prefix:"NATS__"`
	App      AppConfig      `env-prefix:"APP__"`
	Logs     LogsConfig     `env-prefix:"LOGS__"`
	Telegram TelegramConfig `env-prefix:"TELEGRAM__"`
}

type LogsConfig struct {
	LogLevel    string `env:"LEVEL"        env-default:"info"  validate:"oneof=debug info warn error"`
	IsPretty    bool   `env:"IS_PRETTY"    env-default:"true"`
	WithContext bool   `env:"WITH_CONTEXT" env-default:"true"`
	WithSources bool   `env:"WITH_SOURCES" env-default:"false"`
}

type SqliteConfig struct {
	URL string `env:"URL" env-default:"file::memory:?cache=shared" validate:"required"`
}

type PostgresConfig struct {
	URL string `env:"URL" env-default:"postgres://app:app@127.0.0.1:5432/app" validate:"required"`
}

type TelegramConfig struct {
	Token    TelegramToken `env:"TOKEN"     validate:"required"`
	AdminIDs []int64       `env:"ADMIN_IDS" validate:""`
	Debug    bool          `env:"DEBUG"                         env-default:"false"`
}

type AppConfig struct {
	GormDialector      string `env:"GORM_DIALECTOR"        env-default:"sqlite" validate:"oneof=sqlite postgres"`
	MaxRequestsPerUser int    `env:"MAX_REQUESTS_PER_USER" env-default:"3"      validate:"min=1"`
	MaxGameCount       int    `env:"MAX_GAME_COUNT"        env-default:"10"     validate:"min=1"`
}

type NatsConfig struct {
	URL               string `env:"URL"                env-default:"nats://nats:4222" validate:"required"`
	NKeySeed          string `env:"NKEY_SEED"                                         validate:"required"`
	NKeyPublic        string `env:"NKEY_PUBLIC"                                       validate:"required"`
	MetastoreReplicas int    `env:"METASTORE_REPLICAS" env-default:"1"                validate:"min=1,max=5"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("failed to load .env file: %w", err)
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
