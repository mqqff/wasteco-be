package env

import (
	"github.com/mqqff/wasteco-be/pkg/log"
	"github.com/spf13/viper"
)

type Env struct {
	AppName string `mapstructure:"APP_NAME"`
	AppHost string `mapstructure:"APP_HOST"`
	AppPort string `mapstructure:"APP_PORT"`
	AppEnv  string `mapstructure:"APP_ENV"`

	DBName string `mapstructure:"DB_NAME"`
	DBHost string `mapstructure:"DB_HOST"`
	DBPort string `mapstructure:"DB_PORT"`
	DBUser string `mapstructure:"DB_USER"`
	DBPass string `mapstructure:"DB_PASS"`
}

var AppEnv = getEnv()

func getEnv() *Env {
	env := &Env{}
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(log.LogInfo{"error": err.Error()}, "[Env][getEnv] failed to read env file")
	}

	if err := viper.Unmarshal(env); err != nil {
		log.Fatal(log.LogInfo{"error": err.Error()}, "[Env][getEnv] failed to unmarshal")
	}

	return env
}
