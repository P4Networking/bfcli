package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/southbound/bfrt"
	"github.com/P4Networking/pisc/util/enums/id"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump TABLE-NAME",
	Short: "Dump the existed entries in the specific table",
	Long:  `Display all existed entries in the specific table`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		argsList, _ := p4Info.GuessTableName(toComplete)
		for k, v := range argsList {
			if strings.Contains(v, preFixIgPar) || strings.Contains(v, preFixEgPar) {
				argsList[k] = argsList[len(argsList)-1] // Copy last element to index i.
				argsList[len(argsList)-1] = ""   // Erase last element (write zero value).
				argsList = argsList[:len(argsList)-1]
			}
		}
		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		switch all {
		case false:
			if len(args) <= 0 {
				cmd.Help()
				return
			}
			tableId, ok := p4Info.GetTableId(args[0])
			if uint32(tableId) == bfrt.ID_NOT_FOUND || !ok {
				fmt.Printf("Can not found table with name: %s\n", args[0])
				return
			}
			table, ok := p4Info.GetTableById(tableId)
			if !ok {
				fmt.Printf("Can not found table with ID: %d\n", tableId)
				return
			}
			stream, err := cli.Read(ctx, genReadRequestWithId(table.ID))
			if err != nil {
				log.Fatalf("Got error, %v \n", err.Error())
			}

			DumpEntries(&stream, table)
		case true:
			for _, v := range p4Info.Tables {
				if strings.HasPrefix(v.Name, preFixIg) || strings.HasPrefix(v.Name, preFixEg) {
					table, _ := p4Info.GetTableById(id.TableId(v.ID))
					if table == nil {
						fmt.Printf("Can not found table with ID: %v\n", v.ID)
						return
					}
					stream, err := cli.Read(ctx, genReadRequestWithId(v.ID))
					if err != nil {
						log.Fatalf("Got error, %v \n", err.Error())
					}

					DumpEntries(&stream, table)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolVarP(&all, "all", "a", false, "dump all of the tables")
}
