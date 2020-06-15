package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	pb "github.com/breezestars/go-bfrt/proto"
	"github.com/breezestars/go-bfrt/util"
	"log"
	"google.golang.org/grpc"
	"context"
)

func initConfigClient() (pb.ConfigClient, context.Context, context.CancelFunc, util.BfRtInfoStruct, util.BfRtInfoStruct) {
	conn, err := grpc.Dial(":50000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	cli := pb.NewConfigClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithCancel(context.Background())

	rsp, err := cli.GetForwardingPipelineConfig(ctx, &pb.GetForwardingPipelineConfigRequest{DeviceId: uint32(77)})
	if err != nil {
		fmt.Printf("Error with", err)
	}

	var p4, nonP4 util.BfRtInfoStruct
	err = gob.NewDecoder(bytes.NewReader(rsp.Config[0].BfruntimeInfo)).Decode(&p4)
	if err != nil {
		log.Fatal("decode error 1:", err)
	}

	err = gob.NewDecoder(bytes.NewReader(rsp.NonP4Config.BfruntimeInfo)).Decode(&nonP4)
	if err != nil {
		log.Fatal("decode error 1:", err)
	}
	return cli, ctx, cancel, p4, nonP4
}
