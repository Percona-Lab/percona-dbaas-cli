package pxc

import (
	"encoding/json"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
	"github.com/pkg/errors"
)

// CreateDBCluster start creating DB cluster
func (p *PXC) CreateDBCluster(name, opts string) error {
	err := p.ParseOptions(opts)
	if err != nil {
		return err
	}
	p.conf.SetName(name)
	p.conf.SetUsersSecretName(name)
	cr, err := p.getCR(p.conf)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.CreateCluster("pxc", p.conf.GetOperatorImage(), name, cr, p.bundle(objects, p.conf.GetOperatorImage()))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// DeleteDBCluster delete cluster by name
func (p *PXC) DeleteDBCluster(name, opts string, delePVC bool) (string, error) {
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
func (p *PXC) GetDBCluster(name, opts string) (structs.DB, error) {
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
	ns, err := p.cmd.GetCurrentNamespace()
	if err != nil {
		return db, errors.Wrap(err, "get namspace name")
	}
	if len(ns) == 0 {
		ns = "default"
	}
	db.Provider = provider
	db.Engine = engine
	db.ResourceName = name
	db.ResourceEndpoint = st.Status.Host + "." + ns + ".pxc.svc.local"
	db.Port = 3306
	db.User = "root"
	db.Pass = string(secrets["root"])
	db.Status = string(st.Status.Status)
	if st.Status.Status == "ready" {
		db.Message = "To access database please run the following commands:\nkubectl port-forward svc/" + name + "-proxysql 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD"
	}
	if st.Status.Status == "unknown" && st.Status.PXC.Status == "ready" {
		db.Status = "ready"
		db.Message = "To access database please run the following commands:\nkubectl port-forward pod/" + name + "-pxc-0 3306:3306 &\nmysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD"
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
func (p *PXC) UpdateDBCluster(name, opts string) error {
	c := objects[defaultVersion].pxc
	oldCR, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return errors.Wrap(err, "get cluster cr")
	}
	err = json.Unmarshal(oldCR, &c)
	if err != nil {
		return errors.Wrap(err, "unmarshal cr")
	}
	p.conf = c
	p.ParseOptions(opts)
	p.conf.SetName(name)
	p.conf.SetUsersSecretName(name)
	cr, err := p.getCR(p.conf)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.Upgrade("pxc", name, cr)
	if err != nil {
		return errors.Wrap(err, "upgrade cluster")
	}

	return nil
}
