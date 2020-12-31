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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		var argsList []string
		for _, v := range p4Info.Tables {
			if strings.Contains(v.Name, preFixIg) || strings.Contains(v.Name, preFixEg) {
				argsList = append(argsList, v.Name)
			}
		}
		return argsList, cobra.ShellCompDirectiveNoFileComp
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
