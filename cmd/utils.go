package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/P4Networking/pisc/southbound/bfrt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/pisc/util/enums"
	"github.com/P4Networking/pisc/util/enums/id"
	"github.com/P4Networking/proto/go/p4"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"
	"io"
	"log"
	"math"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// MAC_TYPE for Exact match
// IP_TYPE for Exact match
// VALUE_TYPE for Exact match
// CIDR_TYPE for LPM match
// MASK_TYPE for Ternary match
// NON_TYPE for unexpect value
const (
	MAC_TYPE = iota //0
	IP_TYPE
	CIDR_TYPE
	MASK_TYPE
	VALUE_TYPE
	HEX_TYPE
	NON_TYPE //6

	INT8 //7
	INT16
	INT32
	INT64 //10

	IP_MASK //11
	ETH_MASK
	HEX_MASK
	VALUE_MASK //14
)

type MatchSet struct {
	fieldId    uint32
	matchValue string
	matchType  uint
	bitWidth   int
}
type ActionSet struct {
	fieldId        uint32
	actionValue    string
	bitWidth       int
	parsedBitWidth int
}

// Struct Embedding gives the capability to expand its own new method with an existing type.
type BfRtInfo struct {
	bfrt.BfRtInfo
}

type EtnryObj struct {

	p4Info      BfRtInfo
	nonP4Info   BfRtInfo

	table []bfrt.Table

	actions map[uint32] string
	actionName string
	actionId uint32
}

var (
	preFixIg = "pipe.SwitchIngress"
	preFixEg = "pipe.SwitchEgress"
	preFixIgPar = "pipe.SwitchIngressParser"
	preFixEgPar = "pipe.SwitchEgressParser"
	DEFAULT_ADDR = ":50000"

	// true : not support to read; false : support to read
	NotSupportToReadTable = map[uint32] bool {
		2432705822 : true, // "pipe.$SNAPSHOT_EGRESS"
		2449483038 : true, // "pipe.$SNAPSHOT_EGRESS_LIVENESS"
		2432703071 : true, // "pipe.$SNAPSHOT_INGRESS"
		2449480287 : true, // "pipe.$SNAPSHOT_INGRESS_LIVENESS"
		2415950168 : true} // "pipe.SwitchIngressParser.$PORT_METADATA"

	Obj          EtnryObj
)

func ObjInit() {
	Obj.actionId = 0
	Obj.actionName = ""
	Obj.table = []bfrt.Table{}
	Obj.actions = make(map[uint32]string, 0)
}

func initConfigClient() (*p4.BfRuntimeClient, *context.Context, *grpc.ClientConn, context.CancelFunc, *BfRtInfo, *BfRtInfo) {
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
	rsp, err := cli.GetForwardingPipelineConfig(ctx, &p4.GetForwardingPipelineConfigRequest{DeviceId: uint32(0)})
	if err != nil {
		log.Fatalf("Error with %v", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.Config[0].BfruntimeInfo)).Decode(&Obj.p4Info)
	if err != nil {
		log.Fatal("decode error 1:", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.NonP4Config.BfruntimeInfo)).Decode(&Obj.nonP4Info)
	if err != nil {
		log.Fatal("decode error 2:", err)
	}
	return &cli, &ctx, conn, cancel, &Obj.p4Info, &Obj.nonP4Info
}

func PrintNameById(actionName string, id uint32) (string, bool) {
	var name string
	var ok bool

	name, ok = Obj.p4Info.GetActionNameById(id)
	if ok == true {
		return name, true
	}

	name, ok = Obj.p4Info.GetDataNameById(id)
	if ok == true {
		return name, true
	}

	name, ok = Obj.p4Info.GetActionParameterNameById(actionName, id)
	if ok == true {
		return name, true
	}

	return "", false
}

