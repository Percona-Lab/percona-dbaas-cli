package apitester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/broker"
	"github.com/Percona-Lab/percona-dbaas-cli/integtests/datafactory"
)

// Config describes apitester configuration
type Config struct {
	Address string
	Cases   []Case
}

type Case struct {
	Endpoint string
	ReqType  string
	ReqData  []byte
	RespData string
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

func (c *Config) check(testCase Case) error {
	controller, err := broker.New()
	ts := httptest.NewServer(http.HandlerFunc(controller.CreateServiceInstance))
	defer ts.Close()

	if testCase.Endpoint == datafactory.GetGetPXCInstanceData().Endpoint {
		return c.checkStatus(ts, testCase)
	}
	respData, respStatus, err := c.Request(ts.URL, testCase.ReqType, testCase.ReqData)
	if err != nil {
		return err
	}
	if respStatus != http.StatusOK || respStatus != http.StatusAccepted {
		log.Println(respStatus)
		return fmt.Errorf("Wrong status")
	}
	log.Println(respStatus, string(respData))
	if len(testCase.ReqData) > 0 && string(respData) != testCase.RespData {
		return fmt.Errorf("Wrong responce")
	}
	return nil
}

func (c *Config) checkStatus(ts *httptest.Server, testCase Case) error {
	startTime := time.Now()
	endTime := startTime.Local().Add(time.Second * time.Duration(180))
	ticker := time.NewTicker(2 * time.Second)

	var instResp GetInstanceResp
	for t := range ticker.C {
		respData, _, err := c.Request(ts.URL, testCase.ReqType, testCase.ReqData)
		if err != nil {
			return err
		}
		err = json.Unmarshal(respData, &instResp)
		if err != nil {
			return err
		}
		if instResp.State != "succeeded" {
			return nil
		}
		if t.Unix() >= endTime.Unix() {
			ticker.Stop()
			return fmt.Errorf("cluster not started")
		}
	}
	ticker.Stop()

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
	createPXCInstanceData := datafactory.GetCreatePXCInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: createPXCInstanceData.Endpoint,
		ReqType:  createPXCInstanceData.ReqType,
		ReqData:  createPXCInstanceData.ReqData,
		RespData: createPXCInstanceData.RespData,
	})
	getPXCInstanceData := datafactory.GetGetPXCInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: getPXCInstanceData.Endpoint,
		ReqType:  getPXCInstanceData.ReqType,
		ReqData:  getPXCInstanceData.ReqData,
		RespData: getPXCInstanceData.RespData,
	})
	updatePXCInstanceData := datafactory.GetGetPXCInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: updatePXCInstanceData.Endpoint,
		ReqType:  updatePXCInstanceData.ReqType,
		ReqData:  updatePXCInstanceData.ReqData,
		RespData: updatePXCInstanceData.RespData,
	})
	getPXCInstanceData = datafactory.GetGetPXCInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: getPXCInstanceData.Endpoint,
		ReqType:  getPXCInstanceData.ReqType,
		ReqData:  getPXCInstanceData.ReqData,
		RespData: getPXCInstanceData.RespData,
	})
	deletePXCInstanceData := datafactory.GetDeletePXCInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: deletePXCInstanceData.Endpoint,
		ReqType:  deletePXCInstanceData.ReqType,
		ReqData:  deletePXCInstanceData.ReqData,
		RespData: deletePXCInstanceData.RespData,
	})
	createPSMDBInstanceData := datafactory.GetCreatePSMDBInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: createPSMDBInstanceData.Endpoint,
		ReqType:  createPSMDBInstanceData.ReqType,
		ReqData:  createPSMDBInstanceData.ReqData,
		RespData: createPSMDBInstanceData.RespData,
	})
	getPSMDBInstanceData := datafactory.GetGetPSMDBInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: getPSMDBInstanceData.Endpoint,
		ReqType:  getPSMDBInstanceData.ReqType,
		ReqData:  getPSMDBInstanceData.ReqData,
		RespData: getPSMDBInstanceData.RespData,
	})
	updatePSMDBInstanceData := datafactory.GetGetPSMDBInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: updatePSMDBInstanceData.Endpoint,
		ReqType:  updatePSMDBInstanceData.ReqType,
		ReqData:  updatePSMDBInstanceData.ReqData,
		RespData: updatePSMDBInstanceData.RespData,
	})
	getPSMDBInstanceData = datafactory.GetGetPSMDBInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: getPSMDBInstanceData.Endpoint,
		ReqType:  getPSMDBInstanceData.ReqType,
		ReqData:  getPSMDBInstanceData.ReqData,
		RespData: getPSMDBInstanceData.RespData,
	})
	deletePSMDBInstanceData := datafactory.GetDeletePSMDBInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: deletePSMDBInstanceData.Endpoint,
		ReqType:  deletePSMDBInstanceData.ReqType,
		ReqData:  deletePSMDBInstanceData.ReqData,
		RespData: deletePSMDBInstanceData.RespData,
	})
}

type GetInstanceResp struct {
	State       string `json:"state"`
	Description string `json:"description"`
}
