package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/P4Networking/proto/go/p4"
	"google.golang.org/grpc"
	"io"
	"log"
	"reflect"
	"strings"
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
							for k, data := range tbl.Data.Fields {
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
