package datafactory

type CaseData struct {
	Endpoint string
	ReqType  string
	ReqData  []byte
	RespData string
}

func GetCreatePXCInstanceData() CaseData {
	return CaseData{
		Endpoint: "/v2/service_instances/test-pxc-instance",
		ReqType:  "PUT",
		ReqData: []byte(`{
			"service_id":"percona-xtradb-cluster-id",
			"plan_id":"percona-xtradb-id",
			"parameters":{
				"cluster_name":"test",
				"replicas":3,
				"proxy_sql_replicas":1,
				"topology_key": "none",
				"size": "1Gi"
				}
			}`),
		RespData: `{"dashboard_url":"","last_operation":{"state":"in progress","description":"creating service instance...","async_poll_interval_seconds":10}}`,
	}
}

func GetGetPXCInstanceData() CaseData {
	return CaseData{
		Endpoint: "/v2/service_instances/test-pxc-instance",
		ReqType:  "GET",
		ReqData:  []byte(``),
		RespData: `{"state":"in progress","description":"creating service instance...","async_poll_interval_seconds":10}`,
	}
}

func GetDeletePXCInstanceData() CaseData {
	return CaseData{
		Endpoint: "/v2/service_instances/test-pxc-instance",
		ReqType:  "DELETE",
		ReqData:  []byte(``),
		RespData: "{}",
	}
}
