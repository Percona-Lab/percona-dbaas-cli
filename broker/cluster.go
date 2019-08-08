package broker

import (
	"log"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "default"
)

func (p *Controller) DeployCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	dbgeneric, err := dbaas.New("")
	if err != nil {
		return err
	}
	switch instance.ServiceID {
	case pxcServiceID:
		app := pxc.New(instance.Parameters.ClusterName, defaultVersion, *dbgeneric)
		conf := dbaas.ClusterConfig{}
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

		var s3stor *dbaas.BackupStorageSpec

		setupmsg, err := app.Setup(conf, s3stor)
		if err != nil {
			log.Println("[Error] set configuration:", err)
			return nil
		}

		log.Println(setupmsg)

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		go dbgeneric.Create("pxc", app, created, msg, cerr)
		go p.listenCreateChannels(created, msg, cerr, instanceID)
	case psmdbServiceID:
		app := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion, *dbgeneric)
		conf := dbaas.ClusterConfig{}
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
		var s3stor *dbaas.BackupStorageSpec

		setupmsg, err := app.Setup(conf, s3stor)
		if err != nil {
			log.Println("[Error] set configuration:", err)
			return nil
		}

		log.Println(setupmsg)

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		go dbgeneric.Create("psmdb", app, created, msg, cerr)
		go p.listenCreateChannels(created, msg, cerr, instanceID)
	}
	return nil
}

func (p *Controller) listenCreateChannels(created chan string, msg chan dbaas.OutuputMsg, cerr chan error, instanceID string) {
	for {
		select {
		case okmsg := <-created:
			p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
			p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
			log.Printf("Starting...[done] %s", okmsg)
			return
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				// fmt.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				p.instanceMap[instanceID].LastOperation.State = FailedOperationState
				p.instanceMap[instanceID].LastOperation.Description = FailedOperationDescription
				log.Printf("[operator log error] %s\n", omsg)
			}
			return
		case err := <-cerr:
			p.instanceMap[instanceID].LastOperation.State = FailedOperationState
			p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
			log.Println("Create error:", err)
			return
		}
	}
}

func (p *Controller) DeleteCluster(instance *ServiceInstance) error {
	ok := make(chan string)
	cerr := make(chan error)
	delePVC := false
	name := instance.Parameters.ClusterName
	dbgeneric, err := dbaas.New("")
	if err != nil {
		return err
	}
	switch instance.ServiceID {
	case pxcServiceID:
		go dbgeneric.Delete("pxc", pxc.New(name, defaultVersion, *dbgeneric), delePVC, ok, cerr)
		p.listenDeleteChannels(ok, cerr)
	case psmdbServiceID:
		go dbgeneric.Delete("psmdb", psmdb.New(name, name, defaultVersion, *dbgeneric), delePVC, ok, cerr)
		p.listenDeleteChannels(ok, cerr)
	}
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
