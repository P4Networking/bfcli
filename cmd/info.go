package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/southbound/bfrt"
	"github.com/spf13/cobra"
	"strings"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info TABLE-NAME",
	Short: "Show information about table",
	Long:  `Display the detail of table.`,
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		ret := make([]string, 0)
		if len(args) < 1 {
			argsList, _ := Obj.p4Info.GuessTableName(toComplete)
			for _, v := range argsList {
				if strings.Contains(v, preFixIg) || strings.Contains(v, preFixEg) {
					name := strings.Split(v, ".")
					ret = append(ret, name[len(name)-2]+"."+name[len(name)-1])
				}
			}
			return ret, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		_, _, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		resultTables := make([]bfrt.Table, 0)
		notFounded := make([]string, 0)

		if len(args) == 0 {
			resultTables = Obj.p4Info.Tables
		}

		for _, name := range args {
			found := false
			for _, table := range Obj.p4Info.Tables {
				if strings.Contains(table.Name, name) {
					resultTables = append(resultTables, table)
					found = true
				}
			}
			if !found {
				notFounded = append(notFounded, name)
			}
		}

		for _, v := range resultTables {
			if NotSupportToReadTable[v.ID] {
				continue
			}
			InfoEntries(v)
		}

		for _, v := range notFounded {
			fmt.Println(fmt.Errorf("Couldn't found %s table", v))
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
