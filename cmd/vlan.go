package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	vlanPort 	string
	vlanName 	string
	vlanType 	string
	vlanId   	[]string
	opt			map[string]interface{}
	optSet		[]string
	protoOpt	[]string
)
type reponseMessage struct {
	IsSuccess  bool
	Message   string
}
// setVlanCmd represents the add port command
var setVlanCmd = &cobra.Command{
	Use:   "vlan",
	Short: "Set VLAN",
	Long:  "Define port where belongs to the specific VLANs",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) < 1 {
			return optSet, cobra.ShellCompDirectiveNoFileComp
		}
		ret := make([]string,0)
		for _, v := range optSet {
			if strings.Contains(toComplete, v) {
				ret = append(ret, v)
			}
		}
		return ret, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case optSet[0]:
			// VLAN Create
			if !cmd.Flag("id").Changed || !cmd.Flag("name").Changed {
				fmt.Println("check VLAN ID and VLAN Name")
				_ = cmd.Help()
				return
			}
			opt[args[0]].(func(string, string))(vlanId[0], vlanName)

		case optSet[1]:
			// VLAN Add
			if !cmd.Flag("type").Changed || !cmd.Flag("port").Changed || !cmd.Flag("id").Changed {
				fmt.Println("Check VLAN port, VLAN type and VLAN Id")
				_ = cmd.Help()
				return
			}
			opt[args[0]].(func(string, string, []string))(vlanType, vlanPort, vlanId)

		case optSet[2]:
			// VLAN Delete
			if !cmd.Flag("id").Changed && !cmd.Flag("port").Changed{
				fmt.Println("vlanId and portId are not given.")
				_ = cmd.Help()
				return
			}
			if cmd.Flag("id").Changed && !cmd.Flag("port").Changed {
				opt[args[0]].(func([]string, string, int))(vlanId, vlanPort, 0)
			} else if cmd.Flag("id").Changed && cmd.Flag("port").Changed {
				opt[args[0]].(func([]string, string, int))(vlanId, vlanPort, 1)
			}
		case optSet[3]:
			// VLAN show
			if !cmd.Flag("id").Changed && !cmd.Flag("port").Changed {
				opt[args[0]].(func(interface{}, int))(0, 2)
			} else if cmd.Flag("id").Changed && !cmd.Flag("port").Changed {
				opt[args[0]].(func(interface{}, int))(vlanId, 0)
			} else if !cmd.Flag("id").Changed && cmd.Flag("port").Changed {
				opt[args[0]].(func(interface{}, int))(vlanPort, 1)
			} else {
				_ = cmd.Help()
			}
		default:
			_ = cmd.Help()
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(setVlanCmd)
	setVlanCmd.Flags().StringVarP(&vlanPort, "port", "p", "", "vlan port")
	setVlanCmd.Flags().StringVarP(&vlanName, "name", "n", "", "vlan name")
	setVlanCmd.Flags().StringVarP(&vlanType, "type", "t", "", "vlan type (Access, Trunk)")
	setVlanCmd.Flags().StringArrayVarP(&vlanId, "id", "v", []string{}, "vlan id")

	// command setting
	optSet = []string{"create","add", "delete","show"}
	opt  = map[string]interface{} {
		optSet[0] : vlanCreate,
		optSet[1] : vlanAdd,
		optSet[2] : vlanDelete,
		optSet[3] :	vlanShow,
	}
	protoOpt = []string{"create", "portupdate", "delete", "get"}
}

func makePost(opt string, body *strings.Reader) (resp *http.Response, err error) {
	httpAddr := "http://localhost:50101/v1/vlan/"
	contentType := "application/x-www-form-urlencoded"
	return http.Post(httpAddr+opt,contentType, body)
}

func vlanCreate(vid, vlanName string) {
	body := strings.NewReader(fmt.Sprintf(`{"vlanId": %s, "name": "%s"}`, vid, vlanName))
	res, err := makePost(protoOpt[0], body)
	if err != nil {
		log.Fatal(err)
	}
	resbody, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	printMessage(resbody)
}

func vlanAdd(vlanType string, port string, vid []string) {
	ports := rangeSplit(port)
	for _, p := range ports {
		switch vlanType {
		case "trunk":
			vlanSetTagged(p, vid)
			break
		case "access":
			if len(vid) > 1 {
				log.Panic("access port can only configure with 1 vlanId")
			}
			vlanSetUntag(p, vid)
			break
		default:
			log.Fatal("check vlan type what you inputted.")
		}
	}
}

