package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"strconv"
	"strings"
)

var (
	reset          bool
	delEntryNumber []string
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
		argsList, _ := p4Info.GuessTableName(toComplete)
		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {

		if (all || len(delEntryNumber) > 0) && len(args) <= 0 {
			cmd.Help()
			return
		}
		for a, v := range delEntryNumber {
			delEntryNumber[a] = strings.Replace(v, " ", "", -1)
		}

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		for _, v := range p4Info.Tables {
			if strings.HasPrefix(v.Name, preFixIg) || strings.HasPrefix(v.Name, preFixEg) {
				if (all || len(delEntryNumber) > 0) && v.Name != args[0] {
					continue
				}
				stream, err := cli.Read(ctx, genReadRequestWithId(uint32(v.ID)))
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
							break
						}
						var cnt, deletedEntries []int
						var err error
						if all || reset {
							cnt, err = DeleteEntries(&rsp, &cli, &ctx, -1)
							if err != nil {
								fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", v.Name, rsp.Entities[cnt[0]].GetTableEntry().Key.Fields)
							} else {
								fmt.Printf("%d entires of \"%s\" table have cleared\n", len(cnt), v.Name)
							}
						} else if len(delEntryNumber) > 0 {
							for _, dv := range delEntryNumber{
								n, _ := strconv.Atoi(dv)
								cnt, err = DeleteEntries(&rsp, &cli, &ctx, n)
								if err != nil {
									fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", v.Name, rsp.Entities[cnt[0]].GetTableEntry().Key.Fields)
								} else {
									deletedEntries = append(deletedEntries, cnt[0])
								}
							}
							if len(deletedEntries) > 0 {
								fmt.Printf("%d entires of \"%s\" table have cleared : %v\n", len(deletedEntries), v.Name, deletedEntries)
							}
						}
					}
				}
			}
		}
		if reset {
			fmt.Println("Reset complete.")
		}
	},
}

func init() {
	rootCmd.AddCommand(delFlowCmd)
	delFlowCmd.Flags().BoolVarP(&all, "all", "a", false, "delete all entries of the table")
	delFlowCmd.Flags().StringSliceVarP(&delEntryNumber, "entry", "e", []string{}, "delete specific entry by given entry number")
	delFlowCmd.Flags().BoolVarP(&reset, "reset", "r", false, "clear all of the tables")
}
