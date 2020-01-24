module github.com/Percona-Lab/percona-dbaas-cli

go 1.13

replace github.com/percona/percona-server-mongodb-operator/v110 => github.com/percona/percona-server-mongodb-operator v0.0.0-20190707075059-f6a9dada369e

replace github.com/percona/percona-server-mongodb-operator/v120 => github.com/percona/percona-server-mongodb-operator v0.0.0-20190920082537-70382eaea511

replace github.com/percona/percona-xtradb-cluster-operator/v110 => github.com/percona/percona-xtradb-cluster-operator v0.0.0-20190729091217-40c60f550f94

replace github.com/percona/percona-xtradb-cluster-operator/v120 => github.com/percona/percona-xtradb-cluster-operator v0.0.0-20190920084837-406b624f1baf

require (
	github.com/Percona-Lab/pxc-service-broker v0.0.0-20190719091759-795f14c411c4
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/percona/percona-server-mongodb-operator v0.0.0-20200108163216-2e5889710e1b // indirect
	github.com/percona/percona-server-mongodb-operator/v110 v110.0.0
	github.com/percona/percona-server-mongodb-operator/v120 v120.0.0
	github.com/percona/percona-xtradb-cluster-operator/v110 v110.0.0-00010101000000-000000000000
	github.com/percona/percona-xtradb-cluster-operator/v120 v120.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)
