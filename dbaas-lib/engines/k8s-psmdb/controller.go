package psmdb

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	mrand "math/rand"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
	"github.com/pkg/errors"
)

// CreateDBCluster start creating DB cluster
func (p *PSMDB) CreateDBCluster(name, opts, rootPass, version string) error {
	err := p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return errors.Wrap(err, "version check")
	}
	err = p.ParseOptions(opts)
	if err != nil {
		return errors.Wrap(err, "parse opts")
	}
	p.conf.SetName(name)
	p.conf.SetUsersSecretName(name)

	switch p.platformType {
	case k8s.PlatformMinishift, k8s.PlatformMinikube:
		p.conf.SetupMiniConfig()
	}

	if len(rootPass) > 0 {
		err = p.SetupPasswords(name, rootPass)
		if err != nil {
			return errors.Wrap(err, "set root password")
		}
	}

	cr, err := p.getCR(p.conf)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}
	_, err = p.cmd.GetObjectsElement("deployment", p.operatorName(), ".spec.template.spec.containers[0].image")
	if err != nil && err == k8s.ErrNotFound {
		p.cmd.ApplyBundles(p.bundle)
	}

	err = p.cmd.CreateCluster("psmdb", p.conf.GetOperatorImage(), name, cr, p.bundle)
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// DeleteDBCluster delete cluster by name
func (p *PSMDB) DeleteDBCluster(name, opts, version string, delePVC bool) (string, error) {
	ext, err := p.cmd.IsObjExists("psmdb", name)
	if err != nil {
		return "", errors.Wrap(err, "check if cluster exists")
	}
	if !ext {
		return "", errors.New("unable to find cluster psmdb/" + name)
	}
	err = p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return "", errors.Wrap(err, "version check")
	}
	p.conf.SetDefaults()
	p.conf.SetName(name)

	err = p.cmd.DeleteCluster("psmdb", p.operatorName(), name, delePVC)
	if err != nil {
		return "", errors.Wrap(err, "delete cluster")
	}
	if !delePVC {
		pvcObj, err := p.cmd.GetObject("pvc", "mongod-data-"+name+"-rs0-0")
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
	err = p.cmd.DeleteObject("secret", name+"-psmdb-users-secrets")
	if err != nil {
		return "", errors.Wrap(err, "delete secret")
	}

	return "", nil
}

// GetDBCluster return DB object
func (p *PSMDB) GetDBCluster(name, opts string) (structs.DB, error) {
	var db structs.DB
	secrets, err := p.cmd.GetSecrets(name + "-psmdb-users-secrets")
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
	db.ResourceEndpoint = name + "-rs0." + ns + ".psmdb.svc.local"
	db.Port = 27017
	db.User = string(secrets["MONGODB_CLUSTER_ADMIN_USER"])
	db.Pass = string(secrets["MONGODB_CLUSTER_ADMIN_PASSWORD"])
	db.Status = string(st.Status.Status)
	if st.Status.Status == "ready" {
		db.Message = "To access database please run the following commands:\nkubectl port-forward svc/" + name + "-rs0 27017:27017 &\nmongo mongodb://" + db.User + ":PASSWORD@localhost:27017/admin?ssl=false"
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
func (p *PSMDB) UpdateDBCluster(name, opts, version string) error {
	err := p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return errors.Wrap(err, "version check")
	}

	oldCR, err := p.cmd.GetObject("psmdb", name)
	if err != nil {
		return errors.Wrap(err, "get cluster cr")
	}
	err = json.Unmarshal(oldCR, &p.conf)
	if err != nil {
		return errors.Wrap(err, "unmarshal cr")
	}
	p.ParseOptions(opts)
	p.conf.SetName(name)
	p.conf.SetUsersSecretName(name)

	cr, err := p.getCR(p.conf)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}
	err = p.cmd.Upgrade("psmdb", name, cr)
	if err != nil {
		return errors.Wrap(err, "upgrade cluster")
	}

	return nil
}

func (p *PSMDB) SetupPasswords(clusterName, rootPass string) error {
	secretName := clusterName + "-psmdb-users-secrets"
	ext, err := p.cmd.IsObjExists("secret", secretName)
	if err != nil {
		return errors.Wrap(err, "check if secrets exists")
	}
	data := map[string][]byte{}
	if ext {
		data, err = p.cmd.GetSecrets(secretName)
		if err != nil {
			return errors.Wrap(err, "get secrets")
		}
		for k := range data {
			if k == "MONGODB_CLUSTER_ADMIN_PASSWORD" {
				data[k] = []byte(rootPass)
			}
		}
		err = p.cmd.UpdateSecrets(secretName, data)
		if err != nil {
			return errors.Wrap(err, "update secrets")
		}

		return nil
	}

	data["MONGODB_BACKUP_USER"] = []byte("backup")
	data["MONGODB_BACKUP_PASSWORD"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create backup users pass")
	}
	data["MONGODB_CLUSTER_ADMIN_USER"] = []byte("clusterAdmin")
	data["MONGODB_CLUSTER_ADMIN_PASSWORD"] = []byte(rootPass)
	data["MONGODB_CLUSTER_MONITOR_USER"] = []byte("clusterMonitor")
	data["MONGODB_CLUSTER_MONITOR_PASSWORD"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create cluster monitor users pass")
	}
	data["MONGODB_USER_ADMIN_USER"] = []byte("userAdmin")
	data["MONGODB_USER_ADMIN_PASSWORD"], err = generatePass()
	if err != nil {
		return errors.Wrap(err, "create admin users pass")
	}

	err = p.cmd.CreateSecret(secretName, data)
	if err != nil {
		return errors.Wrap(err, "create secrets")
	}

	return nil
}

const (
	passwordMaxLen = 20
	passwordMinLen = 16
	passSymbols    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789"
)

func generatePass() ([]byte, error) {
	mrand.Seed(time.Now().UnixNano())
	ln := mrand.Intn(passwordMaxLen-passwordMinLen) + passwordMinLen
	b := make([]byte, ln)
	for i := 0; i < ln; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(passSymbols))))
		if err != nil {
			return nil, err
		}
		b[i] = passSymbols[randInt.Int64()]
	}

	return b, nil
}

func (p *PSMDB) PreCheck(name, opts, version string) ([]string, error) {
	err := p.setVersionObjectsWithDefaults(Version(version))
	if err != nil {
		return nil, errors.Wrap(err, "version check")
	}
	supportedVersions := make(map[string]string)
	for v, obj := range objects {
		supportedVersions[string(v)] = obj.psmdb.GetOperatorImage()
	}

	return p.cmd.PreCheck(name, version, p.operatorName(), p.conf.GetOperatorImage(), "psmdb", supportedVersions)
}
