package cmd

import (
	"fmt"

	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
	"github.com/GeraAnggaraPutra/go-backup/internal/dashboard"
	"github.com/spf13/cobra"
)

var dashPort int

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the static web dashboard to view and download backups",
	Long:  `This command starts a web server that serves the dashboard and allows you to browse the backups folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s[System] Preparing Dashboard Service...%s\n", constant.ColorBlue, constant.ColorReset)

		err := dashboard.StartServer(dashPort)
		if err != nil {
			fmt.Printf("%s[Error] Failed to start server: %v%s\n", constant.ColorRed, err, constant.ColorReset)
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.Flags().IntVarP(&dashPort, "port", "p", 8080, "Port for the dashboard server")
}
