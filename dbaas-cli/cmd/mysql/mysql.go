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

	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/dp"
)

var dotPrinter dp.DotPrinter

// PXCCmd represents the mysql command
var PXCCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Manage your MySQL cluster on Kubernetes",
}

func parseArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	if a := strings.Split(args[0], "/"); len(a) == 2 {
		args = a
	}

	return args
}

func init() {
	dotPrinter = dp.New()
}
