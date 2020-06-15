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

// tableCmd represents the table command
var tableCmd = &cobra.Command{
	Use:   "table",
	Args:  cobra.ExactArgs(0),
	Short: "List all tables",
	Long:  `List all tables in BFRTInfo which include P4 and Non-P4 tables`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _, cancel, p4, nonP4 := initConfigClient()
		defer cancel()

		fmt.Println("------ The following is for P4 table ------")
		for _, v := range p4.Tables {
			fmt.Println(v.Name)
		}

		fmt.Println("------ The following is for non-P4 table ------")
		for _, v := range nonP4.Tables {
			fmt.Println(v.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tableCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
