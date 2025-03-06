package transaction_inbox

import (
	"github.com/caarlos0/env"
)

type Config struct {
	KafkaAccrualsPartitionWatchInterval int    `json:"kafka_accruals_partition_watch_interval" env:"KAFKA_ACCRUAL_PARTITION_WATCH_INTERVAL" envDefault:"5000"`
	KafkaAccrualsTopic                  string `json:"kafka_accruals_topic" env:"KAFKA_ACCRUALs_TOPIC" envDefault:"accruals"`
	KafkaAccrualsGroupID                string `json:"kafka_accruals_group_id" env:"KAFKA_ACCRUAL_GROUP_ID" envDefault:"billing_accruals_consumer_group"`
	KafkaAccrualsdMaxWaitInterval       int    `json:"kafka_accruals_max_wait_interval" env:"KAFKA_ACCRUALS_MAX_WAIT_INTERVAL" envDefault:"5000"`
}

func MustNewConfig() *Config {
	c := new(Config)
	env.Parse(c)

	return c
}
