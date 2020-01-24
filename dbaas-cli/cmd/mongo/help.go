package mongo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help <command_name>",
	Short: "Show help for given command. Usage: 'percona-dbaas mongodb help [command]'",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("you have to specify command name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		a := os.Args[0]
		c := exec.Command(a, MongoCmd.Name(), args[0], "--help")
		o, err := c.Output()
		if err != nil {
			return
		}
		fmt.Println(string(o))
	},
}

func init() {
	MongoCmd.AddCommand(helpCmd)
}
