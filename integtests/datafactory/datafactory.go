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
		RespData:   []byte(`{"dashboard_url":"","last_operation":{"state":"in progress","description":"creating service instance...","async_poll_interval_seconds":10}}`),
	}
}

func GetGetPXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData:   []byte(`{"id":"test-pxc-instance","dashboard_url":"","internalId":"test-pxc-instance","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","last_operation":{"state":"succeeded","description":"successfully created service instance","async_poll_interval_seconds":10},"parameters":{"cluster_name":"test-pxc","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{},"credentials":{"host":"","port":0,"users":{"clustercheck":"dkpPNjNYTUZUYmJyaldRRkNz","monitor":"MDU3QWFjcDJZOXdTMjg5M0pz","proxyadmin":"UlFQWUg1dTNjdWpSU0FaUzFKMQ==","root":"djBSOENpeUk3TWU0T1pBOVlR","xtrabackup":"anlRMVZiUkdyT09GRXJIdHRDWA=="}},"credentialData":{"message":"MySQL cluster started successfully","host":"test-pxc-proxysql","port":3306,"user":"root","pass":"v0R8CiyI7Me4OZA9YQ"}}`),
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
		RespData:   []byte(`{"dashboard_url":"","last_operation":{"state":"in progress","description":"creating service instance...","async_poll_interval_seconds":10}}`),
	}
}
func GetGetPXCInstanceUpdatedData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData:   []byte(`{"id":"test-pxc-instance","dashboard_url":"","internalId":"test-pxc-instance","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","last_operation":{"state":"succeeded","description":"successfully created service instance","async_poll_interval_seconds":10},"parameters":{"cluster_name":"test-pxc","replicas":5},"context":{},"credentials":{"host":"","port":0,"users":{"clustercheck":"dkpPNjNYTUZUYmJyaldRRkNz","monitor":"MDU3QWFjcDJZOXdTMjg5M0pz","proxyadmin":"UlFQWUg1dTNjdWpSU0FaUzFKMQ==","root":"djBSOENpeUk3TWU0T1pBOVlR","xtrabackup":"anlRMVZiUkdyT09GRXJIdHRDWA=="}},"credentialData":null}`),
	}
}

func GetDeletePXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "DELETE",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData:   []byte(""),
	}
}

func GetGetDeletedPXCInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-pxc-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 404,
		RespData:   []byte(""),
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
		RespData:   []byte(`{"dashboard_url":"","last_operation":{"state":"in progress","description":"creating service instance...","async_poll_interval_seconds":10}}`),
	}
}

func GetGetPSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData:   []byte(`{"id":"test-psmdb-instance","dashboard_url":"","internalId":"test-psmdb-instance","service_id":"percona-server-for-mongodb-id","plan_id":"percona-server-for-mongodb-id","organization_guid":"","space_guid":"","last_operation":{"state":"succeeded","description":"successfully created service instance","async_poll_interval_seconds":10},"parameters":{"cluster_name":"test-psmdb","replicas":3,"topology_key":"none","size":"1Gi"},"context":{},"credentials":{"host":"","port":0,"users":{"MONGODB_BACKUP_PASSWORD":"T1UxQnRWV2s3dGdyZVdsUQ==","MONGODB_BACKUP_USER":"YmFja3Vw","MONGODB_CLUSTER_ADMIN_PASSWORD":"eU1pZnVPOEpjMmY5Y2RjdGs=","MONGODB_CLUSTER_ADMIN_USER":"Y2x1c3RlckFkbWlu","MONGODB_CLUSTER_MONITOR_PASSWORD":"NVYxdzZCcm4zVFhSZEdkV283QQ==","MONGODB_CLUSTER_MONITOR_USER":"Y2x1c3Rlck1vbml0b3I=","MONGODB_USER_ADMIN_PASSWORD":"S0c3ZVlmQ0EzRFoxQlRtSQ==","MONGODB_USER_ADMIN_USER":"dXNlckFkbWlu"}},"credentialData":{"message":"MomgoDB cluster started successfully","host":"test-psmdb-test-psmdb-0.test-psmdb-test-psmdb","port":27017,"clusterAdminUser":"clusterAdmin","clusterAdminPass":"5V1w6Brn3TXRdGdWo7A","userAdminUser":"userAdmin","userAdminPass":"KG7eYfCA3DZ1BTmI"}}`),
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
		RespData:   []byte(`{"dashboard_url":"","last_operation":{"state":"in progress","description":"creating service instance...","async_poll_interval_seconds":10}}`),
	}
}

func GetGetPSMDBInstanceUpdatedData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData:   []byte(`{"id":"test-psmdb-instance","dashboard_url":"","internalId":"test-psmdb-instance","service_id":"percona-server-for-mongodb-id","plan_id":"percona-server-for-mongodb-id","organization_guid":"","space_guid":"","last_operation":{"state":"succeeded","description":"successfully created service instance","async_poll_interval_seconds":10},"parameters":{"cluster_name":"test-psmdb","replicas":5},"context":{},"credentials":{"host":"","port":0,"users":{"MONGODB_BACKUP_PASSWORD":"T1UxQnRWV2s3dGdyZVdsUQ==","MONGODB_BACKUP_USER":"YmFja3Vw","MONGODB_CLUSTER_ADMIN_PASSWORD":"eU1pZnVPOEpjMmY5Y2RjdGs=","MONGODB_CLUSTER_ADMIN_USER":"Y2x1c3RlckFkbWlu","MONGODB_CLUSTER_MONITOR_PASSWORD":"NVYxdzZCcm4zVFhSZEdkV283QQ==","MONGODB_CLUSTER_MONITOR_USER":"Y2x1c3Rlck1vbml0b3I=","MONGODB_USER_ADMIN_PASSWORD":"S0c3ZVlmQ0EzRFoxQlRtSQ==","MONGODB_USER_ADMIN_USER":"dXNlckFkbWlu"}},"credentialData":null}`),
	}
}

func GetDeletePSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "DELETE",
		ReqData:    []byte(``),
		RespStatus: 200,
		RespData:   []byte(""),
	}
}

func GetGetDeletedPSMDBInstanceData() structs.CaseData {
	return structs.CaseData{
		Endpoint:   "/v2/service_instances/test-psmdb-instance",
		ReqType:    "GET",
		ReqData:    []byte(``),
		RespStatus: 404,
		RespData:   []byte(""),
	}
}
