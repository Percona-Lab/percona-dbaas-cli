package datafactory

import "github.com/Percona-Lab/percona-dbaas-cli/integtests/structs"

func GetCreatePXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint: "/v2/service_instances/test-pxc-instance",
		ReqType:  "PUT",
		ReqData: []byte(`{
			"service_id":"percona-xtradb-cluster-id",
			"plan_id":"percona-xtradb-id",
			"parameters":{
				"cluster_name":"test-pxc",
				"replicas":3,
				"proxy_sql_replicas":1,
				"topology_key": "none",
				"size": "1Gi"
				}
			}`),
		RespStatus: 202,
		RespData: structs.ServiceInstance{
			LastOperation: &structs.LastOperation{
				State:                    "in progress",
				Description:              "creating service instance...",
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}

type parameters struct {
	structs.Parameters
}

func GetGetPXCInstanceData() structs.CaseData {
	var params parameters
	params.ClusterName = "test-pxc"
	params.Replicas = 3

	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData: structs.ServiceInstance{
			ID:        "test-pxc-instance",
			ServiceID: "percona-xtradb-cluster-id",
			PlanID:    "percona-xtradb-id",
			LastOperation: &structs.LastOperation{
				State:                    "succeeded",
				Description:              "successfully created service instance",
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func GetUpdatePXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint: "/v2/service_instances/test-pxc-instance",
		ReqType:  "UPDATE",
		ReqData: []byte(`{
			"service_id":"percona-xtradb-cluster-id",
			"plan_id":"percona-xtradb-id",
			"parameters":{
				"cluster_name":"test-pxc",
				"replicas":5
				}
			}`),
		RespStatus: 202,
		RespData: structs.ServiceInstance{
			LastOperation: &structs.LastOperation{
				State:                    "in progress",
				Description:              "creating service instance...",
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}
func GetGetPXCInstanceUpdatedData() structs.CaseData {
	var params parameters
	params.ClusterName = "test-pxc"
	params.Replicas = 5
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData: structs.ServiceInstance{
			ID:        "test-pxc-instance",
			ServiceID: "percona-xtradb-cluster-id",
			PlanID:    "percona-xtradb-id",
			LastOperation: &structs.LastOperation{
				State:                    "succeeded",
				Description:              "successfully created service instance",
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func GetDeletePXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "DELETE",
		ReqData:    []byte(``),
		RespStatus: 200,
	}
}

func GetGetDeletedPXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 404,
	}
}

func GetCreatePSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint: "/v2/service_instances/test-psmdb-instance",
		ReqType:  "PUT",
		ReqData: []byte(`{
			"service_id":"percona-server-for-mongodb-id",
			"plan_id":"percona-server-for-mongodb-id",
			"parameters":{
				"cluster_name":"test-psmdb",
				"replicas":3,
				"topology_key": "none",
				"size": "1Gi"
				}
			}`),
		RespStatus: 202,
		RespData: structs.ServiceInstance{
			LastOperation: &structs.LastOperation{
				State:                    "in progress",
				Description:              "creating service instance...",
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}

func GetGetPSMDBInstanceData() structs.CaseData {
	var params parameters
	params.ClusterName = "test-psmdb"
	params.Replicas = 3
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData: structs.ServiceInstance{
			ID:        "test-psmdb-instance",
			ServiceID: "percona-server-for-mongodb-id",
			PlanID:    "percona-server-for-mongodb-id",
			LastOperation: &structs.LastOperation{
				State:                    "succeeded",
				Description:              "successfully created service instance",
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func GetUpdatePSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint: "/v2/service_instances/test-psmdb-instance",
		ReqType:  "UPDATE",
		ReqData: []byte(`{
			"service_id":"percona-server-for-mongodb-id",
			"plan_id":"percona-server-for-mongodb-id",
			"parameters":{
				"cluster_name":"test-psmdb",
				"replicas":5
				}
			}`),
		RespStatus: 202,
		RespData: structs.ServiceInstance{
			LastOperation: &structs.LastOperation{
				State:                    "in progress",
				Description:              "creating service instance...",
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}

func GetGetPSMDBInstanceUpdatedData() structs.CaseData {
	var params parameters
	params.ClusterName = "test-psmdb"
	params.Replicas = 5
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData: structs.ServiceInstance{
			ID:        "test-psmdb-instance",
			ServiceID: "percona-server-for-mongodb-id",
			PlanID:    "percona-server-for-mongodb-id",
			LastOperation: &structs.LastOperation{
				State:                    "succeeded",
				Description:              "successfully created service instance",
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func GetDeletePSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "DELETE",
		ReqData:    []byte(``),
		RespStatus: 200,
	}
}

func GetGetDeletedPSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 404,
	}
}
