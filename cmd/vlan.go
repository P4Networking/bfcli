package cmd

import (
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
)

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
		// Create(vid string, vlanName string)
		// Add(type string, port string, vid []string)
		// Modify(port string, vid []string)
		// Delete(vName string)
		// Show()
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
			// VLAN Modify
			if !cmd.Flag("port").Changed || !cmd.Flag("id").Changed {
				fmt.Println("Check VLAN port and VLAN Id")
				_ = cmd.Help()
				return
			}
			opt[args[0]].(func(string, []string))(vlanPort, vlanId)

		case optSet[3]:
			// VLAN Delete
			if !cmd.Flag("id").Changed {
				fmt.Println("Check VLAN id")
				_ = cmd.Help()
				return
			}
			opt[args[0]].(func([]string))(vlanId)
		case optSet[4]:
			// VLAN show
			if !cmd.Flag("id").Changed && !cmd.Flag("port").Changed {
				fmt.Println("vlanId/portId is not set")
				_ = cmd.Help()
				return
			} else if cmd.Flag("id").Changed && !cmd.Flag("port").Changed {
				opt[args[0]].(func(interface{}, string))(vlanId, "vlan")
			} else if !cmd.Flag("id").Changed && cmd.Flag("port").Changed {
				opt[args[0]].(func(interface{}, string))(vlanPort, "port")
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
	optSet = []string{"create","add","modify","delete","show"}
	opt  = map[string]interface{} {
		optSet[0] : vlanCreate,
		optSet[1] : vlanAdd,
		optSet[2] : vlanModify,
		optSet[3] : vlanDelete,
		optSet[4] :	vlanShow,
	}
}

func vlanCreate(vid, vlanName string) {
	body := strings.NewReader(fmt.Sprintf(`{"vlanId": %s, "name": "%s"}`, vid, vlanName))
	res, err := http.Post("http://localhost:50101/v1/vlan/create","application/x-www-form-urlencoded", body)
	if err != nil {
		log.Fatal(err)
	}
	resbody, _ := ioutil.ReadAll(res.Body)
	log.Println(string(resbody))
	res.Body.Close()
}

func vlanAdd(vlanType string, port string, vid []string) {
	ports := rangeSplit(port)
	for _, port := range ports {
		switch vlanType {
		case "trunk":
			vlanSetTagged(port, vid)
			break
		case "access":
			if len(vid) > 1 {
				log.Panic("access port can only configure with 1 vlanId")
			}
			vlanSetUntag(port, vid)
			break
		default:
			log.Fatal("check vlan type what you inputted.")
		}
	}
}

func vlanSetUntag(port string, vlanId []string) {
	body := strings.NewReader(fmt.Sprintf(`{"portId": %s, "portType": "ACCESS", "vlanIds": [%s]}`, port, makeVlanArrayString(vlanId)))
	res, err := http.Post("http://localhost:50101/v1/vlan/portupdate","application/x-www-form-urlencoded", body)
	if err != nil {
		log.Fatal(err)
	}
	resbody, _ := ioutil.ReadAll(res.Body)
	log.Println(string(resbody))
	res.Body.Close()
}

func vlanSetTagged(port string, vlanIds []string) {
	body := strings.NewReader(fmt.Sprintf(`{"portId": %s, "portType": "TAGGED", "vlanIds": [%s]}`, port, makeVlanArrayString(vlanIds)))
	res, err := http.Post("http://localhost:50101/v1/vlan/portupdate","application/x-www-form-urlencoded", body)
	if err != nil {
		log.Fatal(err)
	}
	resbody, _ := ioutil.ReadAll(res.Body)
	log.Println(string(resbody))
	res.Body.Close()
}

func vlanModify(port string, vlanId []string) {
	// TODO: Need to implement, it just an example of config
	body := strings.NewReader(fmt.Sprintf(`{"portId": %s, "portType": "Tagged", "vlanIds": [%s]}`, port, makeVlanArrayString(vlanId)))
	res, err := http.Post("http://localhost:50101/v1/vlan/portupdate","application/x-www-form-urlencoded", body)
	if err != nil {
		log.Fatal(err)
		return
	}
	resbody, _ := ioutil.ReadAll(res.Body)
	log.Println(string(resbody))
	res.Body.Close()
}

func vlanDelete(vlanId []string) {
	for _, v := range vlanId {
		body := strings.NewReader(fmt.Sprintf(`{"vlanId": %s, "name": "%s"}`, v, ""))
		res, err := http.Post("http://localhost:50101/v1/vlan/delete","application/x-www-form-urlencoded", body)
		if err != nil {
			log.Fatal(err)
			return
		}
		resbody, _ := ioutil.ReadAll(res.Body)
		log.Println(string(resbody))
		res.Body.Close()
	}
}

func vlanShow(data interface{}, choice string ) {
	if choice == "vlan" {
		for _, v := range data.([]string) {
			value, _ := strconv.Atoi(v)
			body := strings.NewReader(fmt.Sprintf(`{"vlanId": %d, "portId": %d, "choice": "vlan"}`, value, 0))
			res, err := http.Post("http://localhost:50101/v1/vlan/get","application/x-www-form-urlencoded", body)
			if err != nil {
				log.Fatal(err)
				return
			}
			resbody, _ := ioutil.ReadAll(res.Body)
			log.Println(string(resbody))
			res.Body.Close()
		}
	} else {
		value, _ := strconv.Atoi(data.(string))
		body := strings.NewReader(fmt.Sprintf(`{"vlanId": %d, "portId": %d, "choice": "port"}`, 0, value))
		res, err := http.Post("http://localhost:50101/v1/vlan/get","application/x-www-form-urlencoded", body)
		if err != nil {
			log.Fatal(err)
			return
		}
		resbody, _ := ioutil.ReadAll(res.Body)
		log.Println(string(resbody))
		res.Body.Close()
	}
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