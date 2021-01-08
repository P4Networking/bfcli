package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"strings"
)

var (
	clear    bool
	delEntry []string
)

// delFlowCmd represents the delFlow command
var delFlowCmd = &cobra.Command{
	Use:   "del-flow TABLE-NAME ",
	Short: "Delete entries from table",
	Long:  `del-flow can remove all of the entries from specific table`,
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		var argsList []string
		for _, v := range p4Info.Tables {
			if strings.Contains(v.Name, preFixIg) || strings.Contains(v.Name, preFixEg) {
				strs := strings.Split(v.Name, ".")
				if toComplete == "" || strings.Contains(toComplete, "pipe") {
					argsList = append(argsList, v.Name)
				} else {
					argsList = append(argsList, strs[2])
				}
			}
		}
		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cliAddr, ctxAddr, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		if all && !clear && len(args) <= 0 && !cmd.Flag("match").Changed {
			// flag "-a" Clear all tables
			for _, tb := range Obj.p4Info.Tables {
				if NotSupportToReadTable[tb.ID] {
					continue
				}
				Obj.table = append(Obj.table, tb)
			}
		} else if !all && clear && len(args) > 0 && !cmd.Flag("match").Changed {
			// flag "-c" Clear a table
			for _, tb := range Obj.p4Info.Tables {
				if NotSupportToReadTable[tb.ID] {
					continue
				}
				if strings.Contains(tb.Name, args[0]) {
					Obj.table = append(Obj.table, tb)
				}
			}
		} else if !all && !clear && len(args) > 0 && cmd.Flag("match").Changed {
			for a, v := range delEntry {
				delEntry[a] = strings.TrimSpace(v)
			}

			for _, tb := range Obj.p4Info.Tables {
				if NotSupportToReadTable[tb.ID] {
					continue
				}
				if strings.Contains(tb.Name, args[0]) {
					Obj.table = append(Obj.table, tb)
				}
			}

			if len(Obj.table) > 1 {
				fmt.Println(fmt.Errorf("Too many tables matched."))
				for _, k := range Obj.table {
					fmt.Printf("Table : %s\n", k.Name)
				}
				return
			}
			if len(Obj.table) <=0 {
				fmt.Println(fmt.Errorf("No tables matched.\n"))
				return
			}
		} else {
			fmt.Println(fmt.Errorf("check the flag or args"))
			cmd.Help()
			return
		}

		for _, tb := range Obj.table {
			stream, err := cli.Read(ctx, genReadRequestWithId(tb.ID))
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
						fmt.Printf("%s table is empty\n", tb.Name)
						break
					}

					var cnt []int
					var err error
					if cmd.Flag("match").Changed {
						cnt, err = DeleteEntries(&rsp, &cli, &ctx, delEntry)
					} else {
						cnt, err = DeleteEntries(&rsp, &cli, &ctx, nil)
					}
					if err != nil {
						fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", tb.Name, delEntry)
						fmt.Println(err.Error())
						return
					}
					fmt.Printf("%d entires of \"%s\" table have cleared\n", len(cnt), tb.Name)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(delFlowCmd)
	delFlowCmd.Flags().BoolVarP(&all, "all", "a", false, "Clear all the table")
	delFlowCmd.Flags().BoolVarP(&clear, "clear", "c", false, "Clear a table")
	delFlowCmd.Flags().StringSliceVarP(&delEntry, "match", "m", []string{}, "Delete entry by given match key")
}
