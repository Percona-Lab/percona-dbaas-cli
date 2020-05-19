package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type KuberPSMDB struct {
	cmd    string
	subCmd string
	dbName string
}

func (psmdb *KuberPSMDB) Run() error {
	fmt.Println("Run k8s-psmdb test")
	rootPass := "clisecretpass"
	err := psmdb.CreateDBWithPass(rootPass)
	if err != nil {
		return errors.Wrap(err, "create-db")
	}
	err = psmdb.CheckDBClusterReady()
	if err != nil {
		return errors.Wrap(err, "check db ready")
	}
	err = psmdb.ListDB()
	if err != nil {
		return errors.Wrap(err, "list")
	}
	err = psmdb.DescribeDB()
	if err != nil {
		return errors.Wrap(err, "describe-db")
	}
	err = psmdb.ModifyDB()
	if err != nil {
		return errors.Wrap(err, "modify-db")
	}
	err = psmdb.CheckDBClusterReady()
	if err != nil {
		return errors.Wrap(err, "check db ready")
	}
	err = psmdb.DescribeDB()
	if err != nil {
		return errors.Wrap(err, "describe-db")
	}
	err = psmdb.DeleteDB(true)
	if err != nil {
		return errors.Wrap(err, "delete-db")
	}
	err = psmdb.CheckPVCExist()
	if err != nil {
		return errors.Wrap(err, "check pvc")
	}
	err = psmdb.CreateDBAfterPreserve(rootPass)
	if err != nil {
		return errors.Wrap(err, "create-db after preserve")
	}
	err = psmdb.CheckDBClusterReady()
	if err != nil {
		return errors.Wrap(err, "check db ready")
	}
	err = psmdb.DeleteDB(false)
	if err != nil {
		return errors.Wrap(err, "delete-db")
	}
	err = psmdb.CreateDBWithNoWait()
	if err != nil {
		return errors.Wrap(err, "create-db after preserve")
	}
	err = psmdb.DeleteDB(false)
	if err != nil {
		return errors.Wrap(err, "delete-db")
	}
	err = psmdb.CreateDBWithPass(rootPass)
	if err != nil {
		return errors.Wrap(err, "create-db")
	}
	err = psmdb.DeleteDBWithoutOperator()
	if err != nil {
		return errors.Wrap(err, "delete-db without operator")
	}

	return nil
}

func (psmdb *KuberPSMDB) CreateDBWithPass(rootPass string) error {
	fmt.Println("Run create-db " + psmdb.dbName + " with pass")
	o, err := psmdb.CreateDB(rootPass)
	if err != nil {
		return errors.Wrap(err, "create")
	}
	fmt.Println(o)
	if !strings.Contains(o, rootPass) && !strings.Contains(o, "ready") {
		return errors.New("database starting error")
	}

	return nil
}

func (psmdb *KuberPSMDB) CreateDBAfterPreserve(rootPass string) error {
	fmt.Println("Run create-db " + psmdb.dbName + " after preserve")
	o, err := psmdb.CreateDB("")
	if err != nil {
		return errors.Wrap(err, "create")
	}
	if !strings.Contains(o, rootPass) && !strings.Contains(o, "ready") {
		return errors.New("data not preserve")
	}

	fmt.Println(o)

	var data struct {
		DB DB `json:"database"`
	}
	err = json.Unmarshal([]byte(o), &data)
	if err != nil {
		return errors.Wrap(err, "unmarshal json out")
	}
	return nil
}

func (psmdb *KuberPSMDB) CreateDBWithNoWait() error {
	fmt.Println("Run create-db " + psmdb.dbName + " with no-wait")
	o, err := runCmd(psmdb.cmd, psmdb.subCmd, "create-db", psmdb.dbName, "--no-wait")
	if err != nil {
		return errors.Wrap(err, "create")
	}
	if !strings.Contains(o, psmdb.dbName) && !strings.Contains(o, "initializing") {
		return errors.New("wrong output")
	}

	return nil
}

func (psmdb *KuberPSMDB) CreateDB(rootPass string) (o string, err error) {
	if len(rootPass) > 0 {
		return runCmd(psmdb.cmd, psmdb.subCmd, "create-db", psmdb.dbName, "--password", rootPass)
	}
	return runCmd(psmdb.cmd, psmdb.subCmd, "create-db", psmdb.dbName, "-o", "json")
}

