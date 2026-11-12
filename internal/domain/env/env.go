package env

import (
	"time"

	"github.com/mqqff/wasteco-be/pkg/log"
	"github.com/spf13/viper"
)

type Env struct {
	AppName string `mapstructure:"APP_NAME"`
	AppHost string `mapstructure:"APP_HOST"`
	AppPort string `mapstructure:"APP_PORT"`
	AppEnv  string `mapstructure:"APP_ENV"`

	APIKey string `mapstructure:"API_KEY"`

	DBName string `mapstructure:"DB_NAME"`
	DBHost string `mapstructure:"DB_HOST"`
	DBPort string `mapstructure:"DB_PORT"`
	DBUser string `mapstructure:"DB_USER"`
	DBPass string `mapstructure:"DB_PASS"`

	JWTSecretKey   string        `mapstructure:"JWT_SECRET_KEY"`
	JWTExpiredTime time.Duration `mapstructure:"JWT_EXPIRED_TIME"`

	SupabaseURL    string `mapstructure:"SUPABASE_URL"`
	SupabaseSecret string `mapstructure:"SUPABASE_SECRET"`
	SupabaseBucket string `mapstructure:"SUPABASE_BUCKET"`

	GeminiAPIKey string `mapstructure:"GEMINI_API_KEY"`
	GeminiModel  string `mapstructure:"GEMINI_MODEL"`
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