// collectTableMatchTypes function collects the match key's type and bit width.
// the collected type and bit width determine what kind of the input should be used.
func collectTableMatchTypes(matchKey *[]string) ([]MatchSet, bool) {
	table := Obj.table[0]
	if len(table.Key) != len(*matchKey) {
		fmt.Printf("expected match key length : %d,  received match key length : %d\n", len(table.Key), len(*matchKey))
		return nil, false
	}

	var m []MatchSet
	for kv, v := range table.Key {
		bw := parseBitWidth(v.Type.Width)
		switch v.MatchType {
		case "Exact":
			if v.ID == 65537 || v.Name == "$MATCH_PRIORITY" {
				m = append(m, MatchSet{fieldId: v.ID, matchValue: (*matchKey)[kv], matchType: enums.MATCH_EXACT, bitWidth: INT32})
			} else {
				m = append(m, MatchSet{fieldId: v.ID, matchValue: (*matchKey)[kv], matchType: enums.MATCH_EXACT, bitWidth: bw})
			}
		case "LPM":
			m = append(m, MatchSet{fieldId: v.ID, matchValue: (*matchKey)[kv], matchType: enums.MATCH_LPM, bitWidth: bw})
		case "Range":
			m = append(m, MatchSet{fieldId: v.ID, matchValue: (*matchKey)[kv], matchType: enums.MATCH_RANGE, bitWidth: bw})
		case "Ternary":
			m = append(m, MatchSet{fieldId: v.ID, matchValue: (*matchKey)[kv], matchType: enums.MATCH_TERNARY, bitWidth: bw})
		default:
			return nil, false
		}
	}
	return m, true
}

// collectActionFieldIds function collects the bit width of action data.
// the collected bit widths decide what kind of input should be used.
func collectActionFieldIds(table *bfrt.Table, id uint32, values []string) ([]ActionSet, error) {
	var result []ActionSet
	for _, v := range table.ActionSpecs {
		if v.ID == id {
			for kd, d := range v.Data {
				if IsNumber(values[kd]) {
					value, err := strconv.Atoi(values[kd])
					if int(math.Pow(2, float64(d.Type.Width))-1) < value {
						err = fmt.Errorf("inputted action value is overflow: %s\n", values[kd])
						return nil, err
					}
				}
				result = append(result, ActionSet{fieldId: d.ID, actionValue: values[kd], bitWidth: d.Type.Width, parsedBitWidth: parseBitWidth(d.Type.Width)})
			}
			break
		}
	}
	return result, nil
}

