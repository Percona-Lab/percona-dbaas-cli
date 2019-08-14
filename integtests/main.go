package main

import (
	"log"

	"github.com/Percona-Lab/percona-dbaas-cli/integtests/apitester"
)

func main() {
	tester := apitester.New("http://localhost:8081")
	err := tester.Run()
	log.Println(err)
}
