package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	Database   string
	Accural    string
}

var ReadyConfig Config

func ParseConfig() {
	flag.StringVar(&ReadyConfig.ServerAddr, "a", "localhost:8080", "address port to run server")
	flag.StringVar(&ReadyConfig.Database, "d", "postgres://postgres:Xer_0101@localhost/gophermart?sslmode=disable", "database URI")
	flag.StringVar(&ReadyConfig.Accural, "r", "", "accural system address")
	flag.Parse()
	if serverAddr := os.Getenv("RUN_ADDRESS"); serverAddr != "" {
		ReadyConfig.ServerAddr = serverAddr
	}
	if databaseURI := os.Getenv("DATABASE_URI"); databaseURI != "" {
		ReadyConfig.Database = databaseURI
	}
	if accuralAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accuralAddress != "" {
		ReadyConfig.Accural = accuralAddress
	}
}
