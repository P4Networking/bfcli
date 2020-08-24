package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	preFixIg = "pipe.SwitchIngress."
	preFixEg = "pipe.SwitchEgress."
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
			tableId := p4Info.SearchTableId(args[0])
			if uint32(tableId) == util.ID_NOT_FOUND {
				fmt.Printf("Can not found table with name: %s\n", args[0])
				return
			}
			table := p4Info.SearchTableById(tableId)
			if table == nil {
				fmt.Printf("Can not found table with ID: %d\n", tableId)
				return
			}

			stream, err := cli.Read(ctx, genReadRequestWithId(uint32(table.ID)))
			if err != nil {
				log.Fatalf("Got error, %v \n", err.Error())
			}

			DumpEntries(&stream, table)
		case true:
			for _, v := range p4Info.Tables {
				if strings.HasPrefix(v.Name, preFixIg) || strings.HasPrefix(v.Name, preFixEg) {
					table := p4Info.SearchTableById(v.ID)
					if table == nil {
						fmt.Printf("Can not found table with ID: %v\n", v.ID)
						return
					}
					stream, err := cli.Read(ctx, genReadRequestWithId(uint32(v.ID)))
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
