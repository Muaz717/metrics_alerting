package main

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

var flags struct{
	flagRunAddr 		string
	flagReportInterval 	int
	flagPollInterval 	int
	flagLogLevel		string
}

type Config struct{
	Address 		string	`env:"ADDRESS"`
	ReportInterval 	int		`env:"REPORT_INTERVAL"`
	PollInterval	int		`env:"POLL_INTERVAL"`
	LogLevel		string	`env:"LOG_LEVEL"`
}

func parseFlagsAgent()error {
	flag.StringVar(&flags.flagRunAddr, "a", "localhost:8080", "port to send requests")
	flag.IntVar(&flags.flagReportInterval, "r", 10, "set rerpot interval")
	flag.IntVar(&flags.flagPollInterval, "p", 2, "set poll interval")
	flag.StringVar(&flags.flagLogLevel, "l", "info", "log level")

	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil{
		return err
	}

	if cfg.Address != ""{
		flags.flagRunAddr = cfg.Address
	}

	if cfg.ReportInterval > 0{
		flags.flagReportInterval = cfg.ReportInterval
	}

	if cfg.PollInterval > 0{
		flags.flagPollInterval = cfg.PollInterval
	}

	if cfg.LogLevel != ""{
		flags.flagLogLevel = cfg.LogLevel
	}
	return nil
}
