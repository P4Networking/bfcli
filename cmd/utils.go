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

//PISC-CLI use 50000 port basically
var (
	DEFAULT_ADDR = ":50000"
	found     = true
	not_found = false
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
		return name, found
	}

	name, ok = p4Info.SearchDataNameById(id)
	if ok == true {
		return name, found
	}

	return "",not_found
}


// DumpEntries function print entries data of the table.
// function will terminate when the stream occurs an error.
// also, it's not show the priority information.
func DumpEntries(stream *p4.BfRuntime_ReadClient, p4table *util.Table) {
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
							//r := f.GetRange()
							//fmt.Printf("  %-20s %-10s High: %-8x Low: %-8x\n", p4table.Key[k].Name, "LPM", r.High, r.Low)
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

// genReadRequestWithId function generates the read request format to read entries of the table via table ID.
func genReadRequestWithId(tableId uint32) *p4.ReadRequest {
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
