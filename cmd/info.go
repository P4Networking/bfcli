package cmd

import (
	"fmt"
	"github.com/P4Networking/pisc/util"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info TABLE-NAME",
	Short: "Show information about table",
	Long:  `Display the detail of table.`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()
		argsList, _ := p4Info.GuessTableName(toComplete)
		return argsList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {

		_, _, conn, cancel, p4Info, _ := initConfigClient()
		defer conn.Close()
		defer cancel()

		/* non-p4Info not used yet. If you need to see non-p4 object,
		please open the issue to talk about this.
		tableList, ok := nonP4Info.GuessTableName(args[0])
		if !ok {
			fmt.Printf("Can not found the table %s\n", args[0])
			return
		}
		*/

		tableId := p4Info.SearchTableId(args[0])
		if uint32(tableId) == util.ID_NOT_FOUND {
			fmt.Printf("Can not found table with name: %s\n", args[0])
			return
		}

		table := p4Info.SearchTableById(tableId)
		if table == nil {
			fmt.Printf("Can not found table with Id %d\n", tableId)
			return
		}

		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("Table Info")
		if table.Name != "" {
			fmt.Printf("  %-12s: %-6s\n", "Name", table.Name)
		}
		if table.ID != 0 {
			fmt.Printf("  %-12s: %-6d\n", "ID", table.ID)
		}
		if table.TableType != "" {
			fmt.Printf("  %-12s: %-6s\n", "Type", table.TableType)
		}
		if table.Size != 0 {
			fmt.Printf("  %-12s: %-6d\n", "Size", table.Size)
		}
		if table.Annotations != nil {
			fmt.Printf("  %-12s:\n", "Annotations")
			for k, v := range table.Annotations {
				fmt.Printf("%d - Name: %s | Value: %s \n", k+1, v.Name, v.Value)
			}
		}
		if table.DependsOn != nil {
			fmt.Printf("%-12s: %-6s\n", "Table DependsOn", table.DependsOn)
		}

		if table.Key != nil {
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Printf("%-s\n", "Match Key Info")
			fmt.Printf("  %-8s %-20s %-11s %-10s %-9s %-8s %-4s",
				"KeyId", "Name", "Match_type", "Mandatory", "Repeated", "Type", "Width\n")
			for _, v := range table.Key {
				if v.ID == 65537 || v.Name == "$MATCH_PRIORITY" {
					continue
				}
				fmt.Printf("  %-8d %-20s %-11s %-10t %-9t %-8s %-4d\n",
					v.ID, v.Name, v.MatchType, v.Mandatory, v.Repeated, v.Type.Type, v.Type.Width)
			}
		}

		if table.Data != nil {
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Printf("%-s\n", "Table Data Info")
			fmt.Printf("  %-8s %-20s %-10s %-9s %-8s\n",
				"KeyId", "Name", "Mandatory", "Repeated", "Type")
			for _, v := range table.Data {
				fmt.Printf("  %-8d %-20s %-10t %-9t %-8s\n",
					v.Singleton.ID, v.Singleton.Name, v.Mandatory, v.Singleton.Repeated, v.Singleton.Type.Type)
			}
		}

		if table.ActionSpecs != nil {
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Printf("%-s\n", "Action Info")
			for _, v := range table.ActionSpecs {
				fmt.Printf("  ID: %-6d, Name: %-20s \n", v.ID, v.Name)
				if v.Data != nil {
					fmt.Println("    ----------------------------------------------------------------")
					for _, d := range v.Data {
						fmt.Printf("    %-6s %-20s %-10s %-9s %-8s %-4s\n",
							"ID", "Name", "Mandatory", "Repeated", "Type", "Width")
						fmt.Printf("    %-6d %-20s %-10t %-9t %-8s %-4d\n",
							d.ID, d.Name, d.Mandatory, d.Repeated, d.Type.Type, d.Type.Width)
					}
					fmt.Println("    ----------------------------------------------------------------")
				}
			}
			fmt.Println("--------------------------------------------------------------------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
