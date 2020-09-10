package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/southbound/bfrt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/proto/go/p4"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
)

var (
	matchLists   []string
	actionValues []string
	ttl          = ""
)

// setFlowCmd represents the setFlow command
var setFlowCmd = &cobra.Command{
	Use:   "set-flow TABLE_NAME ACTION-NAME ",
	Short: "Set flow into table",
	Long:  `Insert the flow to table with action`,
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		argsList, _ := p4Info.GuessTableName(toComplete)
		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {

		tableName := args[0]
		actionName := args[1]
		for a, v := range matchLists {
			matchLists[a] = strings.Replace(v, " ", "", -1)
		}

		for a, v := range actionValues {
			actionValues[a] = strings.Replace(v, " ", "", -1)
		}

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		tableId, ok := p4Info.GetTableId(tableName)
		if uint32(tableId) == bfrt.ID_NOT_FOUND || !ok {
			fmt.Printf("Can not found table with name: %s\n", tableName)
			return
		}
		table, _ := p4Info.GetTableById(tableId)

		collectedMatchTypes, ok := collectTableMatchTypes(table, &matchLists)
		if !ok {
			fmt.Println("Match keys are not matched")
			return
		}

		actionId, ok := p4Info.GetActionId(tableName, actionName)
		if actionId == bfrt.ID_NOT_FOUND || !ok {
			fmt.Printf("Can not found action with names: %s\n", actionName)
			return
		}
		dataId, ok := p4Info.GetDataId(tableName, "$ENTRY_TTL")
		if ok && ttl == "" {
			fmt.Printf("Please set the TTL value for table %s\n", table.Name)
			return
		}

		collectedActionFieldIds, err := collectActionFieldIds(table, actionId, actionValues)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(collectedActionFieldIds) != len(actionValues) {
			fmt.Printf("Length of action fields [%d] != Length of action args [%d]\n", len(collectedActionFieldIds), len(actionValues))
			fmt.Println("Check action arguments")
			return
		}

		fmt.Printf("Make Match Data...")
		match := BuildMatchKeys(&collectedMatchTypes)

		fmt.Printf("   Make Action Data...")
		action := util.Action()
		if len(collectedActionFieldIds) != 0 {
			for _, v := range collectedActionFieldIds {
				switch mlt, v1, _ := checkMatchListType(v.actionValue); mlt {
				case MAC_TYPE:
					action = append(action, util.GenDataField(v.fieldId, v1.([]byte)))
				case IP_TYPE:
					action = append(action, util.GenDataField(v.fieldId, v1.([]byte)))
				case VALUE_TYPE:
					action = append(action, util.GenDataField(v.fieldId, setBitValue(v1.(int), v.parsedBitWidth)))
				default:
					fmt.Println("Unexpected value for action fields")
					return
				}
			}
		}
		if ttl != "" {
			if !ok {
				fmt.Println("ttl set failed")
				return
			}
			l, err :=strconv.ParseUint(ttl, 10 , 32)
			if err != nil {
				fmt.Printf("Please Check the TTL value %s.\n", ttl)
				return
			}
			action = append(action, util.GenDataField(dataId, util.Int32ToBytes(uint32(l))))
		}

		fmt.Printf("   Make Write Request...")
		var req = util.GenWriteRequestWithId(p4.Update_INSERT, tableId, match, &p4.TableData{ActionId: actionId, Fields: action})
		if _, err := cli.Write(ctx, req); err != nil {
			log.Printf("Got error, %v \n", err.Error())
		} else {
			fmt.Printf("   Write Done.\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(setFlowCmd)
	setFlowCmd.Flags().StringSliceVarP(&matchLists, "match", "m", []string{}, "match arguments")
	setFlowCmd.Flags().StringSliceVarP(&actionValues, "action", "a", []string{}, "action arguments")
	setFlowCmd.Flags().StringVarP(&ttl, "ttl", "t", ttl, "TTL arguments")
}
