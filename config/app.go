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

type ApplicationConfig struct {
	AppConfig AppConfig      `mapstructure:"app"`
	Database  DatabaseConfig `mapstructure:"database"`
}

func LoadConfig() (*ApplicationConfig, error) {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config ApplicationConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
