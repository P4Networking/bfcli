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
	"github.com/P4Networking/proto/go/p4"
	"io"
	"log"

	"github.com/spf13/cobra"
)

var(
	all bool
)
// delFlowCmd represents the delFlow command
var delFlowCmd = &cobra.Command{
	Use:   "del-flow",
	Short: "Remove a entry from a table",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]
		//actionName := args [1]

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		tableId := p4Info.SearchTableId(tableName)
		if tableId == util.ID_NOT_FOUND {
			fmt.Printf("Can not find table ID with name: %s\n", tableName)
			return
		}
		req := &p4.ReadRequest{
			Entities: []*p4.Entity{
				{
					Entity: &p4.Entity_TableEntry{
						TableEntry: &p4.TableEntry{
							TableId: tableId,
						},
					},
				},
			},
		}

		stream, err := cli.Read(ctx, req)
		if err != nil {
			log.Fatalf("Got error, %v \n", err.Error())
			return
		}

		for {
			rsp, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if rsp != nil {
				if len(rsp.GetEntities()) == 0 {
					fmt.Printf("The flows in %s is null\n", tableName)
					break
				}
				if all {
					for _, e := range rsp.Entities {
						tbl := e.GetTableEntry()
						if tbl.GetKey() != nil {
							delReq := util.GenWriteRequestWithId(p4.Update_DELETE, tableId, tbl.Key.Fields, nil)
							delReq.Updates[0].GetEntity().GetTableEntry().Data = nil
							_, err := cli.Write(ctx, delReq)
							if err != nil {
								fmt.Printf("Failed to clear table: %s with entity: %s\n", tableName, e.String())
								fmt.Println(err)
							}
						} else {
							fmt.Println("Get Key Field Error, Please Check the table that has KeyFields.")
							break
						}
					}
					fmt.Printf("Remove all entries in %s talbe.\n", tableName)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(delFlowCmd)
	delFlowCmd.Flags().BoolVarP(&all, "all", "a", false, "Delete all entries")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// delFlowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// delFlowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
