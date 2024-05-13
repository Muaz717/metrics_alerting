package main

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

var flags struct{
	flagRunAddr string
	flagReportInterval int
	flagPollInterval int
}

type Config struct{
	Address 		string	`env:"ADDRESS"`
	ReportInterval 	int		`env:"REPORT_INTERVAL"`
	PollInterval	int		`env:"POLL_INTERVAL"`
}

func parseFlagsAgent() {
	flag.StringVar(&flags.flagRunAddr, "a", "localhost:8080", "port to send requests")
	flag.IntVar(&flags.flagReportInterval, "r", 10, "set rerpot interval")
	flag.IntVar(&flags.flagPollInterval, "p", 2, "set poll interval")

	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil{
		panic(err)
	}

	if cfg.Address != ""{
		flags.flagRunAddr = cfg.Address
	}

	if cfg.ReportInterval >= 0{
		flags.flagReportInterval = cfg.ReportInterval
	}

	if cfg.PollInterval >= 0{
		flags.flagPollInterval = cfg.PollInterval
	}
}
