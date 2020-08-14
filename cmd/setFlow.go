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
	"github.com/P4Networking/pisc/util/enums"
	"github.com/P4Networking/proto/go/p4"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
)


var (
	//Name list of the match field
	matchLists [] string
	//value list fo the action field
	actionValues []string
)
// setFlowCmd represents the setFlow command
var setFlowCmd = &cobra.Command{
	Use:   "set-flow TABLE_NAME ACTION-NAME -m \"match key,...\" -a \"action value,...\" ",
	Short: "Set flow into table",
	Long: `Insert flow into specific table with specific action`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		tableName := args[0]
		actionName := args [1]

		cliAddr, ctxAddr, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		cli := *cliAddr
		ctx := *ctxAddr

		tableId := p4Info.SearchTableId(tableName)
		if tableId == util.ID_NOT_FOUND {
			fmt.Printf("Can not found table with name: %s\n", tableName)
			return
		}
		table := p4Info.SearchTableById(tableId)

		collectedMatchTypes, ok := collectTableMatchTypes(table)
		if !ok {
			fmt.Println("Match type not matched")
			return
		}

		for a, v := range matchLists {
			matchLists[a] = strings.Replace(v, " ", "", -1)
		}

		for a, v := range actionValues {
			actionValues[a] = strings.Replace(v, " ", "", -1)
		}

		if len(collectedMatchTypes) != len(matchLists) {
			fmt.Println("Length of Match keys doesn't matched with Length of args")
			fmt.Println(collectedMatchTypes, matchLists)
			return
		}

		actionId := p4Info.SearchActionId(tableName, actionName)
		if actionId == util.ID_NOT_FOUND {
			fmt.Printf("Can not found action with names: %s\n", actionName)
			return
		}

		collectedActionFieldIds:= collectActionFieldIds(table, actionId)
		if len(collectedActionFieldIds) != len(actionValues) {
			fmt.Println("Length of action field dosn't matched with Length of action args")
			fmt.Println(collectedActionFieldIds, actionValues)
			return
		}

		match := util.Match()
		for k, v := range collectedMatchTypes {
			MLT := checkMatchListType(matchLists[k-1])
			if v.matchType == enums.MATCH_EXACT {
				switch MLT {
				case MAC_TYPE:
					match = append(match, util.GenKeyField(v.matchType, uint32(k), util.MacToBytes(matchLists[k-1])))
				case IP_TYPE:
					match = append(match, util.GenKeyField(v.matchType, uint32(k), util.Ipv4ToBytes(matchLists[k-1])))
				case VALUE_TYPE:
					arg, _ := strconv.Atoi(matchLists[k-1])
					match = append(match, util.GenKeyField(v.matchType, uint32(k), ParseBitWidth(arg, v.bitWidth)))
				default:
					fmt.Printf("Unexpect value for EXACT_MATCH : %s\n", matchLists[k-1])
					return
				}
			} else if v.matchType == enums.MATCH_LPM {
				if MLT == CIDR_TYPE {
					arg := strings.Split(matchLists[k-1], "/")
					subnet, _ := strconv.Atoi(arg[1])
					fmt.Printf("prefix : %d\n", subnet)
					match = append(match, util.GenKeyField(v.matchType, uint32(k), util.Ipv4ToBytes(arg[0]), subnet))
				} else {
					fmt.Printf("Unexpect value for LPM_MATCH : %s\n", matchLists[k-1])
					return
				}
			} else if v.matchType == enums.MATCH_TERNARY {
				if MLT ==  MASK_TYPE {
					arg := strings.Split(matchLists[k-1], "/")
					switch checkMaskType(matchLists[k-1]) {
					case IP_MASK:
						match = append(match, util.GenKeyField(v.matchType, uint32(k), util.Ipv4ToBytes(arg[0]), util.Ipv4ToBytes(arg[1])))
					case ETH_MASK:
						match = append(match, util.GenKeyField(v.matchType, uint32(k), util.MacToBytes(arg[0]), util.MacToBytes(arg[1])))
					case HEX_MASK:
						arg[0] = strings.Replace(arg[0], "0x", "", -1)
						arg[1] = strings.Replace(arg[1], "0x", "", -1)
						match = append(match, util.GenKeyField(v.matchType, uint32(k), util.HexToBytes(uint16(util.HexToInt(arg[0]))), util.HexToBytes(uint16(util.HexToInt(arg[1])))))
					case VALUE_MASK:
						argu1, _ := strconv.Atoi(arg[0])
						argu2, _ := strconv.Atoi(arg[1])
						match = append(match, util.GenKeyField(v.matchType, uint32(k), ParseBitWidth(argu1, v.bitWidth), ParseBitWidth(argu2, v.bitWidth)))
					}
				} else {
					fmt.Printf("Unexpect value for TERNARY_MATCH : %s\n", matchLists[k-1])
					return
				}
			} else if v.matchType == enums.MATCH_RANGE {
				//TODO: Implement range match
			} else {
				fmt.Println("Unexpected Match Type")
			}
		}

		action := util.Action()
		if len(collectedActionFieldIds) !=0 {
			for k, v := range collectedActionFieldIds {
				switch checkMatchListType(actionValues[k-1]) {
				case MAC_TYPE:
					action = append(action, util.GenDataField(uint32(k), util.MacToBytes(actionValues[k-1])))
				case IP_TYPE:
					action = append(action, util.GenDataField(uint32(k), util.Ipv4ToBytes(actionValues[k-1])))
				case VALUE_TYPE:
					actionValue, _ := strconv.ParseInt(actionValues[k-1], 10, 32)
					var result []byte
					switch v {
					case INT32:
						result = util.Int32ToBytes(uint32(actionValue))
					case INT16:
						result = util.Int16ToBytes(uint16(actionValue))
					case INT8:
						result = util.Int8ToBytes(uint8(actionValue))
					}
					action = append(action, util.GenDataField(uint32(k), result))
				}
			}
		}
		var req = util.GenWriteRequestWithId(
					p4.Update_INSERT,
					tableId,
					match,
					&p4.TableData{
							ActionId: actionId,
							Fields: action,
		})

		if _, err := cli.Write(ctx, req); err != nil {
			log.Printf("Got error, %v \n", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(setFlowCmd)
	setFlowCmd.Flags().StringSliceVarP(&matchLists, "match key", "m", []string{}, "")
	setFlowCmd.Flags().StringSliceVarP(&actionValues, "action value", "a", []string{}, "")
}
