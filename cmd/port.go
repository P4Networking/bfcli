package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util/enums/id"
	p "github.com/P4Networking/pisc/util/enums/port"
	"github.com/P4Networking/proto/go/p4"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	portNum 	uint16
	speed		string
	fec			int
	an	 		uint
)

// setPortCmd represents the add port command
var setPortCmd = &cobra.Command{
	Use:   "port",
	Short: "Set switch port properties",
	Long:  "setting switch port properties. \nport speed : 10G, 25G, 40G, 50G, 100G\nport fec : \n",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// WARN: the port command only support an add feature now,
		// it means can't mod or delete the port information through the add command.

		cli, ctx, conn, cancel, _, _ := initConfigClient()
		cliAddr = cli
		ctxAddr = ctx
		defer conn.Close()
		defer cancel()

		// Check Port Number
		if !cmd.Flag("Port Number").Changed {
			fmt.Println("please set the port number.")
			return
		}

		// Check Port Speed
		if !cmd.Flag("Port Speed").Changed {
			fmt.Println("Port Speed isn't set, take the default value of speed : 100G")
			speed = "100G"
		}
		Speed := map[string]int{
			"100G" : p.BF_SPEED_100G,
			"50G"  : p.BF_SPEED_50G,
			"40G"  : p.BF_SPEED_40G,
			"25G"  : p.BF_SPEED_25G,
			"10G"  : p.BF_SPEED_10G,
		}
		if Speed[speed] <= 0 {
			fmt.Println("please check port speed value : 100G, 50G, 40G, 20G, 10G")
			return
		}

		// Check Port Fec
		FEC := map[int] string {
			0 : "BF_FEC_TYP_NONE",
			1 : "BF_FEC_TYP_FIRECODE",
			2 : "BF_FEC_TYP_REED_SOLOMON",
		}
		if !cmd.Flag("Port Fec").Changed {
			fec = 0
			fmt.Printf("Port fec isn't set, take the default value of fec : %s\n", FEC[fec])
		} else {
			if fec > 2 || fec < 0 {
				fmt.Printf("Fec value is out of range, please input the below value :\n" +
					"\t%s(Default) : %d\n" +
					"\t%s : %d\n" +
					"\t%s: %d\n",
					FEC[0], 0,
					FEC[1], 1,
					FEC[2], 2)
				return
			}
		}

		// Check Auto-Negotiation
		An := map[uint]string {
			0 : "PM_AN_FORCE_DISABLE",
			1 : "PM_AN_FORCE_ENABLE",
			2 :"PM_AN_MAX",
		}
		if !cmd.Flag("Auto Negotiation").Changed {
			an = 0
			fmt.Printf("Auto Negotiation isn't set. take default value of Auto Negotiation : %s\n", An[an])
		} else {
			if an > 2 || an < 0 {
				fmt.Printf("please check the auto negotiation value.\n" +
					"%s : 0\n" +
					"%s : 1\n" +
					"%s : 2\n" ,
					An[0],
					An[1],
					An[2])
				return
			}
		}

		// Check Table Name
		tableName := "PORT"
		found := false
		for _, v:= range Obj.nonP4Info.Tables {
			if strings.Contains(v.Name, tableName) {
				found = true
			}
		}
		if !found {
			_ = cmd.Help()
			return
		}

		req := &p4.WriteRequest{
			ClientId: 0,
			Target: &p4.TargetDevice{
				DeviceId: 0,
				PipeId:   0xffff,
			},
			Updates: []*p4.Update {},
		}

		var lanes uint32
		switch speed {
		case "100G":
			lanes = 4
		case "40G":
			lanes = 4
		case "50G":
			lanes = 2
		case "25G":
			lanes = 1
		case "10G":
			lanes = 1
		default:
			lanes = 1
		}
		req.Updates = append(req.Updates, Obj.nonP4Info.AddPort(
			id.PortId(portNum),
			p.PortSpeedType(Speed[speed]),
			p.BF_FEC_Type(fec),
			lanes,
			p.BF_PM_Port_Autoneg_Policy(An[an])))

		if _, err := (*cliAddr).Write(*ctxAddr, req); err != nil {
			log.Printf("Got an error, %v \n", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(setPortCmd)
	setPortCmd.Flags().Uint16VarP(&portNum,"Port Number", "p", 0, "num")
	setPortCmd.Flags().StringVarP(&speed, "Port Speed", "s", "", "speed")
	setPortCmd.Flags().IntVarP(&fec, "Port Fec", "f", 0, "fec")
	setPortCmd.Flags().UintVarP(&an,  "Auto Negotiation", "a", 0, "an")
}