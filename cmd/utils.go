package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	bfrt "github.com/breezestars/go-bfrt/proto/out"
	"github.com/breezestars/go-bfrt/util"
	"google.golang.org/grpc"
	"log"
)

var (
	FOUND        = true
	NOT_FOUND    = false
	DEFAULT_ADDR = ":50000"
	p4           util.BfRtInfoStruct
	nonP4        util.BfRtInfoStruct
)

func initConfigClient() (*bfrt.BfRuntimeClient, *context.Context, *grpc.ClientConn, context.CancelFunc, *util.BfRtInfoStruct, *util.BfRtInfoStruct) {
	if server == "" {
		server = DEFAULT_ADDR
	}
	conn, err := grpc.Dial(server, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	cli := bfrt.NewBfRuntimeClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithCancel(context.Background())

	rsp, err := cli.GetForwardingPipelineConfig(ctx, &bfrt.GetForwardingPipelineConfigRequest{DeviceId: uint32(77)})
	if err != nil {
		log.Fatalf("Error with", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.Config[0].BfruntimeInfo)).Decode(&p4)
	if err != nil {
		log.Fatal("decode error 1:", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.NonP4Config.BfruntimeInfo)).Decode(&nonP4)
	if err != nil {
		log.Fatal("decode error 2:", err)
	}
	return &cli, &ctx, conn, cancel, &p4, &nonP4
}

func printNameById(id uint32) bool {
	var name string
	var ok bool

	name, ok = p4.SearchActionNameById(id)
	if ok == true {
		fmt.Printf("Action Name: %s \n", name)
		return FOUND
	}

	name, ok = p4.SearchActionParameterNameById(id)
	if ok == true {
		fmt.Printf("Action Parameter Name: %s \n", name)
		return FOUND
	}

	name, ok = p4.SearchDataNameById(id)
	if ok == true {
		fmt.Printf("Data Name: %s \n", name)
		return FOUND
	}
	return NOT_FOUND
}
