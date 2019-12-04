package pxc

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
	"github.com/pkg/errors"
)

var currectOptions map[string]string

func (p *PXC) ParseOptions(options string) error {
	var c config.ClusterConfig

	res := config.PodResources{
		Requests: config.ResourcesList{
			CPU:    "600m",
			Memory: "1G",
		},
	}
	topologyKey := "kubernetes.io/hostname"
	aff := config.PodAffinity{
		TopologyKey: topologyKey,
	}
	c.PXC.Size = int32(3)
	c.PXC.Resources = res
	c.PXC.Affinity = aff
	c.ProxySQL.Size = int32(1)
	c.ProxySQL.Resources = res
	c.ProxySQL.Affinity = aff
	c.S3.SkipStorage = true

	if len(options) != 0 {
		currectOptions = make(map[string]string)
		keys(reflect.TypeOf(config.ClusterConfig{}), "", "")

		optArr := strings.Split(options, ",")

		for _, str := range optArr {
			v := strings.Split(str, "=")
			if _, ok := currectOptions[v[0]]; !ok {
				return errors.New("incorrect options")
			}
			if len(v) > 1 {
				err := set(&c, currectOptions[v[0]], v[1])
				if err != nil {
					return errors.Wrap(err, "set value")
				}
			}
		}
	}
	p.config = c

	return nil
}

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
		db.Host = st.Status.Host
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
func (p *PXC) DeleteDBCluster(name string, delePVC bool) error {
	ext, err := p.cmd.IsObjExists("pxc", name)
	if err != nil {
		return errors.Wrap(err, "check if cluster exists")
	}

	if !ext {
		return errors.New("unable to find cluster pxc/" + name)
	}

	err = p.cmd.DeleteCluster("pxc", p.operatorName(), name, delePVC)
	if err != nil {
		return errors.Wrap(err, "delete cluster")
	}
	return nil
}

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

	db.Host = st.Status.Host
	db.Port = 3306
	db.User = "root"
	db.Pass = string(secrets["root"])
	db.Status = string(st.Status.Status)

	return db, nil
}

func (p *PXC) UpdateDBCluster() error {
	return nil
}

func (p *PXC) ListDBClusters() error {
	return nil
}

func (p *PXC) DescribeDBCluster(name string) error {
	return nil
}

func set(i *config.ClusterConfig, field string, value string) error {
	fs := strings.Split(field, ".")
	v := reflect.ValueOf(i).Elem()

	switch len(fs) {
	case 1:
		val, err := getValue(value, v.FieldByName(fs[0]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).Set(val)
	case 2:
		val, err := getValue(value, v.FieldByName(fs[0]).FieldByName(fs[1]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).Set(val)
	case 3:
		val, err := getValue(value, v.FieldByName(fs[0]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).Set(val)
	case 4:
		val, err := getValue(value, v.FieldByName(fs[0]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).FieldByName(fs[3]).Set(val)
	case 5:
		val, err := getValue(value, v.FieldByName(fs[0]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).FieldByName(fs[3]).FieldByName(fs[4]).Set(val)
	}
	return nil
}

func getValue(value string, field reflect.Value) (reflect.Value, error) {
	switch field.Type().Name() {
	case "int":
		v, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.Indirect(reflect.ValueOf(v)), nil
	case "int32":
		v, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.Indirect(reflect.ValueOf(int32(v))), nil
	case "int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.Indirect(reflect.ValueOf(v)), nil
	case "float32":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.Indirect(reflect.ValueOf(float32(v))), nil
	case "float64":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.Indirect(reflect.ValueOf(v)), nil
	case "bool":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.Indirect(reflect.ValueOf(v)), nil
	default:
		return reflect.Indirect(reflect.ValueOf(value)), nil
	}
}

func keys(t reflect.Type, prevType, prevName string) map[string]string {
	var v = make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		name := strings.TrimSpace(strings.Split(t.Field(i).Tag.Get("json"), ",")[0])
		if t.Field(i).Type.Kind() == reflect.Struct {
			for name, nType := range keys(t.Field(i).Type, prevType+t.Field(i).Name+".", prevName+name+".") {
				currectOptions[prevName+name+"."+name] = prevType + t.Field(i).Name + "." + nType
			}
		} else {
			v[name] = t.Field(i).Name
			currectOptions[prevName+name] = prevType + t.Field(i).Name
		}
	}

	return v
}
