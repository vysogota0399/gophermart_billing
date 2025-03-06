package main

import (
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
)

func main() {
	storage.RunMigration(config.MustNewConfig())
}
