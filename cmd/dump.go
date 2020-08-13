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
	"reflect"
	"strings"
)

var endFlag bool = false

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump TABLE-NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Dump the existed flows in specify table",
	Long:  `Display all existed flows in specify table`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		argsList, _ := p4Info.GuessTableName(toComplete)

		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		tableId := p4Info.SearchTableId(tableName)
		if tableId == util.ID_NOT_FOUND {
			tableList, only := p4Info.GuessTableName(tableName)
			if only {
				tableName = tableList[0]
			} else {
				fmt.Printf("Can not found table with name: %s\n", tableName)
				return
			}
		}

		req := &p4.ReadRequest{
			Entities: []*p4.Entity{
				{
					Entity: &p4.Entity_TableEntry{
						TableEntry: &p4.TableEntry{
							TableId: p4Info.SearchTableId(tableName),
						},
					},
				},
			},
		}

		stream, err := cli.Read(ctx, req)
		if err != nil {
			log.Fatalf("Got error, %v \n", err.Error())
		}

		for {
			rsp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Got error: %v", err)
			}
			if len(rsp.GetEntities()) == 0 {
				fmt.Printf("The flows in %s is null\n", tableName)
			}
			fmt.Println("--------------------------------------------------------------------------------")
			for _, v := range rsp.Entities {
				tbl := v.GetTableEntry()
				fmt.Println("Match Key Info")
				if tbl.GetKey() != nil {
					fmt.Printf("  %-20s %-10s %-16s\n", "Field Name:", "Type:", "Value:")
					for k, f := range tbl.Key.Fields {
						if f.FieldId == 65537 {
							continue
						}
						//fmt.Println(p4Info.SearchTableById(tbl.TableId).Name)
						MatchfieldName := p4Info.SearchTableById(tbl.TableId).Key[k].Name
						switch strings.Split(reflect.TypeOf(f.GetMatchType()).String(), ".")[1] {
						case "KeyField_Exact_":
							m := f.GetExact()
							fmt.Printf("  %-20s %-10s %-16d\n", MatchfieldName, "Exact" ,m.Value)
						case "KeyField_Ternary_":
							t := f.GetTernary()
							fmt.Printf("  %-20s %-10s %-16x Mask: %-12x\n", MatchfieldName, "Ternay" ,t.Value, t.Mask)
						case "KeyField_Lpm":
							l := f.GetLpm()
							fmt.Printf("  %-20s %-10s %-16x PreFix: %-12x\n", MatchfieldName, "LPM" ,l.Value, l.PrefixLen)
						case "KeyField_Range_":
							//It will be adjust when range match implement.
							r := f.GetRange()
							fmt.Printf("  %-20s %-10s %-16x High: %-6x Low: %-6x\n", MatchfieldName, "LPM" ,r.High, r.Low)
						}
					}
				} else {
					fmt.Printf("Table default action:\n")
					endFlag = true
				}

				actionName, _ := printNameById(tbl.Data.ActionId)
				fmt.Println("Action:", actionName)

				if tbl.Data.Fields != nil {
					fmt.Printf("  %-10s %-16s\n", "Field:", "Value:")
					for _, d := range tbl.Data.Fields {
						actionFieldName, _ := printNameById(d.FieldId)
						fmt.Printf("  %-10s %-16x\n",actionFieldName, d.GetStream())
					}
				}
				if !endFlag {
					fmt.Printf("------------------\n")
				}
			}
			fmt.Println("--------------------------------------------------------------------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
