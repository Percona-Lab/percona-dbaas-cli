package main

import (
	"fmt"
	"os"
	"os/exec"
)

type TestEngine interface {
	Run() error
	CreateDB(rootPass string) (o string, err error)
	ListDB() error
	DescribeDB() error
	ModifyDB() error
	DeleteDB(bool) error
}

func main() {
	for _, testEngine := range getTestEngines() {
		err := testEngine.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	fmt.Println("Tests done")
}

func getTestEngines() []TestEngine {
	cmd := "../cmd/percona-dbaas"
	k8sPXC := KuberPXC{
		cmd:    cmd,
		subCmd: "mysql",
		dbName: "test",
	}

	k8sPSMDB := KuberPSMDB{
		cmd:    cmd,
		subCmd: "mongodb",
		dbName: "test",
	}

	return []TestEngine{&k8sPXC, &k8sPSMDB}
}

func runCmd(cmd string, args ...string) (string, error) {
	cli := exec.Command(cmd, args...)
	o, err := cli.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(o), nil
}
