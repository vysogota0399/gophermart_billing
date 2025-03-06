package config

import (
	"github.com/caarlos0/env"
)

type Config struct {
	OrdersServerHTTPAddress     string   `json:"http_address" env:"ORDERS_SERVER_HTTP_ADDRESS" envDefault:"127.0.0.1:8010"`
	AccountingServerHTTPAddress string   `json:"accounting_http_address" env:"ACCOUNTING_SERVER_HTTP_ADDRESS" envDefault:"127.0.0.1:8020"`
	LogLevel                    int      `json:"log_level" env:"LOG_LEVEL" envDefault:"0"`
	DatabaseDSN                 string   `json:"database_dsn" env:"DATABASE_DSN" envDefault:"postgres://postgres:secret@127.0.0.1:5432/gophermart_billing_development"`
	KafkaBrokers                []string `json:"kafka_brokers" env:"KAFKA_BROKERS" envDefault:"127.0.0.1:9000,127.0.0.1:9001,127.0.0.1:9002" envSeparator:","`
	KafkaLogLevel               int      `json:"kafka_log_level" env:"KAFKA_LOG_LEVEL" envDefault:"0"`
}

func MustNewConfig() *Config {
	c := new(Config)
	env.Parse(c)

	return c
}
