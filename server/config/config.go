package config

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerConfig           ServerConfig           `mapstructure:"server"`
	TokenAuthConfig        TokenAuthConfig        `mapstructure:"tokenAuth"`
	MongoDatabaseConfig    MongoDatabaseConfig    `mapstructure:"mongoDB"`
	PostgresDatabaseConfig PostgresDatabaseConfig `mapstructure:"postgres"`
	RedisConfig            RedisConfig            `mapstructure:"redis"`
	FirebaseConfig         FirebaseConfig         `mapstructure:"firebase"`
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
	JWTSecret    string `mapstructure:"jwtSecret"`
	JWTExpiresAt string `mapstructure:"jwtExpiresAt"`
	JWTIssuer    string `mapstructure:"jwtIssuer"`
}

type MongoDatabaseConfig struct {
	URI            string        `mapstructure:"uri"`
	Database       string        `mapstructure:"database"`
	ConnectTimeout time.Duration `mapstructure:"connectTimeout"`
}

type PostgresDatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"sslMode"`
	MaxConnections  int32         `mapstructure:"maxConnections"`
	MinConnections  int32         `mapstructure:"minConnections"`
	ConnectTimeout  time.Duration `mapstructure:"connectTimeout"`
	HealthCheckTime time.Duration `mapstructure:"healthCheckTime"`
}

func (pdc *PostgresDatabaseConfig) ConnectionDSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pdc.Host,
		pdc.Port,
		pdc.User,
		pdc.Password,
		pdc.Database,
		pdc.SSLMode,
	)
	return dsn
}

type RedisConfig struct {
	Address        string        `mapstructure:"address"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	Database       int           `mapstructure:"database"`
	TLS            bool          `mapstructure:"tls"`
	ConnectTimeout time.Duration `mapstructure:"connectTimeout"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
}

type FirebaseConfig struct {
	Type                    string `mapstructure:"type" json:"type"`
	ProjectID               string `mapstructure:"project_id" json:"project_id"`
	PrivateKeyID            string `mapstructure:"private_key_id" json:"private_key_id"`
	PrivateKey              string `mapstructure:"private_key" json:"private_key"`
	ClientEmail             string `mapstructure:"client_email" json:"client_email"`
	ClientID                string `mapstructure:"client_id" json:"client_id"`
	AuthURI                 string `mapstructure:"auth_uri" json:"auth_uri"`
	TokenURI                string `mapstructure:"token_uri" json:"token_uri"`
	AuthProviderX509CertURL string `mapstructure:"auth_provider_x509_cert_url" json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `mapstructure:"client_x509_cert_url" json:"client_x509_cert_url"`
	UniverseDomain          string `mapstructure:"universe_domain" json:"universe_domain"`
	StorageBucket           string `mapstructure:"storage_bucket" json:"-"`
}

func (fc *FirebaseConfig) ReturnConfigJSON() []byte {
	configJSON, err := json.Marshal(fc)
	if err != nil {
		log.Fatalf("failed to marshal firebase config: %v", err)
	}

	return configJSON
}

type MiddleWareConfig struct {
}

type AdditionalConfig struct {
	WorkoutPaginationLimit int      `mapstructure:"workoutPaginationLimit"`
	ExerciseCategories     []string `mapstructure:"exerciseCategories"`
	ExerciseMuscles        []string `mapstructure:"exerciseMuscles"`
	ExerciseEquipments     []string `mapstructure:"exerciseEquipments"`
	ExerciseDifficulty     []string `mapstructure:"exerciseDifficulty"`
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
