package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerConfig           ServerConfig           `mapstructure:"server"`
	TokenAuthConfig        TokenAuthConfig        `mapstructure:"tokenAuth"`
	MongoDatabaseConfig    MongoDatabaseConfig    `mapstructure:"mongoDatabase"`
	PostgresDatabaseConfig PostgresDatabaseConfig `mapstructure:"postgresDatabase"`
	RedisConfig            RedisConfig            `mapstructure:"redis"`
	MiddleWareConfig       MiddleWareConfig       `mapstructure:"middleWare"`
	AdditionalConfig       AdditionalConfig       `mapstructure:"additional"`
}

type ServerConfig struct {
	ListenAddress      string        `mapstructure:"listenAddress"`
	Port               string        `mapstructure:"port"`
	WorkerPort         string        `mapstructure:"workerPort"`
	ReadTimeout        time.Duration `mapstructure:"readTimeout"`
	WriteTimeout       time.Duration `mapstructure:"writeTimeout"`
	Environment        string        `mapstructure:"environment"`
	CORSAllowedOrigins []string      `mapstructure:"corsAllowedOrigins"`
	CORSAllowedMethods []string      `mapstructure:"corsAllowedMethods"`
	CORSAllowedHeaders []string      `mapstructure:"corsAllowedHeaders"`
}

type TokenAuthConfig struct {
	JWTSignKey   string `mapstructure:"jwtSignKey"`
	JWTExpiresAt string `mapstructure:"jwtExpiresAt"`
}

type MongoDatabaseConfig struct {
}

type PostgresDatabaseConfig struct {
}

type RedisConfig struct {
}

type MiddleWareConfig struct {
}

type AdditionalConfig struct {
}

func GetConfig() *Config {
	var cfg Config

	v := viper.New()
	v.SetConfigName("default")
	v.AddConfigPath("../conf")
	v.AddConfigPath("../../conf")
	v.AddConfigPath(".")
	v.AddConfigPath("./conf/")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
