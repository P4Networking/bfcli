/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dump called")

		//conn, err := grpc.Dial(":50000", grpc.WithInsecure(), grpc.WithBlock())
		//if err != nil {
		//	log.Fatalf("did not connect: %v", err)
		//}
		//defer conn.Close()
		//cli := pb.NewConfigClient(conn)
		//
		//// Contact the server and print out its response.
		//
		//ctx, cancel := context.WithCancel(context.Background())
		//defer cancel()
		//
		//rsp, err := cli.GetForwardingPipelineConfig(ctx, &pb.GetForwardingPipelineConfigRequest{DeviceId: uint32(77)})
		//if err != nil {
		//	fmt.Printf("Error with", err)
		//}
		//fmt.Printf("Got P4 name: %s \n", rsp.Config[0].P4Name)
		//
		//var p4, nonP4 util.BfRtInfoStruct
		//err = gob.NewDecoder(bytes.NewReader(rsp.Config[0].BfruntimeInfo)).Decode(&p4)
		//if err != nil {
		//	log.Fatal("decode error 1:", err)
		//}
		//
		//err = gob.NewDecoder(bytes.NewReader(rsp.NonP4Config.BfruntimeInfo)).Decode(&nonP4)
		//if err != nil {
		//	log.Fatal("decode error 1:", err)
		//}
		//fmt.Println("Got P4 first table name: ", p4.Tables[0].Name)
		//fmt.Println("Got non-P4 first table name: ", nonP4.Tables[0].Name)
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
