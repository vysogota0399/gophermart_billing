package transaction_outbox

import (
	"github.com/caarlos0/env"
)

type Config struct {
	OrderCreatedPollInterval       int    `json:"order_created_poll_inverval" env:"DAEMON_ORDER_CREATED_EVENT_POLL_INTERVAL" envDefault:"250"`
	OrderUpdatedPollInterval       int    `json:"order_updated_poll_inverval" env:"DAEMON_ORDER_UPDATED_EVENT_POLL_INTERVAL" envDefault:"250"`
	AccountingCreatedPollInterval  int    `json:"accounting_created_poll_inverval" env:"DAEMON_ACCOUNTING_CREATED_EVENT_POLL_INTERVAL" envDefault:"250"`
	WorkersCount                   int64  `json:"workersCount" env:"DAEMON_WORKERS_COUNT" envDefault:"1"`
	KafkaOrderCreatedTopic         string `json:"kafka_order_created_topic" env:"KAFKA_ORDER_CREATED_TOPIC" envDefault:"order_created"`
	KafkaOrderUpdatedTopic         string `json:"kafka_order_updated_topic" env:"KAFKA_ORDER_UPDATED_TOPIC" envDefault:"order_updated"`
	KafkaTransactionProcessedTopic string `json:"kafka_transaction_processed_topic" env:"KAFKA_TRANSACTION_PROCESSED_TOPIC" envDefault:"transaction_processed"`
}

func MustNewConfig() *Config {
	c := &Config{}
	env.Parse(c)

	return c
}
