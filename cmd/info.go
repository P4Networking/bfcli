/*
Copyright © 2020 Chun Ming Ou <breezestars@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info TABLE-NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Show information about table",
	Long:  `Display the detail of table.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		argsList, _ := p4Info.GuessTableName(toComplete)

		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Printf("Got cmd: %s | and args: %s\n", cmd.Name(), args)
		var tableName string

		_, _, conn, cancel, p4Info, nonP4Info := initConfigClient()
		defer conn.Close()
		defer cancel()

		// Guest table name via table name provide form user
		tableList, ok := p4Info.GuessTableName(args[0])
		if !ok {
			tableList, ok = nonP4Info.GuessTableName(args[0])
			if !ok {
				fmt.Printf("Not found the table %s\n", args[0])
				return
			}
		}
		tableName = tableList[0]

		tableId := p4Info.SearchTableId(tableName)
		if tableId == util.ID_NOT_FOUND {
			fmt.Printf("Can not found table with name: %s\n", tableName)
			return
		}

		table := p4Info.SearchTableById(tableId)

		if table.Name != "" {
			fmt.Printf("%-12s: %-6s\n", "Table Name", table.Name)
		}
		if table.ID != 0 {
			fmt.Printf("%-12s: %-6d\n", "Table ID", table.ID)
		}
		if table.TableType != "" {
			fmt.Printf("%-12s: %-6s\n", "Table Type", table.TableType)
		}
		if table.Size != 0 {
			fmt.Printf("%-12s: %-6d\n", "Table Size", table.Size)
		}
		if table.Annotations != nil {
			fmt.Printf("%-12s:\n", "Table Annotations")
			for k, v := range table.Annotations {
				fmt.Printf("%d - Name: %s | Value: %s \n", k+1, v.Name, v.Value)
			}
		}
		if table.DependsOn != nil {
			fmt.Printf("%-12s: %-6s\n", "Table DependsOn", table.DependsOn)
		}

		fmt.Println("==================================================================================================================================")
		fmt.Printf("%-12s:\n", "Table Key")
		for _, v := range table.Key {
			fmt.Printf("KeyId: %-6d, Name: %-20s, Match Type: %-10s, Mandatory: %-6t, Repeated: %-6t, Type: %-8s, Width: %-4d\n", v.ID, v.Name, v.MatchType, v.Mandatory, v.Repeated, v.Type.Type, v.Type.Width)
		}

		if table.Data != nil {
			fmt.Println("==================================================================================================================================")
			fmt.Printf("%-12s:\n", "Table Data")
			for _, v := range table.Data {
				fmt.Printf("KeyId: %-6d, Name: %-20s, Mandatory: %-6t, Repeated: %-6t, Type: %-8s\n", v.Singleton.ID, v.Singleton.Name, v.Mandatory, v.Singleton.Repeated, v.Singleton.Type.Type)
			}
		}

		if table.ActionSpecs != nil {
			fmt.Println("==================================================================================================================================")
			fmt.Printf("%-12s:\n", "Action Specs")
			for _, v := range table.ActionSpecs {
				fmt.Printf("Id: %-6d, Name: %-20s \n", v.ID, v.Name)
				for _, d := range v.Data {
					fmt.Printf("ParameterId: %-6d, Name: %-20s, Mandatory: %-6t, Repeated: %-6t, Type: %-8s, Width: %-4d\n", d.ID, d.Name, d.Mandatory, d.Repeated, d.Type.Type, d.Type.Width)
				}

			}
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
