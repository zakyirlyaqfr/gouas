package config

type AppConfig struct {
	Port string
	Env  string
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Port: GetEnv("APP_PORT", "3000"),
		Env:  GetEnv("APP_ENV", "development"),
	}
}