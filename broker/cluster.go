package broker

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "default"
)

func (p *Controller) DeployPXCCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	dbservice, err := p.createDbservice(&instance)
	if err != nil {
		return err
	}

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return errors.Wrap(err, "marshal instance")
	}

	app := pxc.New(instance.Parameters.ClusterName, defaultVersion, "")
	conf := pxc.ClusterConfig{}
	SetPXCDefaults(&conf)
	if instance.Parameters.Replicas > int32(0) {
		conf.PXC.Instances = instance.Parameters.Replicas
	}
	if len(instance.Parameters.Size) > 0 {
		conf.PXC.StorageSize = instance.Parameters.Size
	}
	if len(instance.Parameters.TopologyKey) > 0 {
		conf.PXC.AntiAffinityKey = instance.Parameters.TopologyKey
	}
	conf.PXC.BrokerInstance = string(brokerInstance)

	if instance.Parameters.PMM.Enabled {
		conf.PMM.Image = instance.Parameters.PMM.Image
		conf.PMM.ServerHost = instance.Parameters.PMM.Host
		conf.PMM.ServerUser = instance.Parameters.PMM.User
		conf.PMM.ServerPass = instance.Parameters.PMM.Pass
		conf.PMM.Enabled = instance.Parameters.PMM.Enabled
	}

	app.ClusterConfig = conf

	var s3stor *dbaas.BackupStorageSpec

	setupmsg, err := app.Setup(conf, s3stor, dbservice.GetPlatformType())
	if err != nil {
		log.Println("[Error] set configuration:", err)
		return errors.Wrap(err, "set configuration")
	}

	log.Println(setupmsg)

	created := make(chan dbaas.Msg)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)
	go dbservice.Create("pxc", app, instance.Parameters.OperatorImage, created, msg, cerr)
	go p.listenCreateChannels(created, msg, cerr, instanceID, "pxc", dbservice)

	return nil
}

func (p *Controller) DeployPSMDBCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	dbservice, err := p.createDbservice(&instance)
	if err != nil {
		return errors.Wrap(err, "create dbservice")
	}

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return errors.Wrap(err, "marshal instance")
	}

	app := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, "")
	conf := psmdb.ClusterConfig{}
	SetPSMDBDefaults(&conf)
	if instance.Parameters.Replicas > int32(0) {
		conf.PSMDB.Instances = instance.Parameters.Replicas
	}
	if len(instance.Parameters.Size) > 0 {
		conf.PSMDB.StorageSize = instance.Parameters.Size
	}
	if len(instance.Parameters.TopologyKey) > 0 {
		conf.PSMDB.AntiAffinityKey = instance.Parameters.TopologyKey
	}
	conf.PSMDB.BrokerInstance = string(brokerInstance)
	app.ClusterConfig = conf

	var s3stor *dbaas.BackupStorageSpec

	setupmsg, err := app.Setup(s3stor, dbservice.GetPlatformType())
	if err != nil {
		log.Println("[Error] set configuration:", err)
		return errors.Wrap(err, "set configuration")
	}

	log.Println(setupmsg)

	created := make(chan dbaas.Msg)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)
	go dbservice.Create("psmdb", app, "", created, msg, cerr)
	go p.listenCreateChannels(created, msg, cerr, instanceID, "psmdb", dbservice)

	return nil
}

func (p *Controller) getClusterSecret(clusterName string) (Secret, error) {
	var secret Secret
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return secret, errors.Wrap(err, "create dbservice")
	}

	s, err := dbservice.GetObject("secret", clusterName+"-secrets")
	if err != nil {
		return secret, errors.Wrap(err, "get secret")
	}
	err = json.Unmarshal(s, &secret)

	return secret, err
}

func (p *Controller) listenCreateChannels(created chan dbaas.Msg, msg chan dbaas.OutuputMsg, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
	for {
		select {
		case okmsg := <-created:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				p.instanceMap[instanceID].CredentialData = okmsg
				instance, err := json.Marshal(p.instanceMap[instanceID])
				if err != nil {
					log.Println("Error marshal instance", err)
				}
				if len(instance) > 0 {
					dbservice.Annotate(resource, p.instanceMap[instanceID].Parameters.ClusterName, "broker-instance", string(instance))
				}
			}
			log.Println(okmsg)
			return
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				//log.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				log.Printf("[operator log error] %s\n", omsg)
			}
		case err := <-cerr:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = FailedOperationState
				p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			}
			log.Println("Create error:", err)
			return
		}
	}
}

