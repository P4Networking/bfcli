package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	all bool
)

// tableCmd represents the table command
var tableCmd = &cobra.Command{
	Use:   "table",
	Args:  cobra.ExactArgs(0),
	Short: "List all tables",
	Long:  `List all tables in BFRTInfo which include P4 and Non-P4 tables`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _, conn, cancel, _, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		fmt.Println("------ The P4 table ------")
		for _, v := range Obj.p4Info.Tables {
			if NotSupportToReadTable[v.ID] {
				continue
			}
			if strings.Contains(v.Name, preFixIg) || strings.Contains(v.Name, preFixEg) {
				fmt.Println(v.Name)
			}
		}

		if all {
			fmt.Println("------ The  non-P4 table ------")
			for _, v := range Obj.nonP4Info.Tables {
				fmt.Println(v.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)
	tableCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all tables")
}
