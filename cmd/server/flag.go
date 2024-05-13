package main

import "flag"

var flagRunAddr string

func parseFlagsServer() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")

	flag.Parse()
}
