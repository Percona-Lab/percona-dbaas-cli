package apitester

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/integtests/datafactory"
	"github.com/Percona-Lab/percona-dbaas-cli/integtests/structs"
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
			return err
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
		return fmt.Errorf("Wrong status")
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
	startTime := time.Now()
	endTime := startTime.Add(time.Second * 250)
	ticker := time.NewTicker(2 * time.Second)

	var instResp GetInstanceResp
	for t := range ticker.C {
		respData, status, err := c.Request(c.Address+testCase.Endpoint, testCase.ReqType, testCase.ReqData)
		if err != nil {
			return err
		}
		err = json.Unmarshal(respData, &instResp)
		if err != nil {
			log.Println(err)
			continue
		}

		if instResp.LastOperation.State == "succeeded" {
			fmt.Println() // just for create new line after ticker
			ticker.Stop()
			if status != testCase.RespStatus {
				return errors.New("Wrong resp status. Have '" + instResp.LastOperation.State + "', want 'succeeded'")
			}
			return checkRespData(testCase.RespData, respData)
		}
		fmt.Printf("\r Wait for cluster. %v tries left  ", (endTime.Unix()-t.Unix())/2)
		if t.Unix() >= endTime.Unix() {
			ticker.Stop()
			fmt.Println() // just for create new line after ticker
			return fmt.Errorf("cluster not started")
		}
	}

	return nil
}

func checkRespData(expected, respData []byte) error {
	if len(expected) == 0 && len(respData) == 0 {
		return nil
	}
	var respInst structs.ServiceInstance
	var expectInst structs.ServiceInstance
	err := json.Unmarshal(expected, &expectInst)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respData, &respInst)
	if err != nil {
		return err
	}
	if respInst.LastOperation.State != expectInst.LastOperation.State {
		return errors.New("wrong state")
	}
	if respInst.Parameters.Replicas != expectInst.Parameters.Replicas {
		return errors.New("wrong replicas number")
	}
	if respInst.Parameters.ClusterName != expectInst.Parameters.ClusterName {
		return errors.New("wrong cluster name")
	}
	if respInst.ID != expectInst.ID {
		return errors.New("wrong ID")
	}
	return nil
}

// Request send test request
func (c *Config) Request(address, reqType string, reqBody []byte) ([]byte, int, error) {
	client := http.Client{}
	req, err := http.NewRequest(reqType, address, bytes.NewReader(reqBody))
	if err != nil {
		return nil, 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}

func (t *Config) setCases() {
	t.Cases = append(t.Cases, datafactory.GetCreatePXCInstanceData())
	t.Cases = append(t.Cases, datafactory.GetGetPXCInstanceData())
	t.Cases = append(t.Cases, datafactory.GetUpdatePXCInstanceData())
	t.Cases = append(t.Cases, datafactory.GetGetPXCInstanceUpdatedData())
	t.Cases = append(t.Cases, datafactory.GetDeletePXCInstanceData())
	t.Cases = append(t.Cases, datafactory.GetGetDeletedPXCInstanceData())

	t.Cases = append(t.Cases, datafactory.GetCreatePSMDBInstanceData())
	t.Cases = append(t.Cases, datafactory.GetGetPSMDBInstanceData())
	t.Cases = append(t.Cases, datafactory.GetUpdatePSMDBInstanceData())
	t.Cases = append(t.Cases, datafactory.GetGetPSMDBInstanceUpdatedData())
	t.Cases = append(t.Cases, datafactory.GetDeletePSMDBInstanceData())
	t.Cases = append(t.Cases, datafactory.GetGetDeletedPSMDBInstanceData())
}
