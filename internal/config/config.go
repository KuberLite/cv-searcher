package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers   []string
	KafkaTopic     string
	MeiliSearchURL string
	HttpPort       string
}

func Load() Config {
	return Config{
		KafkaBrokers:   getBrokers(),
		KafkaTopic:     getTopic(),
		MeiliSearchURL: getMeiliSearchURL(),
		HttpPort:       getHttpPort(),
	}
}

func getBrokers() []string {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	addresses := strings.Split(brokersEnv, ",")
	if len(addresses) == 1 && addresses[0] == "" {
		addresses = []string{"localhost:9092"}
	}
	return addresses
}

func getTopic() string {
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
