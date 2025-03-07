package main

import (
	"github.com/vysogota0399/gophermart_billing/internal/config"
	"github.com/vysogota0399/gophermart_billing/internal/storage"
)

func main() {
	if err := storage.RunMigration(config.MustNewConfig()); err != nil {
		panic(err)
	}
}
