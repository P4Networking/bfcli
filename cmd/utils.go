package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/pisc/util/enums"
	"github.com/P4Networking/proto/go/p4"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)
//MAC_TYPE for Exact match
//IP_TYPE for Exact match
//VALUE_TYPE for Exact match
//CIDR_TYPE for LPM match
//MASK_TYPE for Ternary match
//NON_TYPE for unexpect value comes in
const (
	MAC_TYPE   = iota //0
	IP_TYPE
	CIDR_TYPE
	MASK_TYPE
	VALUE_TYPE
	NON_TYPE //5

	INT8 //6
	INT16
	INT32
	INT64//9

	IP_MASK//10
	ETH_MASK
	HEX_MASK
	VALUE_MASK//13
)

type MatchSet struct {
	matchType uint
	bitWidth int
}

var (
	FOUND        = true
	NOT_FOUND    = false
	DEFAULT_ADDR = ":50000"
	p4Info       util.BfRtInfoStruct
	nonP4Info    util.BfRtInfoStruct
)

func initConfigClient() (*p4.BfRuntimeClient, *context.Context, *grpc.ClientConn, context.CancelFunc, *util.BfRtInfoStruct, *util.BfRtInfoStruct) {
	if server == "" {
		server = DEFAULT_ADDR
	}
	conn, err := grpc.Dial(server, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	cli := p4.NewBfRuntimeClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithCancel(context.Background())

	rsp, err := cli.GetForwardingPipelineConfig(ctx, &p4.GetForwardingPipelineConfigRequest{DeviceId: uint32(77)})
	if err != nil {
		log.Fatalf("Error with", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.Config[0].BfruntimeInfo)).Decode(&p4Info)
	if err != nil {
		log.Fatal("decode error 1:", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.NonP4Config.BfruntimeInfo)).Decode(&nonP4Info)
	if err != nil {
		log.Fatal("decode error 2:", err)
	}
	return &cli, &ctx, conn, cancel, &p4Info, &nonP4Info
}

func printNameById(id uint32) (string, bool) {
	var name string
	var ok bool

	name, ok = p4Info.SearchActionNameById(id)
	if ok == true {
		return name, FOUND
	}

	name, ok = p4Info.SearchDataNameById(id)
	if ok == true {
		return name, FOUND
	}

	name, ok = p4Info.SearchActionParameterNameById(id)
	if ok == true {
		return name, FOUND
	}

	return "",NOT_FOUND
}

func collectTableMatchTypes(table *util.Table) (map[int]MatchSet, bool) {
	m := make(map[int]MatchSet)
	for _, v := range table.Key {
		if v.ID == 65537 || v.Name == "$MATCH_PRIORITY" {
			continue
		}
		bw := checkBitWidth(v.Type.Width)
		switch v.MatchType {
		case "Exact" :
			m[v.ID] = MatchSet{matchType: enums.MATCH_EXACT, bitWidth: bw}
		case "LPM":
			m[v.ID] = MatchSet{matchType: enums.MATCH_LPM, bitWidth: bw}
		case "Range":
			m[v.ID] = MatchSet{matchType: enums.MATCH_RANGE, bitWidth: bw}
		case "Ternary":
			m[v.ID] = MatchSet{matchType: enums.MATCH_TERNARY, bitWidth: bw}
		default:
			return nil, NOT_FOUND
		}
	}
	return m, FOUND
}

func collectActionFieldIds(table *util.Table, id uint32) map[uint32]int {
	result := make(map[uint32]int)
	for _, v := range table.ActionSpecs {
		if v.ID == id {
			for _, d := range v.Data {
				result[d.ID] = checkBitWidth(d.Type.Width)
			}
			break
		}
	}
	return result
}

func IsNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func checkMatchListType(value string) uint {
	if strings.ContainsAny(value, ":") {
		return MAC_TYPE
	} else if strings.ContainsAny(value, ".") {
		return IP_TYPE
	} else if c, n, _ := net.ParseCIDR(value); (c != nil && n != nil) || strings.Contains(value, "/") {
		return CIDR_TYPE
	} else if checkMaskType(value) != -1 {
		return MASK_TYPE
	}else if IsNumber(value) {
		return VALUE_TYPE
	}
	return NON_TYPE
}
/*
Four case in checkMask
case 1 : both arguments are IP presentation
case 2 : both arguments are Ether presentation
case 2 : both arguments are hex presentation (for ethertype, etc...)
case 3 : both arguments are integer
Caution : out of range of the cases are not handled.
*/
func checkMaskType(value string) int {
	if strings.Contains(value, "/") {
		arg := strings.Split(value, "/")
		m1, _ := net.ParseMAC(arg[0])
		m2, _ := net.ParseMAC(arg[1])
		i1 := net.ParseIP(arg[0])
		i2 := net.ParseIP(arg[1])
		if i1 != nil && i2 != nil{
			return IP_MASK
		} else if m1 != nil && m2 != nil{
			return ETH_MASK
		} else if strings.Contains(arg[0], "0x") && strings.Contains(arg[1], "0x") {
			return HEX_MASK
		} else if IsNumber(arg[0]) && IsNumber(arg[1]) {
			return VALUE_MASK
		}
	}
	return -1
}

func checkBitWidth(value int) int {
	if value > 32 {
		return INT64
	} else if (value > 16) && (value <= 32) {
		return INT32
	} else if (value > 8) && (value <= 16) {
		return INT16
	} else {
		return INT8
	}
}

func ParseBitWidth(value int, bitWidth int) []byte {
	var result []byte
	switch bitWidth {
	case INT32:
		result = util.Int32ToBytes(uint32(value))
	case INT16:
		result = util.Int16ToBytes(uint16(value))
	case INT8:
		result = util.Int8ToBytes(uint8(value))
	}
	return result
}

func Ipv4ToBytes(value string) []byte {
	nameSplit := strings.Split(value, ".")
	if len(nameSplit) > 4 || len(nameSplit) < 1 {
		return nil
	}
	result := make([]byte, len(nameSplit))
	for k, v := range nameSplit {
		value, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		result[k] = byte(value)
	}
	return result
}

func MacToBytes(macAddress string) []byte {
	macSplit := strings.Split(macAddress, ":")
	if len(macSplit) > 5 || len(macSplit) < 1 {
		return nil
	}
	result := make([]byte, len(macSplit))
	for k, v := range macSplit {
		data, err := strconv.ParseInt(v, 16, 64)
		if err != nil {
			panic(err)
		}
		result[k] = byte(data)
	}
	return result
}


func dumpEntries(stream *p4.BfRuntime_ReadClient, table *util.Table) {
	for {
		rsp, err := (*stream).Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Got error: %v", err)
		}
		fmt.Println("--------------------------------------------------------------------------------")
		Entities := rsp.GetEntities()
		if len(Entities) == 0 {
			fmt.Printf("The \"%s\" table is empty\n", table.Name)
		}
		for k, v := range Entities {
			tbl := v.GetTableEntry()
			if !tbl.IsDefaultEntry {
				fmt.Printf("Entry %d:\n", k)
			}
			fmt.Println("Match Key Info")
			if tbl.GetKey() != nil {
				fmt.Printf("  %-20s %-10s %-16s\n", "Field Name:", "Type:", "Value:")
				for k, f := range tbl.Key.Fields {
					if f.FieldId == 65537 {
						continue
					}
					switch strings.Split(reflect.TypeOf(f.GetMatchType()).String(), ".")[1] {
					case "KeyField_Exact_":
						m := f.GetExact()
						fmt.Printf("  %-20s %-10s %-16x\n", table.Key[k].Name, "Exact" ,m.Value)
					case "KeyField_Ternary_":
						t := f.GetTernary()
						fmt.Printf("  %-20s %-10s %-16x Mask: %-12x\n", table.Key[k].Name, "Ternay" ,t.Value, t.Mask)
					case "KeyField_Lpm":
						l := f.GetLpm()
						fmt.Printf("  %-20s %-10s %-16x PreFix: %-12d\n", table.Key[k].Name, "LPM" ,l.Value, l.PrefixLen)
					case "KeyField_Range_":
						//TODO: Implement range match
						r := f.GetRange()
						fmt.Printf("  %-20s %-10s %-16x High: %-6x Low: %-6x\n", table.Key[k].Name, "LPM" ,r.High, r.Low)
					}
				}
			}

			if tbl.IsDefaultEntry {
				fmt.Printf("Table default action:\n")
			}

			actionName, _ := printNameById(tbl.Data.ActionId)
			fmt.Println("Action:", actionName)

			if tbl.Data.Fields != nil {
				fmt.Printf("  %-20s %-16s\n", "Field Name:", "Value:")
				for _, d := range tbl.Data.Fields {
					actionFieldName, _ := printNameById(d.FieldId)
					fmt.Printf("  %-20s %-16x\n",actionFieldName, d.GetStream())
				}
			}
			if(k+1!=len(Entities)){
				fmt.Printf("------------------\n")
			}
		}
		fmt.Println("--------------------------------------------------------------------------------")
	}
}


func GenReadRequestWithId(tableId uint32) *p4.ReadRequest {
	req := &p4.ReadRequest{
		Entities: []*p4.Entity{
			{
				Entity: &p4.Entity_TableEntry{
					TableEntry: &p4.TableEntry{
						TableId: tableId,
					},
				},
			},
		},
	}
	return req
}
