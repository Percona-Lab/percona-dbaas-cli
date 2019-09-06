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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

var delePVC *bool

// delCmd represents the list command
var delCmd = &cobra.Command{
	Use:   "delete-db <pxc-cluster-name>",
	Short: "Delete MySQL cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		app, err := pxc.New(args[0], defaultVersion, *createAnswerInJSON, *labels, *envDlt)
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
			if *deleteAnswerInJSON {
				fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("check if cluster exists", err))
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] check if cluster exists: %v\n", err)
			return
		}

		if !ext {
			sp.Stop()
			if *deleteAnswerInJSON {
				fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("Unable to find cluster pxc/"+name, nil))
			} else {
				fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "pxc", name)
			}
			list, err := app.Cmd.List("pxc")
			if err != nil {
				if *deleteAnswerInJSON {
					fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("pxc cluster list", err))
					return
				}
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		if *delePVC {
			sp.Stop()
			var yn string
			fmt.Printf("\nAll current data on \"%s\" cluster will be destroyed.\nAre you sure? [y/N] ", name)
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				yn = strings.TrimSpace(scanner.Text())
				break
			}
			if yn != "y" && yn != "Y" {
				return
			}
			sp.Start()
		}
		sp.Lock()
		sp.Prefix = "Deleting..."
		sp.Unlock()
		ok := make(chan string)
		cerr := make(chan error)

		go app.Delete(*delePVC, ok, cerr)

		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case <-ok:
				sp.FinalMSG = "Deleting...[done]\n"
				return
			case err := <-cerr:
				if *deleteAnswerInJSON {
					fmt.Fprint(os.Stderr, pxc.JSONErrorMsg("delete pxc", err))
					return
				}
				fmt.Fprintf(os.Stderr, "\n[ERROR] delete pxc: %v\n", err)
				return
			}
		}
	},
}

var envDlt *string
var deleteAnswerInJSON *bool

func init() {
	delePVC = delCmd.Flags().Bool("clear-data", false, "Remove cluster volumes")
	envDlt = delCmd.Flags().String("environment", "", "Target kubernetes cluster")
	deleteAnswerInJSON = delCmd.Flags().Bool("json", false, "Answers in JSON format")

	PXCCmd.AddCommand(delCmd)
}
