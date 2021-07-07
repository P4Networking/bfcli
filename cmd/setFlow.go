package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/pisc/util/enums/id"
	"github.com/P4Networking/proto/go/p4"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	matchKeyList	[]string
	actionValues	[]string
	ttl				= ""
	file			string
	cliAddr 		*p4.BfRuntimeClient
	ctxAddr			*context.Context
	filedata [][]string
)

// setFlowCmd represents the setFlow command
var setFlowCmd = &cobra.Command{
	Use:   "set-flow TABLE_NAME ACTION-NAME ",
	Short: "Insert the flows",
	Long:  "Insert the flow to table with action",
	Args:  cobra.MaximumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		if len(args) < 1 {
			ret := make([]string, 0)
			argsList, _ := Obj.p4Info.GuessTableName(toComplete)
			for _, v := range argsList {
				if strings.Contains(v, preFixIg) || strings.Contains(v, preFixEg) {
					name := strings.Split(v, ".")
					ret = append(ret, name[len(name)-2] + "." +  name[len(name)-1])
				}
			}
			return ret, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			for _, v := range Obj.p4Info.Tables {
				if strings.Contains(v.Name, args[0]) {
					ret  := make([]string, 0)
					for _, action := range v.ActionSpecs {
						name := strings.Split(action.Name, ".")
						ret = append(ret, name[len(name)-1])
					}
					return ret, cobra.ShellCompDirectiveNoFileComp
				}
			}
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("file").Changed && (cmd.Flag("match").Changed || cmd.Flag("action").Changed || cmd.Flag("ttl").Changed) {
			fmt.Println(fmt.Errorf("file flag can only exist alone."))
			os.Exit(1)
		}
		if cmd.Flag("file").Changed {
			fptr := flag.String("fpath", file, "file path to read from")
			flag.Parse()
			f, err := os.Open(*fptr)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err = f.Close(); err != nil {
				log.Fatal(err)
			}
			}()
			s := bufio.NewScanner(f)
			for s.Scan() {
				if s.Text() == "" {
					continue
				}
				filedata = append(filedata, strings.Split(strings.ReplaceAll(s.Text(), ", ", ","), " "))
			}
			err = s.Err()
			if err != nil {
				log.Fatal(err)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli, ctx, conn, cancel, _, _ := initConfigClient()
		cliAddr = cli
		ctxAddr = ctx
		defer conn.Close()
		defer cancel()
		if cmd.Flag("file").Changed {
			setFlowSub.Flags().StringSliceVarP(&matchKeyList, "match", "m", []string{}, "match key arguments")
			setFlowSub.Flags().StringSliceVarP(&actionValues, "action", "a", []string{}, "action arguments")
			setFlowSub.Flags().StringVarP(&ttl, "ttl", "t", "", "TTL arguments")
			cnt := 0
			bar := progressbar.Default(int64(len(filedata)))
			for _, line := range filedata {
				if line[0] == "" {
					continue
				}
				err := setFlowSub.ParseFlags(line)
				matchKeyList = strings.Split(matchKeyList[0], ",")
				if len(actionValues) < 1 || actionValues[0] == "" {
					actionValues = nil
				} else {
					actionValues = strings.Split(actionValues[0], ",")
				}
				if err != nil {
						panic(err)
				}
				setFlowSub.Run(cmd, line[:2])
				matchKeyList = nil
				actionValues = nil
				ttl = ""
				cnt++
				_ = bar.Add(1)
			}
			fmt.Printf("Have successfully wrote %d entries.\n", cnt)
		} else {
			setFlowSub.Flags().StringSliceVarP(&matchKeyList, "match", "m", matchKeyList, "match key arguments")
			setFlowSub.Flags().StringSliceVarP(&actionValues, "action", "a", actionValues, "action arguments")
			setFlowSub.Flags().StringVarP(&ttl, "ttl", "t", ttl, "TTL arguments")
			setFlowSub.Run(cmd, args)
		}
	},
}

var setFlowSub = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		ObjInit()
		for a, v := range matchKeyList {
			matchKeyList[a] = strings.TrimSpace(v)
		}
		for a, v := range actionValues {
			actionValues[a] = strings.TrimSpace(v)
		}
		// find the table that have substring
		for _, tb := range Obj.p4Info.Tables {
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
		if len(Obj.table) <= 0 {
			fmt.Println(fmt.Errorf("No tables matched.\n"))
			return
		}

		for _, table := range Obj.table {
			for _, actionSpec := range table.ActionSpecs {
				if strings.Compare(args[1], actionSpec.Name[strings.LastIndex(actionSpec.Name, ".")+1:]) == 0 {
					Obj.actions[actionSpec.ID] = actionSpec.Name
					Obj.actionId = actionSpec.ID
					Obj.actionName = actionSpec.Name
				}
			}
		}
		if len(Obj.actions) > 1 {
			fmt.Println(fmt.Errorf("Too many actions matched.\n"))
			return
		}
		if len(Obj.actions) <= 0 {
			fmt.Println(fmt.Errorf("No actions have matched.\n"))
			return
		}

		collectedMatchTypes, ok := collectTableMatchTypes(&matchKeyList)
		if !ok {
			fmt.Println("Match key length isn't match")
			return
		}

		ttlId, ttlok := Obj.p4Info.GetDataId(Obj.table[0].Name, "$ENTRY_TTL")
		if ttlok && ttl == "" {
			fmt.Println("table has ttl entry but it's not set, using default value 1000 for ttl")
			ttl = "1000"
		}

		collectedActionFieldIds, err := collectActionFieldIds(&Obj.table[0], Obj.actionId, actionValues)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(collectedActionFieldIds) != len(actionValues) {
			fmt.Printf("the expected action field length is : %d, the received action filed length is : %d\n", len(collectedActionFieldIds), len(actionValues))
			return
		}

		//fmt.Printf("Make Match Data...")
		match := BuildMatchKeys(&collectedMatchTypes)
		if match == nil {
			return
		}

		//fmt.Printf("   Make Action Data...")
		action := util.Action()
		if len(collectedActionFieldIds) > 0 {
			for _, v := range collectedActionFieldIds {
				switch mlt, v1, _ := checkMatchListType(v.actionValue, v.parsedBitWidth); mlt {
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

		if ttlok {
			l, err := strconv.ParseUint(ttl, 10, 32)
			if err != nil {
				fmt.Printf("Please Check the TTL value %s.\n", ttl)
				return
			}
			action = append(action, util.GenDataField(ttlId, util.Int32ToBytes(uint32(l))))
		}

		var req = util.GenWriteRequestWithId(p4.Update_INSERT, id.TableId(Obj.table[0].ID), match, &p4.TableData{ActionId: Obj.actionId, Fields: action})
		if _, err := (*cliAddr).Write(*ctxAddr, req); err != nil {
			log.Printf("Got an error, %v \n", err.Error())
			return
		}
	},
}


func init() {
	rootCmd.AddCommand(setFlowCmd)
	setFlowCmd.Flags().StringVarP(&file, "file", "f", "", "read flow file to insert the flow entry")
	setFlowCmd.Flags().StringSliceVarP(&matchKeyList, "match", "m", []string{}, "match key arguments")
	setFlowCmd.Flags().StringSliceVarP(&actionValues, "action", "a", []string{}, "action arguments")
	setFlowCmd.Flags().StringVarP(&ttl, "ttl", "t", "", "TTL arguments")
}