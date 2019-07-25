package pxc

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "default"

	noS3backupWarn = `[Error] S3 backup storage options doesn't set: %v. You have specify S3 storage in order to make backups.
You can skip this step by using --s3-skip-storage flag add the storage later with the "add-storage" command.
`
)

func (p *Controller) DeployPXCCluster(params ProvisionParameters, skipS3Storage *bool) error {
	app := pxc.New(params.ClusterName, defaultVersion)
	conf := p.flags

	var s3stor *dbaas.BackupStorageSpec
	if !*skipS3Storage {
		var err error
		s3stor, err = dbaas.S3Storage(app, conf)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				fmt.Printf(noS3backupWarn, err)
			default:
				fmt.Println("[Error] create S3 backup storage:", err)
			}
			return nil
		}
	}

	setupmsg, err := app.Setup(conf, s3stor)
	if err != nil {
		fmt.Println("[Error] set configuration:", err)
		return nil
	}

	fmt.Println(setupmsg)

	created := make(chan string)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)

	go dbaas.Create("pxc", app, created, msg, cerr)
	sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
	sp.Color("green", "bold")
	sp.Prefix = "Starting..."
	sp.Start()
	defer sp.Stop()
	for {
		select {
		case okmsg := <-created:
			sp.FinalMSG = fmt.Sprintf("Starting...[done]\n%s\n", okmsg)
			return nil
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				// fmt.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				sp.Stop()
				fmt.Printf("[operator log error] %s\n", omsg)

				sp.Start()
			}
		case err := <-cerr:
			sp.Stop()
			switch err.(type) {
			case dbaas.ErrAlreadyExists:
				fmt.Fprintf(os.Stderr, "\n[ERROR] %v\n", err)
				list, err := dbaas.List("pxc")
				if err != nil {
					return nil
				}
				fmt.Println("Avaliable clusters:")
				fmt.Print(list)
			default:
				fmt.Fprintf(os.Stderr, "\n[ERROR] create pxc: %v\n", err)
			}
		}
	}
}

func (p *Controller) DeletePXCCluster(name string) error {
	ok := make(chan string)
	cerr := make(chan error)
	delePVC := true
	go dbaas.Delete("pxc", pxc.New(name, defaultVersion), delePVC, ok, cerr)
	tckr := time.NewTicker(1 * time.Second)
	defer tckr.Stop()
	for {
		select {
		case <-ok:
			log.Println("Deleting...[done]")
			return nil
		case err := <-cerr:
			fmt.Fprintf(os.Stderr, "\n[ERROR] delete pxc: %v\n", err)
			return err
		}
	}
}
