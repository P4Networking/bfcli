package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
		_, _, conn, cancel, p4Info, nonP4Info := initConfigClient()
		defer conn.Close()
		defer cancel()

		//fmt.Println("------ The following is for P4 table ------")
		for _, v := range p4Info.Tables {
			fmt.Println(v.Name)
		}

		if all {
			fmt.Println("------ The following is for non-P4 table ------")
			for _, v := range nonP4Info.Tables {
				fmt.Println(v.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)
	tableCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all tables")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tableCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
