package tools

import (
	"errors"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
)

func GetInstance(name, options, engine, provider, rootPass string) dbaas.Instance {
	if len(options) > 0 {
		options = options
	}

	return dbaas.Instance{
		Name:          name,
		EngineOptions: options,
		Engine:        engine,
		Provider:      provider,
		RootPass:      rootPass,
	}
}

func GetDB(instance dbaas.Instance, hidePass, noWait bool, maxTries int) (dbaas.DB, error) {
	cluster := dbaas.DB{}
	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		cluster, err := dbaas.DescribeDB(instance)
		if err != nil && err != k8s.ErrOutOfMemory {
			//log.Error("check db: ", err)
			continue
		}
		if hidePass {
			cluster.Pass = "PASSWORD"
		}
		switch cluster.Status {
		case dbaas.StateReady:
			cluster.Message = strings.Replace(cluster.Message, "PASSWORD", cluster.Pass, 1)
			return cluster, nil
		case dbaas.StateInit:
			if noWait {
				cluster.Message = strings.Replace(cluster.Message, "PASSWORD", cluster.Pass, 1)
				return cluster, nil
			}
		case dbaas.StateError:
			return cluster, err
		}

		if tries >= maxTries {
			return cluster, errors.New("cluster status: " + string(cluster.Status))
		}
		tries++
	}

	return cluster, errors.New("cluster status: " + string(cluster.Status))
}
