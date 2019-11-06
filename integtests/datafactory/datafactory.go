package datafactory

import "github.com/Percona-Lab/percona-dbaas-cli/integtests/structs"

func CreatePXCInstanceData() structs.CaseData {
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
				State:                    structs.InProgressOperationSate,
				Description:              structs.InProgressOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}

type parameters struct {
	structs.Parameters
}

func GetPXCInstanceData() structs.CaseData {
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
				State:                    structs.SucceedOperationState,
				Description:              structs.SucceedOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func UpdatePXCInstanceData() structs.CaseData {
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
				State:                    structs.InProgressOperationSate,
				Description:              structs.InProgressOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}
func GetPXCInstanceUpdatedData() structs.CaseData {
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
				State:                    structs.SucceedOperationState,
				Description:              structs.SucceedOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func DeletePXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "DELETE",
		ReqData:    []byte(``),
		RespStatus: 200,
	}
}

func GetDeletedPXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 404,
	}
}

func CreatePSMDBInstanceData() structs.CaseData {
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
				State:                    structs.InProgressOperationSate,
				Description:              structs.InProgressOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}

func GetPSMDBInstanceData() structs.CaseData {
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
				State:                    structs.SucceedOperationState,
				Description:              structs.SucceedOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func UpdatePSMDBInstanceData() structs.CaseData {
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
				State:                    structs.InProgressOperationSate,
				Description:              structs.InProgressOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
		},
	}
}

func GetPSMDBInstanceUpdatedData() structs.CaseData {
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
				State:                    structs.SucceedOperationState,
				Description:              structs.SucceedOperationDescription,
				AsyncPollIntervalSeconds: 10,
			},
			Parameters: params,
		},
	}
}

func DeletePSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "DELETE",
		ReqData:    []byte(``),
		RespStatus: 200,
	}
}

func GetDeletedPSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 404,
	}
}
