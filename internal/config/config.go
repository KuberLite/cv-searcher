package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers   []string
	KafkaTopic     string
	KafkaGroup     string
	MeiliSearchURL string
	HttpPort       string
}

func Load() Config {
	godotenv.Load()
	return Config{
		KafkaBrokers:   getKafkaBrokers(),
		KafkaTopic:     getKafkaTopic(),
		KafkaGroup:     getKafkaGroup(),
		MeiliSearchURL: getMeiliSearchURL(),
		HttpPort:       getHttpPort(),
	}
}

func getKafkaBrokers() []string {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	addresses := strings.Split(brokersEnv, ",")
	if len(addresses) == 1 && addresses[0] == "" {
		addresses = []string{"localhost:9092"}
	}
	return addresses
}

func getKafkaTopic() string {
	if env := os.Getenv("KAFKA_TOPIC"); env != "" {
		return env
	}
	return "default-topic"
}

func getMeiliSearchURL() string {
	if env := os.Getenv("MEILI_SEARCH_URL"); env != "" {
		return env
	}
	return "localhost:7700"
}

func getHttpPort() string {
	if env := os.Getenv("HTTP_PORT"); env != "" {
		return ":" + env
	}
	return ":8080"
}

func getKafkaGroup() string {
	if env := os.Getenv("KAFKA_GROUP"); env != "" {
		return env
	}
	return "default-group"
}
