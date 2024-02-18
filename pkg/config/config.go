// Package config provides a configuration setup needed to run the app.
package config

import (
	"time"

	"github.com/spf13/viper"
)

type DB struct {
	User         string `mapstructure:"DB_USER"`
	Password     string `mapstructure:"DB_PASSWORD"`
	Host         string `mapstructure:"DB_HOST"`
	Name         string `mapstructure:"DB_NAME"`
	MaxIdleConns int    `mapstructure:"MAX_IDLE_CONNS"`
	MaxOpenConns int    `mapstructure:"MAX_OPEN_CONNS"`
	DisableTLS   bool   `mapstructure:"DISABLE_TLS"`
}

type Auth struct {
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

type Mailer struct {
	MailerName    string `mapstructure:"MAILER_NAME"`
	EmailAddress  string `mapstructure:"MAILER_EMAIL_ADDRESS"`
	EmailPassword string `mapstructure:"MAILER_EMAIL_PASSWORD"`
}

type Config struct {
	DB                `mapstructure:",squash"`
	Auth              `mapstructure:",squash"`
	Mailer            `mapstructure:",squash"`
	RedisHost         string `mapstructure:"REDIS_HOST"`
	MigrationURL      string `mapstructure:"MIGRATION_URL"`
	DatabaseURL       string `mapstructure:"DATABASE_URL"`
	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
}

func Load(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	var cfg Config

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}
		return cfg, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
