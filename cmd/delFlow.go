package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/pisc/util/enums/id"
	"github.com/P4Networking/proto/go/p4"
	"github.com/spf13/cobra"
	"io"
	"log"
	"strings"
)

var (
	reset          bool
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
		argsList, _ := p4Info.GuessTableName(toComplete)
		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {

		if (all || len(delEntry) > 0) && len(args) <= 0 {
			cmd.Help()
			return
		}
		for a, v := range delEntry {
			delEntry[a] = strings.Replace(v, " ", "", -1)
		}

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr
		for _, v := range p4Info.Tables {
			if strings.HasPrefix(v.Name, preFixIg) || strings.HasPrefix(v.Name, preFixEg) {
				if (all || len(delEntry) > 0) && v.Name != args[0] {
					continue
				}
				stream, err := cli.Read(ctx, genReadRequestWithId(v.ID))
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
						var cnt []int
						var err error

						if all || reset {
							cnt, err = DeleteEntries(&rsp, &cli, &ctx)
							if err != nil {
								fmt.Printf("Failed to delete entry: table \"%s\" with fields: %s\n", v.Name, rsp.Entities[cnt[0]].GetTableEntry().Key.Fields)
							} else {
								fmt.Printf("%d entires of \"%s\" table have cleared\n", len(cnt), v.Name)
							}
						} else if len(delEntry) > 0 {
							table, _ := p4Info.GetTableById(id.TableId(v.ID))
							collectedMatchTypes, ok := collectTableMatchTypes(table, &delEntry)
							if !ok {
								fmt.Println("Match keys are not matched")
								return
							}
							fmt.Print("Make Match Key... ")
							if match := BuildMatchKeys(&collectedMatchTypes); match != nil {
								delReq := util.GenWriteRequestWithId(p4.Update_DELETE, id.TableId(v.ID), match, nil)
								fmt.Print("Write Delete Reqeust... ")
								_, err := cli.Write(ctx, delReq)
								if err != nil {
									fmt.Println(err)
									return
								}
								fmt.Println("DONE.")
							} else {
								fmt.Println("Please Check match keys and inputted arguments.")
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
	delFlowCmd.Flags().BoolVarP(&reset, "reset", "r", false, "clear all of the tables")
	delFlowCmd.Flags().StringSliceVarP(&delEntry, "match", "m", []string{}, "delete specific entry by given entry number")
}
