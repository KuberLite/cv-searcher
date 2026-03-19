package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers   []string
	KafkaTopic     string
	KafkaGroup     string
	MeiliSearchURL string
	HttpPort       string
	VectorizerURL  string
	QDRantURL      *QDRantConfig
}

type QDRantConfig struct {
	URL  string
	Port int
}

func Load() Config {
	godotenv.Load()
	return Config{
		KafkaBrokers:   getKafkaBrokers(),
		KafkaTopic:     getKafkaTopic(),
		KafkaGroup:     getKafkaGroup(),
		MeiliSearchURL: getMeiliSearchURL(),
		HttpPort:       getHttpPort(),
		VectorizerURL:  getVectorizerURL(),
		QDRantURL:      getQDRantURL(),
	}
}

func getQDRantURL() *QDRantConfig {
	env := os.Getenv("QDR_URL")
	if env == "" {
		return &QDRantConfig{URL: "localhost", Port: 6334}
	}
	parts := strings.SplitN(env, ":", 2)
	if len(parts) < 2 {
		return &QDRantConfig{URL: parts[0], Port: 6334}
	}
	port, _ := strconv.Atoi(parts[1])
	if port <= 0 {
		port = 6334
	}
	return &QDRantConfig{URL: parts[0], Port: port}
}

func getVectorizerURL() string {
	if env := os.Getenv("VECTORIZER_URL"); env != "" {
		return "http://" + env
	}
	return "http://127.0.0.1:8000"
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
		return "http://" + env
	}
	return "http://localhost:7700"
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
