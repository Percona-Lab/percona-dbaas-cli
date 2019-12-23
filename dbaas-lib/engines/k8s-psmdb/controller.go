package psmdb

import (
	"encoding/json"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
	"github.com/pkg/errors"
)

// CreateDBCluster start creating DB cluster
func (p *PSMDB) CreateDBCluster(name string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].psmdb
	p.config.Name = name
	err := p.setup(c, p.config, s3stor, p.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration: ")
	}
	cr, err := p.getCR(c)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.CreateCluster("psmdb", p.config.OperatorImage, name, cr, p.bundle(objects, p.config.OperatorImage))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// DeleteDBCluster delete cluster by name
func (p *PSMDB) DeleteDBCluster(name string, delePVC bool) (string, error) {
	ext, err := p.cmd.IsObjExists("psmdb", name)
	if err != nil {
		return "", errors.Wrap(err, "check if cluster exists")
	}
	if !ext {
		return "", errors.New("unable to find cluster psmdb/" + name)
	}
	err = p.cmd.DeleteCluster("psmdb", p.operatorName(), name, delePVC)
	if err != nil {
		return "", errors.Wrap(err, "delete cluster")
	}
	if !delePVC {
		pvcObj, err := p.cmd.GetObject("pvc", "datadir-"+name+"-psmdb-0")
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
func (p *PSMDB) GetDBCluster(name string) (structs.DB, error) {
	var db structs.DB
	secrets, err := p.cmd.GetSecrets(name + "-psmdb-users")
	if err != nil {
		return db, errors.Wrap(err, "get cluster secrets")

	}
	cluster, err := p.cmd.GetObject("psmdb", name)
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
	db.ResourceEndpoint = name + "." + name + ".psmdb.svc.local"
	db.Port = 27017
	db.User = string(secrets["MONGODB_CLUSTER_ADMIN_USER"])
	db.Pass = string(secrets["MONGODB_CLUSTER_ADMIN_PASSWORD"])
	db.Status = string(st.Status.Status)
	if st.Status.Status == "ready" {
		db.Message = "To access database please run the following commands:\nkubectl port-forward svc/" + name + "-" + name + " 27017:27017 &\nmongo mongodb://" + db.User + ":PASSWORD@localhost:27017/admin?ssl=false"
	}

	return db, nil
}

// GetDBClusterList return list of existing DB obkects
func (p *PSMDB) GetDBClusterList() ([]structs.DB, error) {
	var dbList []structs.DB
	cluster, err := p.cmd.GetObjects("psmdb")
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
func (p *PSMDB) UpdateDBCluster(name string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].psmdb
	oldCR, err := p.cmd.GetObject("psmdb", name)
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
	err = p.cmd.Upgrade("psmdb", name, cr)
	if err != nil {
		return errors.Wrap(err, "upgrade cluster")
	}

	return nil
}
