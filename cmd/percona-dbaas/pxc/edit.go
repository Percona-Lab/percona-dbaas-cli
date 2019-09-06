// Copyright © 2019 Percona, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pxc

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "modify-db <pxc-cluster-name>",
	Short: "Modify MySQL cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := pxc.New(name, defaultVersion, *editAnswerInJSON, "", *envEdt)
		if err != nil {
			if *addStorageAnswerInJSON {
				fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("new operator", err))
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			return
		}

		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		demo, err := cmd.Flags().GetBool("demo")
		if demo && err == nil {
			sp.UpdateCharSet([]string{""})
		}
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		ext, err := app.Cmd.IsObjExists("pxc", name)
		if err != nil {
			if *editAnswerInJSON {
				fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("check if cluster exists", err))
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] check if cluster exists: %v\n", err)
			return
		}

		if !ext {
			sp.Stop()
			if *editAnswerInJSON {
				fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("Unable to find cluster pxc/"+name, nil))
			} else {
				fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "pxc", name)
			}
			list, err := app.Cmd.List("pxc")
			if err != nil {
				if *editAnswerInJSON {
					fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("pxc clusters list", err))
					return
				}
				return
			}
			fmt.Fprint(os.Stderr, "Avaliable clusters:")
			fmt.Print(list)
			return
		}

		config, err := pxc.ParseEditFlagsToConfig(cmd.Flags())
		if err != nil {
			if *editAnswerInJSON {
				fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("parse flags to config", err))
				return
			}
			fmt.Fprint(os.Stderr, "[Error] parse flags to config:", err)
			return
		}
		app.ClusterConfig = config

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		go app.Edit(nil, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Applying changes..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("pxc", name)
				sp.FinalMSG = fmt.Sprintf("Applying changes...[done]\n\n%s", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					if *editAnswerInJSON {
						fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("operator log error", err))
					} else {
						fmt.Printf("[operator log error] %s\n", omsg)
					}
					sp.Start()
				}
			case err := <-cerr:
				if *editAnswerInJSON {
					fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("edit pxc", err))
					return
				}
				fmt.Fprintf(os.Stderr, "\n[ERROR] edit pxc: %v\n", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envEdt *string
var editAnswerInJSON *bool

func init() {
	editCmd.Flags().Int32("pxc-instances", 0, "Number of PXC nodes in cluster")
	editCmd.Flags().Int32("proxy-instances", -1, "Number of ProxySQL nodes in cluster")
	envEdt = editCmd.Flags().String("environment", "", "Target kubernetes cluster")
	editAnswerInJSON = editCmd.Flags().Bool("json", false, "Answers in JSON format")

	PXCCmd.AddCommand(editCmd)
}