func (p *Controller) DeletePXCCluster(instance *ServiceInstance) error {
	ok := make(chan string)
	cerr := make(chan error)
	delePVC := false
	name := instance.Parameters.ClusterName
	dbservice, err := p.createDbservice(instance)
	if err != nil {
		return errors.Wrap(err, "create dbservice")
	}

	go dbservice.Delete("pxc", pxc.New(name, defaultVersion, ""), delePVC, ok, cerr)
	p.listenDeleteChannels(ok, cerr)

	return nil
}

func (p *Controller) DeletePSMDBCluster(instance *ServiceInstance) error {
	ok := make(chan string)
	cerr := make(chan error)
	delePVC := false
	name := instance.Parameters.ClusterName
	dbservice, err := p.createDbservice(instance)
	if err != nil {
		return errors.Wrap(err, "create dbservice")
	}

	go dbservice.Delete("psmdb", psmdb.New(name, name, defaultVersion, ""), delePVC, ok, cerr)
	p.listenDeleteChannels(ok, cerr)

	return nil
}

func (p *Controller) listenDeleteChannels(ok chan string, cerr chan error) {
	for {
		select {
		case <-ok:
			log.Println("Deleting...[done]")
			return
		case err := <-cerr:
			log.Printf("[ERROR] delete pxc: %v", err)
			return
		}
	}
}

func (p *Controller) UpdatePXCCluster(instance *ServiceInstance) error {
	dbservice, err := p.createDbservice(instance)
	if err != nil {
		return errors.Wrap(err, "create dbservice")
	}

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return errors.Wrap(err, "marshal instance")
	}

	created := make(chan dbaas.Msg)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)

	app := pxc.New(instance.Parameters.ClusterName, defaultVersion, "")
	conf := pxc.ClusterConfig{}
	SetPXCDefaults(&conf)
	if instance.Parameters.Replicas > int32(0) {
		conf.PXC.Instances = instance.Parameters.Replicas
	}

	conf.PXC.BrokerInstance = string(brokerInstance)
	app.ClusterConfig = conf

	go dbservice.Edit("pxc", app, nil, created, msg, cerr)
	go p.listenUpdateChannels(created, msg, cerr, instance.ID, "pxc", p.dbaas)

	return nil
}

func (p *Controller) UpdatePSMDBCluster(instance *ServiceInstance) error {
	dbservice, err := p.createDbservice(instance)
	if err != nil {
		return errors.Wrap(err, "create dbservice")
	}

	brokerInstance, err := json.Marshal(instance)
	if err != nil {
		return errors.Wrap(err, "marshal instance")
	}

	created := make(chan dbaas.Msg)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)

	app := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, "")
	conf := psmdb.ClusterConfig{}
	SetPSMDBDefaults(&conf)
	if instance.Parameters.Replicas > int32(0) {
		conf.PSMDB.Instances = instance.Parameters.Replicas
	}
	conf.PSMDB.BrokerInstance = string(brokerInstance)
	app.ClusterConfig = conf

	go dbservice.Edit("psmdb", app, nil, created, msg, cerr)
	go p.listenUpdateChannels(created, msg, cerr, instance.ID, "psmdb", p.dbaas)

	return nil
}

func (p *Controller) listenUpdateChannels(created chan dbaas.Msg, msg chan dbaas.OutuputMsg, cerr chan error, instanceID, resource string, dbservice *dbaas.Cmd) {
	for {
		select {
		case okmsg := <-created:
			if _, ok := p.instanceMap[instanceID]; ok {
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				instance, err := json.Marshal(p.instanceMap[instanceID])
				if err != nil {
					log.Println("Error marshal oinstance", err)
				}
				if len(instance) > 0 {
					dbservice.Annotate(resource, p.instanceMap[instanceID].Parameters.ClusterName, "broker-instance", string(instance))
				}
			}
			log.Printf(okmsg.String())
			return
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				// fmt.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				log.Printf("[operator log error] %s\n", omsg)
			}
		case err := <-cerr:
			p.instanceMap[instanceID].LastOperation.State = FailedOperationState
			p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			log.Println("Create error:", err)
			return
		}
	}
}

func (p *Controller) createDbservice(instance *ServiceInstance) (*dbaas.Cmd, error) {
	dbservice, err := dbaas.New(p.EnvName)
	if err != nil {
		return nil, errors.Wrap(err, "create dbservice")
	}
	dbservice.Namespace = instance.Context.Namespace

	return dbservice, nil
}
