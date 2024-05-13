package main

import (
	"flag"
	"os"
)

var flagRunAddr string

func parseFlagsServer() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")

	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address != ""{
		flagRunAddr = address
	}
}
