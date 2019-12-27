module github.com/Percona-Lab/percona-dbaas-cli

go 1.12

require (
	github.com/briandowns/spinner v1.8.0
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/percona/percona-server-mongodb-operator v0.0.0-20191225084541-e0ae3a7ef113
	github.com/percona/percona-xtradb-cluster-operator v0.0.0-20191225110028-1fc5469f3719
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apimachinery v0.0.0-20191123233150-4c4803ed55e3
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)
