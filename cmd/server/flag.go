package main

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

var(
	flagRunAddr  		string
 	flagLogLevel 		string
	flagStoreInterval 	int
	flagFileStoragePath string
	flagRestore 		bool
)

type Config struct{
	Address  		string	`env:"ADDRESS"`
	LogLevel 		string	`env:"LOG_LEVEL"`
	StoreInterval 	int 	`env:"STORE_INTERVAL"`
	FileStoragePath	string	`env:"FILE_STORAGE_PATH"`
	Restore 		bool	`env:"RESTORE"`
}

func parseFlagsServer() error{
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.IntVar(&flagStoreInterval, "i", 30, "store interval")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore")

	flag.Parse()

	var cfg Config

	err := env.Parse(&cfg)
	if err != nil{
		return err
	}

	if cfg.Address != ""{
		flagRunAddr = cfg.Address
	}

	if cfg.LogLevel != ""{
		flagLogLevel = cfg.LogLevel
	}

	if cfg.StoreInterval >= 0{
		flagStoreInterval = cfg.StoreInterval
	}

	if cfg.FileStoragePath != ""{
		flagFileStoragePath = cfg.FileStoragePath
	}

	if !cfg.Restore{
		flagRestore = cfg.Restore
	}

	return nil
}
