package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

)

type ServerCfg struct{
	Host		string	`json:"host"`
	StorageCfg	StorageCfg
}

type StorageCfg struct{
	FileStoragePath		string
	StoreInterval		uint64	`json:"store_interval"`
	Restore				bool	`json:"restore"`
}

type AgentCfg struct{
	Host			string	`json:"host"`
	PollInterval 	uint64	`json:"poll_interval"`
	ReportInterval	uint64	`json:"report_interval"`
}

func NewServerConfig() (ServerCfg, error) {
	var cfg ServerCfg
	var storageCfg StorageCfg

	const defaultRunAddr = "localhost:8080"
	const defaultStoreInterval uint64 = 300
	const defaultRestore = true
	const defaultFileStoragePath = "/tmp/metrics-db.json"

	var(
		flagRunAddr  		string
		flagLogLevel 		string
		flagStoreInterval 	uint64
		flagFileStoragePath string
		flagRestore 		bool
	)

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")

	flag.Uint64Var(&flagStoreInterval, "i", defaultStoreInterval, "store interval")
	flag.StringVar(&flagFileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore")
	
	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok{
		cfg.Host = envRunAddr
	}

	storageCfg.StoreInterval = flagStoreInterval
	envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
	if ok{
		tmpStoreInterval, err := strconv.ParseUint(envStoreInterval, 10, 64)
		if err != nil{
			return cfg, fmt.Errorf("failed to parse %d as a report interval value: %w", tmpStoreInterval, err)
		}
		storageCfg.StoreInterval = tmpStoreInterval
	}

	storageCfg.FileStoragePath = flagFileStoragePath
	envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok{
		storageCfg.FileStoragePath = envFileStoragePath
	}

	storageCfg.Restore = flagRestore
	envRestore, ok := os.LookupEnv("RESTORE")
	if ok{
		boolValue, err := strconv.ParseBool(envRestore)
		if err != nil{
			return cfg, fmt.Errorf("failed to parse %v as a bool value: %w", boolValue, err)
		}
		storageCfg.Restore = boolValue
	}
	cfg.StorageCfg = storageCfg

	return cfg, nil
}

func NewAgentConfiguration() (AgentCfg, error) {
	var cfg AgentCfg

	const defaultRunAddr = "localhost:8080"
	const defaultPollInterval uint64 = 2
	const defaultReportInterval uint64 = 10

	var flagRunAddr string
	var flagReportInterval uint64
	var flagPollInterval uint64

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Uint64Var(&flagPollInterval, "p", defaultPollInterval, "data poll interval")
	flag.Uint64Var(&flagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}

	cfg.ReportInterval = flagReportInterval
	envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		ReportInterval, err := strconv.ParseUint(envReportInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %d as a report interval value: %w", ReportInterval, err)
		}
		cfg.ReportInterval = ReportInterval
	}


	cfg.PollInterval = flagPollInterval
	envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		PollInterval, err := strconv.ParseUint(envPollInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %d as a poll interval value: %w", PollInterval, err)
		}
		cfg.PollInterval = PollInterval
	}

	return cfg, nil
}
