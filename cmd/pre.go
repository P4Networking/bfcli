package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util/enums/id"
	"github.com/P4Networking/proto/go/p4"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
)

var (
	array 	[]string
	node	string
	mgid	string
	l2xid	string
	lag 	string
)

// setPreCmd represents the Packet Replication Engine command
var setPreCmd = &cobra.Command{
	Use:   "pre",
	Short: "Set Packet Replication Engine (PRE) entry",
	Long:  "Insert the flow to table with action",
	Args:  cobra.MaximumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		toComplete = "mirror"
		ret := make([]string, 0)
		if len(args) < 1 {
			//argsList, _ := Obj.nonP4Info.String()
			for _, v := range Obj.nonP4Info.Tables {
				ret = append(ret, v.Name)
			}
			//ret = append(ret, Obj.nonP4Info.Tables)
			return ret, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli, ctx, conn, cancel, _, _ := initConfigClient()
		cliAddr = cli
		ctxAddr = ctx
		defer conn.Close()
		defer cancel()

		if len(args) < 1 {
			fmt.Println("please checkout the table name.")
			return
		}

		tableName := args[0]
		req := &p4.WriteRequest{
			ClientId: 0,
			Target: &p4.TargetDevice{
				DeviceId: 0,
				PipeId:   0xffff,
			},
			Updates: []*p4.Update {},
		}
		found := false
		for _, v := range Obj.nonP4Info.Tables {
			if NotSupportToReadTable[v.ID] {
				continue
			}
			if strings.Contains(v.Name, tableName) && node != "" && len(array) > 0 {
				portIds := make([]id.PortId, len(array))
				for k, v := range array {
					v = strings.ReplaceAll(v, " ", "")
					value, err := strconv.ParseUint(v, 10, 16)
					if err != nil {
						fmt.Println(err)
						panic("Unable to convert the ports value type from string to uint32")
					}
					portIds[k] = id.PortId(uint16(value))
				}
				value, err := strconv.ParseUint(node, 10, 32)
				if err != nil {
					fmt.Println(err)
					panic("Unable to convert the node value type from string to uint32")
				}
				req.Updates = append(req.Updates, Obj.nonP4Info.SetNode(uint32(value), portIds))
				found = true
				break
			} else if strings.Contains(v.Name, tableName) && mgid != "" && node != "" {
				mgidValue, err := strconv.ParseUint(mgid, 10, 16)
				if err != nil {
					fmt.Println(err)
					panic("Unable to convert the mgid value type from string to uint16")
				}
				nodeValue, err := strconv.ParseUint(node, 10, 32)
				if err != nil {
					fmt.Println(err)
					panic("Unable to convert the node value type from string to uint32")
				}
				req.Updates = append(req.Updates, Obj.nonP4Info.SetMGID(uint16(mgidValue), uint32(nodeValue)))
				found = true
				break
			} else if strings.Contains(v.Name, tableName) && l2xid != "" && len(array) > 0 {
				// pre.prune
				nodeArray := make([]uint32, len(array))
				for k, v := range array {
					v = strings.ReplaceAll(v, " ", "")
					value, err := strconv.ParseUint(v, 10, 16)
					if err != nil {
						fmt.Println(err)
						panic("Unable to convert the ports value type from string to uint32")
					}
					nodeArray[k] = uint32(value)
				}
				l2xidValue, err := strconv.ParseUint(l2xid, 10, 32)
				if err != nil {
					fmt.Println(err)
					panic("Unable to convert the l2xid value type from string to uint32")
				}
				ret := Obj.nonP4Info.SetPrune(uint16(l2xidValue), nodeArray)
				req.Updates = append(req.Updates, ret.Updates[0])
				found = true
				break
			} else if strings.Contains(v.Name, tableName) && lag != "" && len(array) > 0 {
				// PRE.LAG, Link aggregation.
				//TODO: couldn't insert the flow with the 0xff(255) entry number.
				lagValue, err := strconv.ParseUint(lag, 10, 8)
				if err != nil {
					fmt.Println(err)
					panic("Unable to convert the lag value type from string to uint32")
				}
				nodeArray := make([]uint32, len(array))
				for k, v := range array {
					v = strings.ReplaceAll(v, " ", "")
					value, err := strconv.ParseUint(v, 10, 32)
					if err != nil {
						fmt.Println(err)
						panic("Unable to convert the ports value type from string to uint32")
					}
					nodeArray[k] = uint32(value)
				}
				ret := Obj.nonP4Info.SetLag(uint8(lagValue), nodeArray)
				req.Updates = append(req.Updates, ret.Updates[0])
				found = true
				break
			}
		}

		if !found {
			_ = cmd.Help()
			return
		}
		if _, err := (*cliAddr).Write(*ctxAddr, req); err != nil {
			log.Printf("Got an error, %v \n", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(setPreCmd)
	setPreCmd.Flags().StringSliceVarP(&array, "PORTS", "p", []string{}, "ports list")
	setPreCmd.Flags().StringVarP(&node, "NODE ID", "n", "", "node id")
	setPreCmd.Flags().StringVarP(&mgid, "MGID ID", "m", "", "mgid id")
	setPreCmd.Flags().StringVarP(&l2xid, "L2XID ID", "x", "", "l2xid id")
	setPreCmd.Flags().StringVarP(&lag, "LAG ID", "l", "", "lag id")
}