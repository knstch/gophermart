package config

import (
	"flag"
	"os"
)

// A config setup struct.
type Config struct {
	ServerAddr string
	Database   string
	Accural    string
	SecretKey  string
}

// A config variable.
var ReadyConfig Config

// A function that parses config from flags and environmental variables.
func ParseConfig() {
	flag.StringVar(&ReadyConfig.ServerAddr, "a", "localhost:8080", "address port to run server")
	flag.StringVar(&ReadyConfig.Database, "d", "postgres://postgres:Xer_0101@localhost/gophermart?sslmode=disable", "database URI")
	flag.StringVar(&ReadyConfig.Accural, "r", "http://localhost:8081", "accural system address")
	flag.StringVar(&ReadyConfig.SecretKey, "k", "aboba", "secret key to encode cookies")
	flag.Parse()
	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		ReadyConfig.SecretKey = secretKey
	}
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
