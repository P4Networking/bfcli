package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/breezestars/bfruntime/go/p4"
	"github.com/breezestars/go-bfrt/util"
	"google.golang.org/grpc"
	"log"
)

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

func printNameById(id uint32) bool {
	var name string
	var ok bool

	name, ok = p4Info.SearchActionNameById(id)
	if ok == true {
		fmt.Printf("Action Name: %s \n", name)
		return FOUND
	}

	name, ok = nonP4Info.SearchActionParameterNameById(id)
	if ok == true {
		fmt.Printf("Action Parameter Name: %s \n", name)
		return FOUND
	}

	name, ok = p4Info.SearchDataNameById(id)
	if ok == true {
		fmt.Printf("Data Name: %s \n", name)
		return FOUND
	}
	return NOT_FOUND
}
