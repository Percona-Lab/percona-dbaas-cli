package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type KuberPXC struct {
	cmd    string
	subCmd string
	dbName string
}

func (pxc *KuberPXC) Run() error {
	fmt.Println("Run k8s-pxc test")
	rootPass := "clisecretpass"
	err := pxc.CreateDBWithPass(rootPass)
	if err != nil {
		return errors.Wrap(err, "create-db")
	}
	err = pxc.CheckDBClusterReady()
	if err != nil {
		return errors.Wrap(err, "check db ready")
	}
	err = pxc.ListDB()
	if err != nil {
		return errors.Wrap(err, "list")
	}
	err = pxc.DescribeDB()
	if err != nil {
		return errors.Wrap(err, "describe-db")
	}
	err = pxc.ModifyDB()
	if err != nil {
		return errors.Wrap(err, "modify-db")
	}
	err = pxc.CheckDBClusterReady()
	if err != nil {
		return errors.Wrap(err, "check db ready")
	}
	err = pxc.DeleteDB(true)
	if err != nil {
		return errors.Wrap(err, "delete-db with preserve")
	}
	err = pxc.CheckPVCExist()
	if err != nil {
		return errors.Wrap(err, "check pvc")
	}
	err = pxc.CreateDBAfterPreserve(rootPass)
	if err != nil {
		return errors.Wrap(err, "create-db after preserve")
	}
	err = pxc.CheckDBClusterReady()
	if err != nil {
		return errors.Wrap(err, "check db ready")
	}
	err = pxc.DeleteDB(false)
	if err != nil {
		return errors.Wrap(err, "delete-db")
	}
	err = pxc.CreateDBWithNoWait()
	if err != nil {
		return errors.Wrap(err, "create-db after preserve")
	}
	err = pxc.DeleteDB(false)
	if err != nil {
		return errors.Wrap(err, "delete-db")
	}

	return nil
}

func (pxc *KuberPXC) CreateDBWithPass(rootPass string) error {
	fmt.Println("Run create-db " + pxc.dbName + " with pass")
	o, err := pxc.CreateDB(rootPass)
	if err != nil {
		return errors.Wrap(err, "create")
	}
	if !strings.Contains(o, rootPass) && !strings.Contains(o, "ready") {
		return errors.New("database starting error")
	}

	return nil
}

func (pxc *KuberPXC) CreateDBAfterPreserve(rootPass string) error {
	fmt.Println("Run create-db " + pxc.dbName + " after preserve")
	o, err := pxc.CreateDB("")
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

func (pxc *KuberPXC) CreateDBWithNoWait() error {
	fmt.Println("Run create-db " + pxc.dbName + " with no-wait")
	o, err := runCmd(pxc.cmd, pxc.subCmd, "create-db", pxc.dbName, "--no-wait")
	if err != nil {
		return errors.Wrap(err, "create")
	}
	if !strings.Contains(o, pxc.dbName) && !strings.Contains(o, "initializing") {
		return errors.New("wrong output")
	}

	return nil
}

func (pxc *KuberPXC) CreateDB(rootPass string) (o string, err error) {
	if len(rootPass) > 0 {
		return runCmd(pxc.cmd, pxc.subCmd, "create-db", pxc.dbName, "--password", rootPass)
	}
	return runCmd(pxc.cmd, pxc.subCmd, "create-db", pxc.dbName, "-o", "json")
}

func (pxc *KuberPXC) CheckDBClusterReady() error {
	var cluster k8sStatus
	o, err := GetK8SObject("pxc", pxc.dbName)
	if err != nil {
		return errors.Wrap(err, "get k8s object pxc/"+pxc.dbName)
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

func (pxc *KuberPXC) CheckPVCExist() error {
	o, err := GetK8SObject("pvc", "datadir-"+pxc.dbName+"-pxc-0")
	if err != nil {
		return errors.Wrap(err, "get k8s object pvc/datadir-"+pxc.dbName+"-pxc-0")
	}

	if strings.Contains(string(o), "NotFound") {
		return errors.New("pvc not exist")
	}
	return nil
}

func (pxc *KuberPXC) ListDB() error {
	fmt.Println("Run describe-db")
	o, err := runCmd(pxc.cmd, pxc.subCmd, "describe-db")
	if err != nil {
		return errors.Wrap(err, "run describe-db cmd")
	}
	fmt.Println(o)
	if !strings.Contains(o, pxc.dbName) {
		return errors.New("list db not work. Output: " + o)
	}

	fmt.Println("Run describe-db in JSON")
	o, err = runCmd(pxc.cmd, pxc.subCmd, "describe-db", "-o", "json")
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

func (pxc *KuberPXC) DescribeDB() error {
	fmt.Println("Run describe-db " + pxc.dbName)
	o, err := runCmd(pxc.cmd, pxc.subCmd, "describe-db", pxc.dbName)
	if err != nil {
		return errors.Wrap(err, "run describe-db for"+pxc.dbName+" cmd")
	}
	fmt.Println(o)
	if !strings.Contains(o, "ready") && !strings.Contains(o, pxc.dbName) {
		return errors.New("db not start correctly. Output: " + o)
	}

	fmt.Println("Run describe-db " + pxc.dbName + " with json")
	o, err = runCmd(pxc.cmd, pxc.subCmd, "describe-db", pxc.dbName, "-o", "json")
	if err != nil {
		return errors.Wrap(err, "run describe-db for"+pxc.dbName+" cmd")
	}
	fmt.Println(o)
	var data struct {
		DB DB `json:"database"`
	}
	err = json.Unmarshal([]byte(o), &data)
	if err != nil {
		return errors.Wrap(err, "unmarshal json out")
	}
	if data.DB.ResourceName != pxc.dbName {
		return errors.New("Wrong name")
	}
	return nil
}

func (pxc *KuberPXC) ModifyDB() error {
	fmt.Println("Run modify-db " + pxc.dbName)
	o, err := runCmd(pxc.cmd, pxc.subCmd, "modify-db", pxc.dbName, "--options", "pxc.resources.requests.memory=1G")
	if err != nil {
		return errors.Wrap(err, "run modify-db cmd")
	}
	if !strings.Contains(o, "ready") {
		return errors.New("db not modified correctly. Output: " + o)
	}
	return nil
}

func (pxc *KuberPXC) DeleteDB(preserve bool) error {
	fmt.Println("Run delete-db "+pxc.dbName+". Preserve flag is", preserve)
	preserveFlag := ""
	if preserve {
		preserveFlag = "--preserve-data"
	}
	o, err := runCmd(pxc.cmd, pxc.subCmd, "delete-db", pxc.dbName, "-y", preserveFlag)
	if err != nil {
		return errors.Wrap(err, "run delete-db cmd")
	}
	fmt.Println(o)
	if !strings.Contains(o, "done") {
		return errors.New("db not deleted correctly. Output: " + o)
	}
	return nil
}
