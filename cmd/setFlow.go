package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/pisc/util/enums"
	"github.com/P4Networking/proto/go/p4"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	matchLists   []string
	actionValues []string
	ttl          int32 = -1
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

		tableId := p4Info.SearchTableId(tableName)
		if uint32(tableId) == util.ID_NOT_FOUND {
			fmt.Printf("Can not found table with name: %s\n", tableName)
			return
		}
		table := p4Info.SearchTableById(tableId)

		collectedMatchTypes, ok := collectTableMatchTypes(table)
		if !ok {
			fmt.Println("Match keys are not matched")
			return
		}

		actionId := p4Info.SearchActionId(tableName, actionName)
		if actionId == util.ID_NOT_FOUND {
			fmt.Printf("Can not found action with names: %s\n", actionName)
			return
		}

		collectedActionFieldIds := collectActionFieldIds(table, actionId)
		if len(collectedActionFieldIds) != len(actionValues) {
			fmt.Printf("Length of action fields [%d] != Length of action args [%d]\n", len(collectedActionFieldIds), len(actionValues))
			fmt.Println("Check action arguments")
			return
		}

		fmt.Printf("Make Match Data...")
		match := util.Match()
		for k, v := range collectedMatchTypes {
			mlt, v1, v2 := checkMatchListType(v.matchValue)
			// In EXACT case, v2 value is nil.
			if v.matchType == enums.MATCH_EXACT {
				switch mlt {
				case MAC_TYPE:
					match = append(match, util.GenKeyField(v.matchType, uint32(k), v1.([]byte)))
				case IP_TYPE:
					match = append(match, util.GenKeyField(v.matchType, uint32(k), v1.([]byte)))
				case VALUE_TYPE:
					match = append(match, util.GenKeyField(v.matchType, uint32(k), setBitValue(v1.(int), v.bitWidth)))
				case HEX_TYPE:
					match = append(match, util.GenKeyField(v.matchType, uint32(k), util.HexToBytes(uint16(v1.(uint64)))))
				default:
					fmt.Printf("Unexpect value for EXACT_MATCH : %s\n", v.matchValue)
					return
				}
			} else if v.matchType == enums.MATCH_LPM {
				if mlt == CIDR_TYPE {
					match = append(match, util.GenKeyField(v.matchType, uint32(k), v1.([]byte), v2.(int)))
				} else {
					fmt.Printf("Unexpect value for LPM_MATCH : %s\n", v.matchValue)
					return
				}
			} else if v.matchType == enums.MATCH_TERNARY {
				//Ternary match only support the complete address format(aa:aa:aa:aa:aa:aa/ff:ff:ff:ff:ff:ff, x.x.x.x/255.255.255.255)
				if mlt == MASK_TYPE {
					switch v1.(int) {
					case IP_MASK:
						arg := v2.([]string)
						match = append(match, util.GenKeyField(v.matchType, uint32(k), util.Ipv4ToBytes(arg[0]), util.Ipv4ToBytes(arg[1])))
					case ETH_MASK:
						arg := v2.([]string)
						match = append(match, util.GenKeyField(v.matchType, uint32(k), util.MacToBytes(arg[0]), util.MacToBytes(arg[1])))
					case HEX_MASK:
						arg := v2.([]uint16)
						match = append(match, util.GenKeyField(v.matchType, uint32(k), util.HexToBytes(arg[0]), util.HexToBytes(arg[1])))
					case VALUE_MASK:
						arg := v2.([]int)
						match = append(match, util.GenKeyField(v.matchType, uint32(k), setBitValue(arg[0], v.bitWidth), setBitValue(arg[1], v.bitWidth)))
					}
				} else {
					fmt.Printf("Unexpect value for TERNARY_MATCH : %s\n", v.matchValue)
					return
				}
			} else if v.matchType == enums.MATCH_RANGE {
				//TODO: Implement range match
				fmt.Println("Range_Match Not Supported Yet")
				return
			} else {
				fmt.Println("Unexpected Match Type")
			}
		}

		fmt.Printf("   Make Action Data...")
		action := util.Action()
		if len(collectedActionFieldIds) != 0 {
			for k, v := range collectedActionFieldIds {
				switch mlt, v1, _ := checkMatchListType(actionValues[k-1]); mlt {
				case MAC_TYPE:
					action = append(action, util.GenDataField(uint32(k), v1.([]byte)))
				case IP_TYPE:
					action = append(action, util.GenDataField(uint32(k), v1.([]byte)))
				case VALUE_TYPE:
					action = append(action, util.GenDataField(uint32(k), setBitValue(v1.(int), v)))
				default:
					fmt.Println("Unexpected value for action fields")
					return
				}
			}
		}
		if ttl >= 0 {
			action = append(action, util.GenDataField(p4Info.SearchDataId(tableName, "$ENTRY_TTL"), util.Int32ToBytes(uint32(ttl)*1000)))
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
	setFlowCmd.Flags().Int32VarP(&ttl, "ttl", "t", ttl, "TTL arguments")
}