// IsNumber function check that the string is numeric.
func IsNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// checkMatchListType function checks that the input arguments are what kind of the match values.
func checkMatchListType(value string, bitWidth int) (uint, interface{}, interface{}) {
	if _, e := net.ParseMAC(value); e == nil && bitWidth == INT64 {
		return MAC_TYPE, util.MacToBytes(value), nil
	} else if e := parseIPv4(value); (strings.IndexByte(value, '/') < 0) && (e != nil) && (bitWidth <= INT32) {
		return IP_TYPE, e, nil
	} else if i, a, e := parseCIDR(value); e && strings.IndexByte(value, '/') >= 0 {
		return CIDR_TYPE, i, a
	} else if maskType, arg := checkMaskType(value); maskType != -1 && arg != nil {
		return MASK_TYPE, maskType, arg
	} else if strings.HasPrefix(strings.ToLower(value), "0x") {
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

// checkMaskType function checks that the input arguments have what kind of the mask types.
// checkMaskType function only work with the argument what it has "/" character.
func checkMaskType(value string) (int, interface{}) {
	if strings.Count(value, "/") == 1 {
		var i = strings.IndexByte(value, '/')
		m1, _ := net.ParseMAC(value[:i])
		m2, _ := net.ParseMAC(value[i+1:])
		if m1 != nil && m2 != nil && (len(m1) == len(m2)) {
			return ETH_MASK, []string{value[:i], value[i+1:]}
		}
		i1 := net.ParseIP(value[:i])
		i2 := net.ParseIP(value[i+1:])
		if i1 != nil && i2 != nil && (len(i1) == len(i2)) {
			return IP_MASK, []string{value[:i], value[i+1:]}
		}
		if strings.HasPrefix(strings.ToLower(value[:i]), "0x") &&
			strings.HasPrefix(strings.ToLower(value[i+1:]), "0x") {
			c1, _ := strconv.ParseUint(value[:i], 0, 16)
			c2, _ := strconv.ParseUint(value[i+1:], 0, 16)
			return HEX_MASK, []uint16{uint16(c1), uint16(c2)}
		}
		if IsNumber(value[:i]) && IsNumber(value[i+1:]) {
			a1, _ := strconv.Atoi(value[:i])
			a2, _ := strconv.Atoi(value[i+1:])
			return VALUE_MASK, []int{a1, a2}
		}
	}
	return -1, nil
}

// parseBitWidth function determines that the input value is what kind of sizes.
func parseBitWidth(value int) int {
	if (value > 32) && (value <= 64) {
		return INT64
	} else if (value > 16) && (value <= 32) {
		return INT32
	} else if (value > 8) && (value <= 16) {
		return INT16
	} else if (value > 0) && (value <= 8) {
		return INT8
	} else {
		return -1
	}
}

// setBitValue function based on the input bit width to decide what size of the function should be used.
// If the function can't determine the type of the bit width, it will return nil.
func setBitValue(value int, bitWidth int) []byte {
	var result []byte
	switch bitWidth {
	case INT32:
		result = util.Int32ToBytes(uint32(value))
	case INT16:
		result = util.Int16ToBytes(uint16(value))
	case INT8:
		result = util.Int8ToBytes(uint8(value))
	default:
		return nil
	}
	return result
}

// parseCIDR function copies from net packages.
// parseCIDR function checking the input string has a "/" character or not.
// If does, the function parses the string and checks that the string is CIDR form.
// parseCIDR function only used when the collectedMatchType function has LPM type.
func parseCIDR(s string) (interface{}, int, bool) {
	i := strings.IndexByte(s, '/')
	if i < 0 {
		return nil, -1, false
	}
	addr, mask := s[:i], s[i+1:]
	ip := parseIPv4(addr)
	n, i, ok := dtoi(mask)
	if ip == nil || !ok || i != len(mask) || n < 0 || n > 32 {
		return nil, -1, false
	}
	m, _ := strconv.Atoi(mask)
	return ip, m, true
}

// dtoi function copies from net package
// dtoi function checks character is numeric and makes it to integer than return.
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

// parseIPv4 function copy form net package
// parseIPv4 function checks that the string is IPv4 form.
// the string must have three cases : "x.x", "x.x.x", "x.x.x.x"
func parseIPv4(s string) []byte {
	var size = strings.Count(s, ".")
	if size <= 0 || size > 3 {
		return nil
	}
	size++
	p := make([]byte, size)
	for i := 0; i < size; i++ {
		if len(s) == 0 {
			goto ERROR
		}
		if i > 0 {
			if s[0] != '.' {
				goto ERROR
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			goto ERROR
		}
		s = s[c:]
		p[i] = byte(n)
	}
	if len(s) != 0 {
		goto ERROR
	}
	return p
ERROR:
	return nil
}

func InfoEntries(table bfrt.Table) {

	result := fmt.Sprintln("--------------------------------------------------------------------------------")
	result += "Table Info\n"
	if table.Name != "" {
		result += fmt.Sprintf("  %-12s: %-6s\n", "Name", table.Name)
	}
	if table.ID != 0 {
		result += fmt.Sprintf("  %-12s: %-6d\n", "ID", table.ID)
	}
	if table.TableType != "" {
		result += fmt.Sprintf("  %-12s: %-6s\n", "Type", table.TableType)
	}
	if table.Size != 0 {
		result += fmt.Sprintf("  %-12s: %-6d\n", "Size", table.Size)
	}
	if table.Annotations != nil {
		result += fmt.Sprintf("  %-12s:\n", "Annotations")
		for k, v := range table.Annotations {
			if v.Value != "" {
				result += fmt.Sprintf("  %d - Name: %s | Value: %s \n", k+1, v.Name, v.Value)
			} else {
				result += fmt.Sprintf("    %d - Name: %s \n", k+1, v.Name)
			}
		}
	}
	if table.DependsOn != nil {
		result += fmt.Sprintf("%-12s: %-6s\n", "Table DependsOn", table.DependsOn)
	}

	if table.Key != nil {
		result += fmt.Sprintln("--------------------------------------------------------------------------------")
		result += fmt.Sprintf("%-s\n", "Match Key Info")
		result += fmt.Sprintf("  %-8s %-20s %-11s %-10s %-9s %-8s %-4s",
			"KeyId", "Name", "Match_type", "Mandatory", "Repeated", "Type", "Width\n")
		for _, v := range table.Key {
			if v.Name == "$MATCH_PRIORITY" {
				result += fmt.Sprintf("  %-8d %-20s %-11s %-10t %-9t %-8s %-4d\n",
					v.ID, v.Name, v.MatchType, v.Mandatory, v.Repeated, v.Type.Type, 32)
			} else {
				result += fmt.Sprintf("  %-8d %-20s %-11s %-10t %-9t %-8s %-4d\n",
					v.ID, v.Name, v.MatchType, v.Mandatory, v.Repeated, v.Type.Type, v.Type.Width)
			}
		}
	}

	if table.Data != nil {
		result += fmt.Sprintln("--------------------------------------------------------------------------------")
		result += fmt.Sprintf("%-s\n", "Table Data Info")
		result += fmt.Sprintf("  %-8s %-20s %-10s %-9s %-8s\n",
			"KeyId", "Name", "Mandatory", "Repeated", "Type")
		for _, v := range table.Data {
			result += fmt.Sprintf("  %-8d %-20s %-10t %-9t %-8s\n",
				v.Singleton.ID, v.Singleton.Name, v.Mandatory, v.Singleton.Repeated, v.Singleton.Type.Type)
		}
	}

	if table.ActionSpecs != nil {
		result += fmt.Sprintln("--------------------------------------------------------------------------------")
		result += fmt.Sprintf("%-s\n", "Action Info")
		for _, v := range table.ActionSpecs {
			result += fmt.Sprintf("  ID: %-6d, Name: %-20s \n", v.ID, v.Name)
			if v.Data != nil {
				result += fmt.Sprintln("    ----------------------------------------------------------------")
				for _, d := range v.Data {
					result += fmt.Sprintf("    %-6s %-20s %-10s %-9s %-8s %-4s\n",
						"ID", "Name", "Mandatory", "Repeated", "Type", "Width")
					result += fmt.Sprintf("    %-6d %-20s %-10t %-9t %-8s %-4d\n",
						d.ID, d.Name, d.Mandatory, d.Repeated, d.Type.Type, d.Type.Width)
				}
				result += fmt.Sprintln("    ----------------------------------------------------------------")
			}
		}
		result += fmt.Sprintln("--------------------------------------------------------------------------------")
	}
	fmt.Print(result)
}

// DumpEntries function reads the entries from PISC via ReadRequest.
func DumpEntries(stream *p4.BfRuntime_ReadClient, p4table *bfrt.Table) {
	for {
		mkWrt := tablewriter.NewWriter(os.Stdout)
		mkWrt.SetAlignment(tablewriter.ALIGN_CENTER)

		adWrt:= tablewriter.NewWriter(os.Stdout)
		adWrt.SetAlignment(tablewriter.ALIGN_CENTER)

		rsp, err := (*stream).Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Got error: %v", err)
			return
		}
		Entities := rsp.GetEntities()
		if len(Entities) == 0 {
			fmt.Printf("The \"%s\" table is empty\n", p4table.Name)
			return
		} else {
			if count {
				fmt.Printf("Table \"%s\" has %d entries\n", p4table.Name, len(Entities))
			} else {
				str := "--------------------------------------------------------------------------------\n"
				str += fmt.Sprintf("Table Name : %s", p4table.Name)
				fmt.Println(str)
				mkWrt.SetHeader([]string{"Field Name", "Type", "Value", "Mask/PreFix"})
				for kv, v := range Entities {
					tbl := v.GetTableEntry()
					fmt.Println(fmt.Sprintf("\nEntry : %d\nMatch Key Info", kv))
					// Match Keys
					if tbl.GetKey() != nil {
						iter := 0
						for k, mk := range p4table.Key {
							f := tbl.Key.Fields[iter]
							if mk.ID == f.FieldId {
								switch strings.Split(reflect.TypeOf(f.GetMatchType()).String(), ".")[1] {
								case "KeyField_Exact_":
									m := f.GetExact()
									mkWrt.Append([]string{p4table.Key[k].Name, "Exact", hex.EncodeToString(m.Value), "None"})
								case "KeyField_Ternary_":
									t := f.GetTernary()
									mkWrt.Append([]string{p4table.Key[k].Name, "Ternary", hex.EncodeToString(t.Value), hex.EncodeToString(t.Mask)})
								case "KeyField_Lpm":
									l := f.GetLpm()
									mkWrt.Append([]string{p4table.Key[k].Name, "LPM", string(l.Value), string(l.PrefixLen)})
								//	TODO : Range match field
								//case "KeyField_Range_":
									//	//r := f.GetRange()
									//	//fmt.Printf("  %-20s %-10s High: %-8x Low: %-8x\n", p4table.Key[k].Name, "LPM", r.High, r.Low)
								}
								iter++
							} else {
								mkWrt.Append([]string{p4table.Key[k].Name, "None", "None", "None"})
							}
						}
						mkWrt.Render()
						mkWrt.ClearRows()
					}

					// Action Field
					actionName, _ := Obj.p4Info.GetActionNameById(tbl.Data.ActionId)
					fmt.Println("Action:", actionName)
					if tbl.Data.Fields != nil {
						adWrt.SetHeader([]string{"Field Name", "Value"})
						for _, datafield := range tbl.Data.Fields {
							an, ok := PrintNameById(actionName, datafield.FieldId)
							if ok {
								adWrt.Append([]string{an, hex.EncodeToString(datafield.GetStream())})
							}
						}
						adWrt.Render()
						adWrt.ClearRows()
					}
				}
				fmt.Println("--------------------------------------------------------------------------------")
			}
		}
	}
}

// genReadRequestWithId function generates the read request to read table's entries.
func genReadRequestWithId(tableId uint32) *p4.ReadRequest {
	var req = &p4.ReadRequest{}
	req.Entities = append(req.Entities, genEntity(tableId))
	return req
}

func genEntity(tableId uint32) *p4.Entity {
	return &p4.Entity{
		Entity: &p4.Entity_TableEntry{
			TableEntry: &p4.TableEntry{
				TableId: tableId,
			},
		},
	}
}

// DeleteEntries function read entries from the response to delete all entries of a table
func DeleteEntries(rsp **p4.ReadResponse, cli *p4.BfRuntimeClient, ctx *context.Context) ([]int, error) {
	var result []int
	var delReq *p4.WriteRequest = nil
	for k, e := range (*rsp).Entities {
		tbl := e.GetTableEntry()
		if delReq == nil {
			delReq = util.GenWriteRequestWithId(p4.Update_DELETE, id.TableId(tbl.TableId), tbl.Key.Fields, nil)
		} else {
			delReq.Updates = append(delReq.Updates, &p4.Update{
				Type: p4.Update_DELETE,
				Entity: &p4.Entity{
					Entity: &p4.Entity_TableEntry{
						TableEntry: &p4.TableEntry{
							TableId: tbl.TableId,
							Key: &p4.TableKey{
								Fields: tbl.Key.Fields,
							},
							Data: nil,
						},
					},
				},
			})
		}
		result = append(result, k)
	}
	_, err := (*cli).Write(*ctx, delReq)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BuildMatchKeys function using the input arguments make the KeyField what a table need to match.
func BuildMatchKeys(collectedMatchTypes *[]MatchSet) []*p4.KeyField {
	match := util.Match()
	for _, v := range *collectedMatchTypes {
		mlt, v1, v2 := checkMatchListType(v.matchValue, v.bitWidth)
		// In EXACT case, v2 value is always nil.
		if v.matchType == enums.MATCH_EXACT {
			switch mlt {
			case MAC_TYPE:
				match = append(match, util.GenKeyField(v.matchType, v.fieldId, v1.([]byte)))
			case IP_TYPE:
				match = append(match, util.GenKeyField(v.matchType, v.fieldId, v1.([]byte)))
			case VALUE_TYPE:
				match = append(match, util.GenKeyField(v.matchType, v.fieldId, setBitValue(v1.(int), v.bitWidth)))
			case HEX_TYPE:
				match = append(match, util.GenKeyField(v.matchType, v.fieldId, util.HexToBytes(uint16(v1.(uint64)))))
			default:
				fmt.Printf("Unexpect value for EXACT_MATCH : %s\n", v.matchValue)
				return nil
			}
		} else if v.matchType == enums.MATCH_LPM {
			if mlt == CIDR_TYPE {
				match = append(match, util.GenKeyField(v.matchType, v.fieldId, v1.([]byte), v2.(int)))
			} else {
				fmt.Printf("Unexpect value for LPM_MATCH : %s\n", v.matchValue)
				return nil
			}
		} else if v.matchType == enums.MATCH_TERNARY {
			//Ternary match only support the complete address format(aa:aa:aa:aa:aa:aa/ff:ff:ff:ff:ff:ff, x.x.x.x/255.255.255.255)
			if mlt == MASK_TYPE {
				switch v1.(int) {
				case IP_MASK:
					arg := v2.([]string)
					match = append(match, util.GenKeyField(v.matchType, v.fieldId, util.Ipv4ToBytes(arg[0]), util.Ipv4ToBytes(arg[1])))
				case ETH_MASK:
					arg := v2.([]string)
					match = append(match, util.GenKeyField(v.matchType, v.fieldId, util.MacToBytes(arg[0]), util.MacToBytes(arg[1])))
				case HEX_MASK:
					arg := v2.([]uint16)
					match = append(match, util.GenKeyField(v.matchType, v.fieldId, util.HexToBytes(arg[0]), util.HexToBytes(arg[1])))
				case VALUE_MASK:
					arg := v2.([]int)
					match = append(match, util.GenKeyField(v.matchType, v.fieldId, setBitValue(arg[0], v.bitWidth), setBitValue(arg[1], v.bitWidth)))
				}
			} else {
				fmt.Printf("Unexpect value for TERNARY_MATCH : %s\n", v.matchValue)
				return nil
			}
		} else if v.matchType == enums.MATCH_RANGE {
			//TODO: Implement range match
			fmt.Println("Range_Match Not Supported Yet")
			return nil
		} else {
			fmt.Println("Unexpected Match Type")
		}
	}
	return match
}