func vlanSetUntag(port string, vlanId []string) {
	body := strings.NewReader(fmt.Sprintf(`{"portId": %s, "portType": "ACCESS", "vlanIds": [%s]}`, port, makeVlanArrayString(vlanId)))
	res, err := makePost(protoOpt[1], body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	resbody, _ := ioutil.ReadAll(res.Body)
	printMessage(resbody)
}

func vlanSetTagged(port string, vlanIds []string) {
	body := strings.NewReader(fmt.Sprintf(`{"portId": %s, "portType": "TAGGED", "vlanIds": [%s]}`, port, makeVlanArrayString(vlanIds)))
	res, err := makePost(protoOpt[1], body)
	if err != nil {
		log.Fatal(err)
	}
	resbody, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	printMessage(resbody)
}

func vlanDelete(vlanId []string, portId string, choice int) {
	switch choice {
	case 0:
		// delete vlan instance
		for _, v := range vlanId {
			body := strings.NewReader(fmt.Sprintf(`{"vlanId": %s, "port": "%s", "choice": %d}`, v, 0, 0))
			res, err := makePost(protoOpt[2], body)
			if err != nil {
				log.Fatal(err)
				return
			}
			resbody, _ := ioutil.ReadAll(res.Body)
			printMessage(resbody)
			res.Body.Close()
		}
	case 1:
		// delete portId from vlan instance by portId
		ports := rangeSplit(portId)
		for _, p := range ports {
			body := strings.NewReader(fmt.Sprintf(`{"vlanId": %s, "portId": "%s", "choice": %d}`, vlanId[0], p, 1))
			res, err := makePost(protoOpt[2], body)
			if err != nil {
				log.Fatal(err)
				return
			}
			resbody, _ := ioutil.ReadAll(res.Body)
			printMessage(resbody)
			res.Body.Close()
		}
	default:
		return
	}
}

func vlanShow(data interface{}, choice int ) {
	switch choice {
	case 0:
		for _, v := range data.([]string) {
			value, _ := strconv.Atoi(v)
			body := strings.NewReader(fmt.Sprintf(`{"vlanId": %d, "portId": %d, "choice": %d}`, value, 0, 0))
			res, err := makePost(protoOpt[3], body)
			if err != nil {
				log.Fatal(err)
				return
			}
			resbody, _ := ioutil.ReadAll(res.Body)
			printMessage(resbody)
			res.Body.Close()
		}
		break
	case 1:
		value, _ := strconv.Atoi(data.(string))
		body := strings.NewReader(fmt.Sprintf(`{"vlanId": %d, "portId": %d, "choice": %d}`, 0, value, 1))
		res, err := makePost(protoOpt[3], body)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer res.Body.Close()

		resbody, _ := ioutil.ReadAll(res.Body)
		printMessage(resbody)
		break
	case 2:
		// show all of the vlan information
		body := strings.NewReader(fmt.Sprintf(`{"vlanId": %d, "portId": %d, "choice": %d}`, 0, 0, 2))
		res, err := makePost(protoOpt[3], body)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer res.Body.Close()

		resbody, _ := ioutil.ReadAll(res.Body)
		printMessage(resbody)
		break
	default:
		return
	}
}
func printMessage(msg []byte) {
	var message reponseMessage
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(message.Message)
}
// rangeSplit function split the port range value into two values,
// the one is lower where it needs to start to add the port number,
// and the other is upper where it needs to end adding the port number.
func rangeSplit(portsRange string) []string {
	spt := strings.Split(portsRange, "-")
	ret := make([]string, 0)
	var lower, upper int
	var err error
	// Check string value is digit.
	if len(spt) > 1 {
		// Take range ports
		if lower, err = strconv.Atoi(spt[0]); err != nil {
			log.Fatal(err)
		}
		if upper, err = strconv.Atoi(spt[1]); err != nil {
			log.Fatal(err)
		}
		if lower > upper {
			tmp := lower
			lower = upper
			upper = tmp
		}
		for i := lower; i <= upper; i++ {
			ret = append(ret, strconv.Itoa(i))
		}
	} else {
		// Take only one port
		if _, err = strconv.Atoi(spt[0]); err != nil {
			log.Fatal(err)
		}
		ret = append(ret, spt[0])
	}
	return ret
}

func makeVlanArrayString(str []string) string {
	ret := ""
	for k, v := range str {
		if k == 0 {
			ret = v
		} else {
			ret = ret + "," + v
		}
	}
	return ret
}