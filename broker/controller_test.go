package broker

import (
	"testing"
)

func TestUpdateInstanceMapTests(t *testing.T) {
	var c Controller
	c.instanceMap = make(map[string]*ServiceInstance)
	type test struct {
		instQuant    int
		clusterNames map[string]string
		data         []byte
	}
	tests := map[string]test{
		"one instance": {
			instQuant: 1,
			clusterNames: map[string]string{
				"pxc-1": "test",
			},
			data: []byte(`'{"id":"pxc-1","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}}'`),
		},
		"two instances": {
			instQuant: 2,
			clusterNames: map[string]string{
				"pxc-1":   "test",
				"psmdb-1": "psmdb",
			},
			data: []byte(`'{"id":"pxc-1","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}} {"id":"psmdb-1","dashboard_url":"","service_id":"percona-server-for-mongodb-id","plan_id":"percona-server-for-mongodb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"psmdb","replicas":5,"proxy_sql_replicas":1,"topology_key":"none","size":"2Gi"},"context":{}}'`),
		},
		"three instances": {
			instQuant: 3,
			clusterNames: map[string]string{
				"pxc-1":   "test",
				"psmdb-1": "psmdb",
				"pxc-2":   "pxc-2",
			},
			data: []byte(`'{"id":"pxc-1","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}} {"id":"psmdb-1","dashboard_url":"","service_id":"percona-server-for-mongodb-id","plan_id":"percona-server-for-mongodb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"psmdb","replicas":5,"proxy_sql_replicas":1,"topology_key":"none","size":"2Gi"},"context":{}} {"id":"pxc-2","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"pxc-2","replicas":5,"proxy_sql_replicas":1,"topology_key":"none","size":"3Gi"},"context":{}}'`),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := c.updateInstanceMap(test.data)
			if err != nil {
				t.Error(name, err)
			}
			if len(c.instanceMap) != test.instQuant {
				t.Error(name+":", "Wrong instance quantity")
			}
			for id, clusterName := range test.clusterNames {
				if clusterName != c.instanceMap[id].Parameters.ClusterName {
					t.Error(name+":", "Wrong instance data")
				}
			}
		})
	}
}
