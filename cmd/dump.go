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
	"log"
	"strings"
)


// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump TABLE-NAME",
	Short: "Dump the existed entries in the specific table",
	Long:  `Display all existed entries in the specific table`,
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		argsList, _ := p4Info.GuessTableName(toComplete)

		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		var tableName string

		switch all {
		case false:
			tableId := p4Info.SearchTableId(tableName)
			if tableId == util.ID_NOT_FOUND {
				fmt.Printf("Can not found table with name: %s\n", tableName)
				return
			}
			table := p4Info.SearchTableById(tableId)
			if table == nil {
				fmt.Printf("Can not found table with ID: %s\n", tableId)
				return
			}

			stream, err := cli.Read(ctx, GenReadRequestWithId(table.ID))
			if err != nil {
				log.Fatalf("Got error, %v \n", err.Error())
			}

			dumpEntries(&stream, table)
		case true:
			for _, v := range p4Info.Tables {
				if strings.HasPrefix(v.Name, preIg) || strings.HasPrefix(v.Name, preEg) {
					table := p4Info.SearchTableById(v.ID)
					if table == nil {
						fmt.Printf("Can not found table with ID: %s\n", v.ID)
						return
					}
					stream, err := cli.Read(ctx, GenReadRequestWithId(v.ID))
					if err != nil {
						log.Fatalf("Got error, %v \n", err.Error())
					}

					dumpEntries(&stream, table)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolVarP(&all, "all", "a", false, "dump all of the tables")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
