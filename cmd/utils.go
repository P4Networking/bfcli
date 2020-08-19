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
	HEX_TYPE
	NON_TYPE //6

	INT8 //7
	INT16
	INT32
	INT64//10

	IP_MASK//11
	ETH_MASK
	HEX_MASK
	VALUE_MASK//14
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
		log.Fatalf("Error with %v", err)
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

	//name, ok = p4Info.SearchActionParameterNameById(id)
	//if ok == true {
	//	return name, FOUND
	//}

	return "",NOT_FOUND
}

func collectTableMatchTypes(table *util.Table) (map[int]MatchSet, bool) {
	m := make(map[int]MatchSet)
	for _, v := range table.Key {
		if v.ID == 65537 || v.Name == "$MATCH_PRIORITY" {
			continue
		}
		bw := parseBitWidth(v.Type.Width)
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
				result[d.ID] = parseBitWidth(d.Type.Width)
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

//
func checkMatchListType(value string) (uint, interface{}, interface{}) {
	if _, e := net.ParseMAC(value); e == nil {
		return MAC_TYPE, util.MacToBytes(value), nil
	} else if e := parseIPv4(value); strings.IndexByte(value, '/') < 0 && e != nil{
		return IP_TYPE, e, nil
	} else if i, a, e := ParseCIDR(value); e == 0{
		return CIDR_TYPE, i, a
	} else if maskType, arg  := checkMaskType(value); maskType != -1 && arg != nil {
		return MASK_TYPE, maskType, arg
	} else if strings.HasPrefix(strings.ToLower(value), "0x") {
		fmt.Println("hex")
		hexValue, e := strconv.ParseUint(value, 0, 16)
		return HEX_TYPE, hexValue, e
	} else if IsNumber(value) {
		arg, _ := strconv.Atoi(value)
		return VALUE_TYPE, arg, nil
	}
	return NON_TYPE, nil, nil
}
/*
Four case in checkMask
case 1 : both arguments are IP presentation
case 2 : both arguments are Ether presentation
case 2 : both arguments are hex presentation (for ethertype, etc...)
case 3 : both arguments are integer
Caution : out of range of the cases are not handled.
*/
func checkMaskType(value string) (int, interface{}) {
	if strings.Count(value, "/") == 1 {
		var i = strings.IndexByte(value, '/')
		m1, _ := net.ParseMAC(value[:i])
		m2, _ := net.ParseMAC(value[i+1:])
		if m1 != nil && m2 != nil{
			return ETH_MASK, []string{value[:i], value[i+1:]}
		}
		i1 := net.ParseIP(value[:i])
		i2 := net.ParseIP(value[i+1:])
		if i1 != nil && i2 != nil{
			return IP_MASK, []string{value[:i], value[i+1:]}
		}
		if strings.HasPrefix(strings.ToLower(value[:i]), "0x") &&
			strings.HasPrefix(strings.ToLower(value[i+1:]), "0x"){
			c1, _ := strconv.ParseUint(value[:i], 0, 16)
			c2, _ := strconv.ParseUint(value[i+1:], 0, 16)
			return HEX_MASK, []uint16{uint16(c1), uint16(c2)}
		}
		if IsNumber(value[:i]) && IsNumber(value[i+1:]) {
			a1,_ := strconv.Atoi(value[:i])
			a2,_ := strconv.Atoi(value[i+1:])
			return VALUE_MASK, []int{a1, a2}
		}
	}
	return -1, nil
}

func parseBitWidth(value int) int {
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

func setBitValue(value int, bitWidth int) []byte {
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

//Refactor net package's function
func ParseCIDR(s string) (interface{}, int, int) {
	i := strings.IndexByte(s, '/')
	if i < 0 {
		return nil, -1, -1
	}
	addr, mask := s[:i], s[i+1:]
	iplen := 4
	ip := parseIPv4(addr)
	n, i, ok := dtoi(mask)
	if ip == nil || !ok || i != len(mask) || n < 0 || n > 8*iplen {
		return nil, -1, -1
	}
	m, _ := strconv.Atoi(mask)
	return ip, m, 0
}
//Refactor net package's function
func dtoi(s string) (n int, i int, ok bool) {
	var big = 0xFFFFF
	n = 0
	for i = 0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return big, i, false
		}
	}
	if i == 0 {
		return 0, 0, false
	}
	return n, i, true
}
//Refactor net package's function
func parseIPv4(s string) []byte {
	var size = strings.Count(s, ".")
	p := make([]byte, size)
	for i := 0; i < size; i++ {
		if len(s) == 0 {
			// Missing octets.
			return nil
		}
		if i > 0 {
			if s[0] != '.' {
				return nil
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return nil
		}
		s = s[c:]
		p[i] = byte(n)
	}
	if len(s) != 0 {
		return nil
	}
	return p
}


func dumpEntries(stream *p4.BfRuntime_ReadClient, p4table *util.Table) {
	for {
		rsp, err := (*stream).Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Got error: %v", err)
		}
		Entities := rsp.GetEntities()
		if len(Entities) == 0 {
			fmt.Printf("The \"%s\" table is empty\n", p4table.Name)
		}else {
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Printf("Table Name : %-s\n", p4table.Name)
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
							fmt.Printf("  %-20s %-10s %-16x\n", p4table.Key[k].Name, "Exact", m.Value)
						case "KeyField_Ternary_":
							t := f.GetTernary()
							fmt.Printf("  %-20s %-10s %-16x Mask: %-12x\n", p4table.Key[k].Name, "Ternay", t.Value, t.Mask)
						case "KeyField_Lpm":
							l := f.GetLpm()
							fmt.Printf("  %-20s %-10s %-16x PreFix: %-12d\n", p4table.Key[k].Name, "LPM", l.Value, l.PrefixLen)
						case "KeyField_Range_":
							//TODO: Implement range match
							r := f.GetRange()
							//fmt.Printf("  %-20s %-10s %-16x High: %-8x Low: %-8x\n", table.Key[k].Name, "LPM", r.High, r.Low)
							fmt.Printf("  %-20s %-10s High: %-8x Low: %-8x\n", p4table.Key[k].Name, "LPM", r.High, r.Low)
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
					for _, d := range p4table.ActionSpecs {
						if d.Name == actionName {
							for _, data := range tbl.Data.Fields {
								if d.Data[k].ID == data.FieldId{
									fmt.Printf("  %-20s %-16x\n", d.Data[k].Name, data.GetStream())
								}
							}
						}
					}
				}
				if k+1 != len(Entities) {
					fmt.Printf("------------------\n")
				}
			}
			fmt.Println("--------------------------------------------------------------------------------")
		}
	}
}


func GenReadRequestWithId(tableId uint32) *p4.ReadRequest {
	return &p4.ReadRequest{
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
}
