package pxc

import (
	"encoding/json"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
	"github.com/pkg/errors"
)

// CreateDBCluster start creating DB cluster
func (p *PXC) CreateDBCluster(name string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].pxc
	p.config.Name = name
	err := p.setup(c, p.config, s3stor, p.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration: ")
	}
	cr, err := p.getCR(c)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.CreateCluster("pxc", p.config.OperatorImage, name, cr, p.bundle(objects, p.config.OperatorImage))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// CheckDBClusterStatus status return Cluster object with cluster info
func (p *PXC) CheckDBClusterStatus(name string) (structs.DB, error) {
	var db structs.DB
	secrets, err := p.cmd.GetSecrets(name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster secrets")

	}
	status, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster status")

	}

	st := &k8sStatus{}
	err = json.Unmarshal(status, st)
	if err != nil {
		return db, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		db.ResourceEndpoint = st.Status.Host
		db.Port = 3306
		db.User = "root"
		db.Pass = string(secrets["root"])
		db.Status = k8s.ClusterStateReady
		return db, nil
	case AppStateInit:
		db.Status = k8s.ClusterStateInit
		return db, nil
	case AppStateError:
		db.Status = k8s.ClusterStateError
		return db, errors.New(st.Status.Messages[0])
	default:
		return db, errors.New("unknown status")
	}
}

// DeleteDBCluster delete cluster by name
func (p *PXC) DeleteDBCluster(name string, delePVC bool) (string, error) {
	ext, err := p.cmd.IsObjExists("pxc", name)
	if err != nil {
		return "", errors.Wrap(err, "check if cluster exists")
	}

	if !ext {
		return "", errors.New("unable to find cluster pxc/" + name)
	}

	err = p.cmd.DeleteCluster("pxc", p.operatorName(), name, delePVC)
	if err != nil {
		return "", errors.Wrap(err, "delete cluster")
	}
	if !delePVC {
		pvcObj, err := p.cmd.GetObject("pvc", "datadir-"+name+"-pxc-0")
		if err != nil {
			return "", errors.Wrap(err, "get pvc")
		}
		pvc := &k8sPVC{}
		err = json.Unmarshal(pvcObj, pvc)
		if err != nil {
			return "", errors.Wrap(err, "unmarshal pvc")
		}
		return "pvc/" + pvc.Meta.Name, nil
	}

	return "", nil
}

// GetDBCluster return DB object
func (p *PXC) GetDBCluster(name string) (structs.DB, error) {
	var db structs.DB
	secrets, err := p.cmd.GetSecrets(name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster secrets")

	}
	cluster, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster object")

	}

	st := &k8sStatus{}
	err = json.Unmarshal(cluster, st)
	if err != nil {
		return db, errors.Wrap(err, "unmarshal object")
	}

	db.Provider = provider
	db.Engine = engine
	db.ResourceName = name
	db.ResourceEndpoint = st.Status.Host + "." + name + ".pxc.svc.local"
	db.Port = 3306
	db.User = "root"
	db.Pass = string(secrets["root"])
	db.Status = string(st.Status.Status)
	if st.Status.Status == "ready" {
		db.Message = "To access database please run the following commands:\nkubectl port-forward svc/" + name + "-proxysql 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -p" + db.Pass
	}

	return db, nil
}

// GetDBClusterList return list of existing DB obkects
func (p *PXC) GetDBClusterList() ([]structs.DB, error) {
	var dbList []structs.DB
	cluster, err := p.cmd.GetObjects("pxc")
	if err != nil {
		return dbList, errors.Wrap(err, "get cluster object")

	}
	st := &k8sCluster{}
	err = json.Unmarshal(cluster, st)
	if err != nil {
		return dbList, errors.Wrap(err, "unmarshal object")
	}
	for _, c := range st.Items {
		db := structs.DB{
			ResourceName: c.Meta.Name,
			Status:       string(c.Status.Status),
		}
		dbList = append(dbList, db)
	}
	return dbList, nil
}

// UpdateDBCluster update DB
func (p *PXC) UpdateDBCluster(name string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].pxc
	oldCR, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return errors.Wrap(err, "get cluster cr")
	}
	err = json.Unmarshal(oldCR, &c)
	if err != nil {
		return errors.Wrap(err, "unmarshal cr")
	}
	p.config.Name = name
	err = p.setup(c, p.config, s3stor, p.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration")
	}
	cr, err := p.getCR(c)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.Upgrade("pxc", name, cr)
	if err != nil {
		return errors.Wrap(err, "upgrade cluster")
	}

	return nil
}
