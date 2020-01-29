module github.com/Percona-Lab/percona-dbaas-cli

go 1.13

replace github.com/percona/percona-server-mongodb-operator/v110 => github.com/percona/percona-server-mongodb-operator 1.1.0

replace github.com/percona/percona-server-mongodb-operator/v120 => github.com/percona/percona-server-mongodb-operator 1.2.0

replace github.com/percona/percona-xtradb-cluster-operator/v110 => github.com/percona/percona-xtradb-cluster-operator 1.1.0

replace github.com/percona/percona-xtradb-cluster-operator/v120 => github.com/percona/percona-xtradb-cluster-operator 1.2.0

replace github.com/percona/percona-xtradb-cluster-operator/v130 => github.com/percona/percona-xtradb-cluster-operator 1.3.0

require (
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)