func (psmdb *KuberPSMDB) CheckDBClusterReady() error {
	var cluster k8sStatus
	o, err := GetK8SObject("psmdb", psmdb.dbName)
	if err != nil {
		return errors.Wrap(err, "get k8s object psmdb/"+psmdb.dbName)
	}
	err = json.Unmarshal(o, &cluster)
	if err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	if cluster.Status.Status != "ready" {
		return errors.New("cluster not ready")
	}
	return nil
}

func (psmdb *KuberPSMDB) CheckPVCExist() error {
	o, err := GetK8SObject("pvc", "mongod-data-"+psmdb.dbName+"-rs0-0")
	if err != nil {
		return errors.Wrapf(err, "get k8s object pvc/mongod-data-%s-rs0-0", psmdb.dbName)
	}

	if strings.Contains(string(o), "NotFound") {
		return errors.New("pvc not exist")
	}
	return nil
}

func (psmdb *KuberPSMDB) ListDB() error {
	fmt.Println("Run describe-db")
	o, err := runCmd(psmdb.cmd, psmdb.subCmd, "describe-db")
	if err != nil {
		return errors.Wrap(err, "run describe-db cmd")
	}
	fmt.Println(o)
	if !strings.Contains(o, psmdb.dbName) {
		return errors.Errorf("list db not work. Output: %s", o)
	}

	fmt.Println("Run describe-db in JSON")
	o, err = runCmd(psmdb.cmd, psmdb.subCmd, "describe-db", "-o", "json")
	if err != nil {
		return errors.Wrap(err, "run describe-db cmd with json out")
	}
	fmt.Println(o)
	var data struct {
		List []DB `json:"database-list"`
	}
	err = json.Unmarshal([]byte(o), &data)
	if err != nil {
		return errors.Wrap(err, "unmarshal json out")
	}
	return nil
}

func (psmdb *KuberPSMDB) DescribeDB() error {
	fmt.Println("Run describe-db " + psmdb.dbName)
	o, err := runCmd(psmdb.cmd, psmdb.subCmd, "describe-db", psmdb.dbName)
	if err != nil {
		return errors.Wrapf(err, "run describe-db for %s cmd", psmdb.dbName)
	}
	fmt.Println(o)
	if !strings.Contains(o, "ready") && !strings.Contains(o, psmdb.dbName) {
		return errors.Errorf("db not start correctly. Output: %s", o)
	}

	fmt.Println("Run describe-db " + psmdb.dbName + " in json")
	o, err = runCmd(psmdb.cmd, psmdb.subCmd, "describe-db", psmdb.dbName, "-o", "json")
	if err != nil {
		return errors.Wrapf(err, "run describe-db for %s cmd", psmdb.dbName)
	}
	fmt.Println(o)
	if !strings.Contains(o, "ready") && !strings.Contains(o, psmdb.dbName) {
		return errors.Errorf("db not start correctly. Output: %s", o)
	}
	return nil
}

func (psmdb *KuberPSMDB) ModifyDB() error {
	fmt.Println("Run modify-db " + psmdb.dbName)
	o, err := runCmd(psmdb.cmd, psmdb.subCmd, "modify-db", psmdb.dbName, "--options", "pxc.replesets[rs-0].requests.memory=1G")
	if err != nil {
		return errors.Wrap(err, "run modify-db cmd")
	}
	if !strings.Contains(o, "ready") {
		return errors.Errorf("db not modified correctly. Output: %s", o)
	}
	return nil
}

func (psmdb *KuberPSMDB) DeleteDB(preserve bool) error {
	fmt.Println("Run delete-db "+psmdb.dbName+". Preserve flag is", preserve)
	preserveFlag := ""
	if preserve {
		preserveFlag = "--preserve-data"
	}
	o, err := runCmd(psmdb.cmd, psmdb.subCmd, "delete-db", psmdb.dbName, "-y", preserveFlag)
	if err != nil {
		return errors.Wrap(err, "run delete-db cmd")
	}
	if !strings.Contains(o, "done") {
		return errors.Errorf("db not deleted correctly. Output: %s", o)
	}
	return nil
}

func (psmdb *KuberPSMDB) DeleteDBWithoutOperator() error {
	fmt.Println("Run delete-db " + psmdb.dbName + ". Without Operator")
	_, err := DeleteDeployment("percona-server-mongodb-operator")
	if err != nil {
		return errors.Wrap(err, "delete operator deployment")
	}

	o, err := runCmd(psmdb.cmd, psmdb.subCmd, "delete-db", psmdb.dbName, "-y")
	if err != nil {
		return errors.Wrap(err, "run delete-db cmd")
	}
	fmt.Println(o)
	if strings.Contains(o, "done") {
		return errors.Errorf("db deleted but should not. Output: %s", o)
	}
	return nil
}
