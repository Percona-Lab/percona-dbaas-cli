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

package mysql

import (
	"strings"

	op "github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/output"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/pb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	dotPrinter pb.ProgressBar
	noWait     bool
	maxTries   = 1200
)

// PXCCmd represents the mysql command
var PXCCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Manage your MySQL instance",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Error(errors.Wrap(err, "get output flag value"))
			return
		}
		dotPrinter = op.GetDotprinter(output)
		log.SetFormatter(op.GetFormatter(output))

		noWait, err = cmd.Flags().GetBool("no-wait")
		if err != nil {
			log.Error(errors.Wrap(err, "get no-wait flag"))
			return
		}
	},
}

func addSpec(opts string) string {
	if len(opts) == 0 {
		return ""
	}
	return "spec." + strings.Replace(opts, ",", ",spec.", -1)
}
