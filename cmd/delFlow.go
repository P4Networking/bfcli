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
	"github.com/spf13/cobra"
	"io"
	"log"
	"strconv"
	"strings"
)

var(
	reset bool
	delEntryNumber []string
	preEg = "pipe.SwitchEgress."
	preIg = "pipe.SwitchIngress."
)
// delFlowCmd represents the delFlow command
var delFlowCmd = &cobra.Command{
	Use: "del-flow TABLE-NAME ",
	Short: "Remove a entry from a table",
	Long: `del-flow can remove all of the entries from specific table`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var tableName string
		if !all && len(args) > 0 {
			tableName = args[0]
		}

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		if reset {
			for _, v := range p4Info.Tables {
				if strings.HasPrefix(v.Name, preIg) || strings.HasPrefix(v.Name, preEg){

					stream, err := cli.Read(ctx, GenReadRequestWithId(v.ID))
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
								fmt.Printf("%s table is empty\n", v.Name)
							} else {
								for _, e := range rsp.Entities {
									tbl := e.GetTableEntry()
									if tbl.GetKey() != nil {
										delReq := util.GenWriteRequestWithId(p4.Update_DELETE, v.ID, tbl.Key.Fields, nil)
										delReq.Updates[0].GetEntity().GetTableEntry().Data = nil
										_, err := cli.Write(ctx, delReq)
										if err != nil {
											fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", tableName, tbl.Key.Fields)
											fmt.Println(err)
											return
										}
										fmt.Printf("%s table has cleared.\n", v.Name)
									}
								}
							}
						}
					}
				}
			}
			fmt.Println("Reset complete.")
			return
		}

		tableId := p4Info.SearchTableId(tableName)
		if tableId == util.ID_NOT_FOUND {
			fmt.Printf("Can not find table ID with name: %s\n", tableName)
			return
		}

		stream, err := cli.Read(ctx, GenReadRequestWithId(tableId))
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
					fmt.Printf("%s table is empty\n", tableName)
					break
				}
				if all {
					for _, e := range rsp.Entities {
						tbl := e.GetTableEntry()
						if tbl != nil {
							if tbl.GetKey() != nil {
								delReq := util.GenWriteRequestWithId(p4.Update_DELETE, tableId, tbl.Key.Fields, nil)
								delReq.Updates[0].GetEntity().GetTableEntry().Data = nil
								_, err := cli.Write(ctx, delReq)
								if err != nil {
									fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", tableName, tbl.Key.Fields)
									fmt.Println(err)
									return
								}
							} else if tbl.IsDefaultEntry {
								continue
							}else {
								fmt.Printf("Got key fields error:\n  %s\n",tbl.String())
								return
							}
						} else {
							fmt.Printf("The \"%s\" table is empty, it doesn't have any entries to delete.", tableName)
							return
						}
					}
					fmt.Printf("All entries of \"%s\" table are removed.\n", tableName)
					return
				} else if len(delEntryNumber) > 0 {
					for k, e := range rsp.Entities {
						for i:=0; i < len(delEntryNumber); i++ {
							if n, _ := strconv.Atoi(delEntryNumber[i]); k == n {
								tbl := e.GetTableEntry()
								if tbl.GetKey() != nil {
									delReq := util.GenWriteRequestWithId(p4.Update_DELETE, tableId, tbl.Key.Fields, nil)
									delReq.Updates[0].GetEntity().GetTableEntry().Data = nil
									_, err := cli.Write(ctx, delReq)
									if err != nil {
										fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", tableName, tbl.Key.Fields)
										fmt.Println(err)
										return
									}
									fmt.Printf("Entry number %d is deleted.\nKeyFields: %s\n", k, tbl.Key.Fields)
								}

							} else {
								fmt.Printf("Entry number %d not founded", n)
							}
						}
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(delFlowCmd)
	delFlowCmd.Flags().BoolVarP(&all, "all", "a", false, "delete all entries of the table")
	delFlowCmd.Flags().StringSliceVarP(&delEntryNumber, "entry", "e", []string{}, "delete specific entry by given entry number")
	delFlowCmd.Flags().BoolVarP(&reset, "reset", "r", false, "clear all of the tables")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// delFlowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// delFlowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
