package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/integtests/apitester"
)

var (
	address string
	wait    int
)

func init() {
	flag.StringVar(&address, "addr", "http://localhost:8081", "Server address")
	flag.IntVar(&wait, "wait", 0, "Wait before start")
}

func main() {
	flag.Parse()
	if len(address) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if wait > 0 {
		time.Sleep(time.Duration(wait) * time.Second)
	}
	tester := apitester.New(address)
	err := tester.Run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
