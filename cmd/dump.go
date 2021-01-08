package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/southbound/bfrt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	count bool
)

var dumpCmd = &cobra.Command{
	Use:   "dump TABLE-NAME",
	Short: "Dump the existed entries in the specific table",
	Long:  `Display all existed entries in the specific table`,
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
		cliAddr, ctxAddr, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

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
			stream, err := cli.Read(ctx, genReadRequestWithId(v.ID))
			if err != nil {
				log.Fatalf("Got error, %v \n", err.Error())
				return
			}
			DumpEntries(&stream, &v)
		}

		for _, v := range notFounded {
			fmt.Println(fmt.Errorf("Couldn't found %s table", v))
		}
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolVarP(&count, "count", "c", false, "dump entries number of counts")
}
