package broker

import "testing"

func TestUpdateInstanceMapTests(t *testing.T) {
	var c Controller
	c.instanceMap = make(map[string]*ServiceInstance)
	s := []byte(`'{"id":"","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}}'`)
	err := c.updateInstanceMap(s)
	if err != nil {
		t.Error("one instance:", err)
	}
	s = []byte(`'{"id":"","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}} {"id":"","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}}'`)
	err = c.updateInstanceMap(s)
	if err != nil {
		t.Error("two instance:", err)
	}
	s = []byte(`'{"id":"","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}} {"id":"","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}} {"id":"","dashboard_url":"","service_id":"percona-xtradb-cluster-id","plan_id":"percona-xtradb-id","organization_guid":"","space_guid":"","parameters":{"cluster_name":"test","replicas":3,"proxy_sql_replicas":1,"topology_key":"none","size":"1Gi"},"context":{}}'`)
	err = c.updateInstanceMap(s)
	if err != nil {
		t.Error("three instance:", err)
	}
	s = []byte(`'{}}'`)
	err = c.updateInstanceMap(s)
	if err == nil {
		t.Error("invalid JSON:", err)
	}
}
