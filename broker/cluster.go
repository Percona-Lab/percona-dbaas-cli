package broker

import (
	"log"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "default"

	noS3backupWarn = `[Error] S3 backup storage options doesn't set: %v. You have specify S3 storage in order to make backups.
You can skip this step by using --s3-skip-storage flag add the storage later with the "add-storage" command.
`
)

func (p *Controller) DeployCluster(instance ServiceInstance, skipS3Storage *bool, instanceID string) error {
	switch instance.ServiceID {
	case pxcServiceID:
		app := pxc.New(instance.Parameters.ClusterName, defaultVersion)
		conf := dbaas.ClusterConfig{}
		SetDefault(&conf)

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

		go dbaas.Create("pxc", app, created, msg, cerr)

		for {
			select {
			case okmsg := <-created:
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				log.Printf("Starting...[done] %s", okmsg)
				return nil
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					p.instanceMap[instanceID].LastOperation.State = FailedOperationState
					p.instanceMap[instanceID].LastOperation.Description = FailedOperationDescription
					log.Printf("[operator log error] %s\n", omsg)
				}
			case err := <-cerr:
				switch err.(type) {
				case dbaas.ErrAlreadyExists:
					p.instanceMap[instanceID].LastOperation.State = FailedOperationState
					p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
					log.Printf("[ERROR] %v", err)
					list, err := dbaas.List("pxc")
					if err != nil {
						return nil
					}
					log.Println("Avaliable clusters:")
					log.Print(list)
				default:
					log.Printf("[ERROR] create pxc: %v", err)
				}
			}
		}
	case psmdbServiseID:
		app := psmdb.New(instance.Parameters.ClusterName, instance.Parameters.ClusterName, defaultVersion)
		conf := dbaas.ClusterConfig{}
		SetDefault(&conf)

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

		go dbaas.Create("psmdb", app, created, msg, cerr)

		for {
			select {
			case okmsg := <-created:
				p.instanceMap[instanceID].LastOperation.State = SucceedOperationState
				p.instanceMap[instanceID].LastOperation.Description = SucceedOperationDescription
				log.Printf("Starting...[done] %s", okmsg)
				return nil
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					p.instanceMap[instanceID].LastOperation.State = FailedOperationState
					p.instanceMap[instanceID].LastOperation.Description = FailedOperationDescription
					log.Printf("[operator log error] %s\n", omsg)
				}
			case err := <-cerr:
				switch err.(type) {
				case dbaas.ErrAlreadyExists:
					p.instanceMap[instanceID].LastOperation.State = FailedOperationState
					p.instanceMap[instanceID].LastOperation.Description = InProgressOperationDescription
					log.Printf("[ERROR] %v", err)
					list, err := dbaas.List("pxc")
					if err != nil {
						return nil
					}
					log.Println("Avaliable clusters:")
					log.Print(list)
				default:
					log.Printf("[ERROR] create psmdb: %v", err)
				}
			}
		}
	}
	return nil
}

func (p *Controller) DeletePXCCluster(instance *ServiceInstance) error {
	ok := make(chan string)
	cerr := make(chan error)
	delePVC := true
	name := instance.Parameters.ClusterName
	switch instance.ServiceID {
	case pxcServiceID:
		go dbaas.Delete("pxc", pxc.New(name, defaultVersion), delePVC, ok, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case <-ok:
				log.Println("Deleting...[done]")
				return nil
			case err := <-cerr:
				log.Printf("[ERROR] delete pxc: %v", err)
				return err
			}
		}
	case psmdbServiseID:
		go dbaas.Delete("psmdb", pxc.New(name, defaultVersion), delePVC, ok, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case <-ok:
				log.Println("Deleting...[done]")
				return nil
			case err := <-cerr:
				log.Printf("[ERROR] delete pxc: %v", err)
				return err
			}
		}
	}
	return nil
}
