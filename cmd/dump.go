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
	bfrt "github.com/breezestars/go-bfrt/proto/out"
	"github.com/breezestars/go-bfrt/util"
	"github.com/spf13/cobra"
	"io"
	"log"
	"reflect"
	"strings"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump TABLE-NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Dump the existed flows in specify table",
	Long:  `Display all existed flows in specify table`,
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

		cliAddr, ctxAddr, conn, cancel, p4, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		tableId := p4.SearchTableId(tableName)
		if tableId == util.ID_NOT_FOUND {
			fmt.Printf("Can not found table with name: %s\n", tableName)
			return
		}

		req := &bfrt.ReadRequest{
			Entities: []*bfrt.Entity{
				{
					Entity: &bfrt.Entity_TableEntry{
						TableEntry: &bfrt.TableEntry{
							TableId: p4.SearchTableId(tableName),
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
			for _, v := range rsp.Entities {
				tbl := v.GetTableEntry()
				for _, f := range tbl.Key.Fields {
					fmt.Printf("Match field ID: %d\n", f.FieldId)
					switch strings.Split(reflect.TypeOf(f.GetMatchType()).String(), ".")[1] {
					case "KeyField_Exact_":
						m := f.GetExact()
						fmt.Printf("Match field value: %x\n", m.Value)
					case "KeyField_Ternary_":
						t := f.GetTernary()
						fmt.Printf("Ternary field value: %x, mask: %x\n", t.Value, t.Mask)
					case "KeyField_Lpm":
						l := f.GetLpm()
						fmt.Printf("Lpm field value: %x, prefixLen: %d\n", l.Value, l.PrefixLen)
					case "KeyField_Range_":
						r := f.GetRange()
						fmt.Printf("Range field high value: %x, low value: %x\n", r.High, r.Low)
					}
				}

				printNameById(tbl.Data.ActionId)
				for _, d := range tbl.Data.Fields {
					fmt.Printf("Action parameter field ID: %d\n", d.FieldId)
					printNameById(d.FieldId)
					fmt.Printf("Action parameter value: %x\n", d.GetStream())
				}
			}
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
