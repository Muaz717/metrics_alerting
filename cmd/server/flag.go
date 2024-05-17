package main

import (
	"flag"
	"os"
)

var flagRunAddr  string
var flagLogLevel string

func parseFlagsServer() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")

	flag.Parse()

	envAddress := os.Getenv("ADDRESS")
	if envAddress != ""{
		flagRunAddr = envAddress
	}

	envlogLevel := os.Getenv("LOG_LEVEL")
	if envlogLevel != ""{
		flagLogLevel = envlogLevel
	}
}
