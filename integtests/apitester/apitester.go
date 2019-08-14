package apitester

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
		log.Println(testCase.Endpoint)
		err := c.check(testCase)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) check(testCase Case) error {
	respData, respStatus, err := c.Request(testCase.Endpoint, testCase.ReqType, testCase.ReqData)
	if err != nil {
		return err
	}
	if respStatus != http.StatusOK || respStatus != http.StatusAccepted {
		log.Println(respStatus)
		//return fmt.Errorf("Wrong status")
	}
	log.Println(string(respData))
	if string(respData) != testCase.RespData {
		return fmt.Errorf("Wrong responce")
	}
	return nil
}

// Request send test request
func (c *Config) Request(endpoint, reqType string, reqBody []byte) ([]byte, int, error) {
	client := http.Client{}
	req, err := http.NewRequest(reqType, c.Address+endpoint, bytes.NewReader(reqBody))
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
	//var cases []Case
	/*createPXCInstanceData := datafactory.GetCreatePXCInstanceData()
	t.Cases = append(t.Cases, Case{
		Endpoint: createPXCInstanceData.Endpoint,
		ReqType:  createPXCInstanceData.ReqType,
		ReqData:  createPXCInstanceData.ReqData,
		RespData: createPXCInstanceData.RespData,
	})*/
	getPXCInstanceData := datafactory.GetGetPXCInstanceData()
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
}
