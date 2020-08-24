package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	logo     bool
	version  bool
	piscLogo string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the pisc-cli version information",
	Long:  `Show the pisc-cli version information`,
	Run: func(cmd *cobra.Command, args []string) {
		if logo {
			callLogo()
			fmt.Println(piscLogo)
		}
		fmt.Println("Version Called")
	},
}

func callLogo() {
	piscLogo =
		"           /$$\n" +
			"           |__/\n" +
			"  /$$$$$$  /$$  /$$$$$$$  /$$$$$$$\n" +
			" /$$__  $$| $$ /$$_____/ /$$_____/\n" +
			"| $$  \\ $$| $$|  $$$$$$ | $$\n" +
			"| $$  | $$| $$ \\____  $$| $$\n" +
			"| $$$$$$$/| $$ /$$$$$$$/|  $$$$$$$\n" +
			"| $$____/ |__/|_______/  \\_______/\n" +
			"| $$\n" +
			"| $$\n" +
			"|__/\n"
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&logo, "logo", "l", false, "")
}
