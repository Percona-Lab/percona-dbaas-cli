package apitester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/integtests/datafactory"
	"github.com/Percona-Lab/percona-dbaas-cli/integtests/structs"
	"github.com/pkg/errors"
)

// Config describes apitester configuration
type Config struct {
	Address string
	Cases   []structs.CaseData
}

// New return new apitester
func New(addres string) Config {
	var c = Config{
		Address: addres,
	}
	c.setCases()
	return c
}

// Run starts testing
func (c *Config) Run() error {
	for _, testCase := range c.Cases {
		log.Println(testCase.Endpoint, testCase.ReqType)
		err := c.check(testCase)
		if err != nil {
			return errors.Wrap(err, "check failed")
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (c *Config) check(testCase structs.CaseData) error {
	// Here we waitning
	if testCase.ReqType == "GET" && testCase.RespStatus == 200 {
		return c.waitForSucceed(testCase)
	}
	respData, respStatus, err := c.Request(c.Address+testCase.Endpoint, testCase.ReqType, testCase.ReqData)
	if err != nil {
		return err
	}
	if respStatus != testCase.RespStatus {
		return errors.New("wrong resp status. Have '" + strconv.Itoa(respStatus) + "', want " + strconv.Itoa(testCase.RespStatus))
	}

	return checkRespData(testCase.RespData, respData)
}

type GetInstanceResp struct {
	LastOperation struct {
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"last_operation"`
}

func (c *Config) waitForSucceed(testCase structs.CaseData) error {
	endTime := time.Now().Add(time.Second * 250)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var instResp GetInstanceResp
	for t := range ticker.C {
		respData, status, err := c.Request(c.Address+testCase.Endpoint, testCase.ReqType, testCase.ReqData)
		if err != nil {
			return errors.Wrap(err, "request")
		}
		err = json.Unmarshal(respData, &instResp)
		if err != nil {
			log.Println(err)
			continue
		}

		if instResp.LastOperation.State == "succeeded" {
			fmt.Println() // print a new line after the ticker
			if status != testCase.RespStatus {
				return errors.New("wrong resp status. Have '" + strconv.Itoa(status) + "', want " + strconv.Itoa(testCase.RespStatus))
			}
			return checkRespData(testCase.RespData, respData)
		}
		fmt.Printf("\r Wait for cluster. %v tries left  ", (endTime.Unix()-t.Unix())/2)
		if t.Unix() >= endTime.Unix() {
			fmt.Println() // print a new line after the ticker
			return errors.New("cluster not started")
		}
	}

	return nil
}

func checkRespData(expected structs.ServiceInstance, respData []byte) error {
	if expected.LastOperation == nil && len(respData) == 0 {
		return nil
	}
	var respInst structs.ServiceInstance

	err := json.Unmarshal(respData, &respInst)
	if err != nil {
		return errors.Wrap(err, "check responce unmarshal")
	}
	if respInst.LastOperation.State != expected.LastOperation.State {
		return errors.New(fmt.Sprintf("wrong state. Have %s, want %s", respInst.LastOperation.State, expected.LastOperation.State))
	}
	if respInst.Parameters.Replicas != expected.Parameters.Replicas {
		return errors.New(fmt.Sprintf("wrong replicas number. Have %d, want %d", respInst.Parameters.Replicas, expected.Parameters.Replicas))
	}
	if respInst.Parameters.ClusterName != expected.Parameters.ClusterName {
		return errors.New(fmt.Sprintf("wrong cluster name. Have %s, want %s", respInst.Parameters.ClusterName, expected.Parameters.ClusterName))
	}
	if respInst.ID != expected.ID {
		return errors.New(fmt.Sprintf("wrong ID. Have %s, want %s", respInst.ID, expected.ID))
	}
	return nil
}

// Request send test request
func (c *Config) Request(address, reqType string, reqBody []byte) ([]byte, int, error) {
	client := http.Client{}
	req, err := http.NewRequest(reqType, address, bytes.NewReader(reqBody))
	if err != nil {
		return nil, 0, errors.Wrap(err, "new request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, errors.Wrap(err, "read body")
	}

	return body, resp.StatusCode, nil
}

func (c *Config) setCases() {
	c.Cases = append(c.Cases, datafactory.CreatePXCInstanceData())
	c.Cases = append(c.Cases, datafactory.GetPXCInstanceData())
	c.Cases = append(c.Cases, datafactory.UpdatePXCInstanceData())
	c.Cases = append(c.Cases, datafactory.GetPXCInstanceUpdatedData())
	c.Cases = append(c.Cases, datafactory.DeletePXCInstanceData())
	c.Cases = append(c.Cases, datafactory.GetDeletedPXCInstanceData())

	c.Cases = append(c.Cases, datafactory.CreatePSMDBInstanceData())
	c.Cases = append(c.Cases, datafactory.GetPSMDBInstanceData())
	c.Cases = append(c.Cases, datafactory.UpdatePSMDBInstanceData())
	c.Cases = append(c.Cases, datafactory.GetPSMDBInstanceUpdatedData())
	c.Cases = append(c.Cases, datafactory.DeletePSMDBInstanceData())
	c.Cases = append(c.Cases, datafactory.GetDeletedPSMDBInstanceData())
}
