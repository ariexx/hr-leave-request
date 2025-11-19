package config

import "github.com/spf13/viper"

type AppConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"` // in hours
}

type ApplicationConfig struct {
	AppConfig AppConfig      `mapstructure:"app"`
	Database  DatabaseConfig `mapstructure:"database"`
	JWT       JWTConfig      `mapstructure:"jwt"`
}

func LoadConfig() (*ApplicationConfig, error) {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	// Bind env so the docker can override config values
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.name", "DATABASE_NAME")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config ApplicationConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
