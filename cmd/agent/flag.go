package main

import "flag"

var flags struct{
	flagRunAddr string
	flagRepotInterval int
	flagPollInterval int
}

func parseFlagsAgent() {
	flag.StringVar(&flags.flagRunAddr, "a", "localhost:8080", "port to send requests")
	flag.IntVar(&flags.flagRepotInterval, "r", 10, "set rerpot interval")
	flag.IntVar(&flags.flagPollInterval, "p", 2, "set poll interval")

	flag.Parse()
}
