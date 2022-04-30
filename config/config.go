package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Topic         string   `env:"TOPIC" envDefault:"realtime.go"`
	InternalAddrs []string `env:"INTERNAL_ADDRS" envSeparator:","`
	KafkaAddrs    []string `env:"KAFKA_ADDRS" envSeparator:","`
	PulsarAddrs   []string `env:"PULSAR_ADDRS" envSeparator:","`
	RedisAddrs    []string `env:"REDIS_ADDRS" envSeparator:","`
}

var config Config

func Parse() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	if err := env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func Get() Config {
	return config
}
