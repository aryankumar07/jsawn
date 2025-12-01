/*
Copyright © 2025 Aryan Kumar aryan.pageme@gmail.com
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aryankumar07/jsawn/jsonFormatter"
	"github.com/aryankumar07/jsawn/viewPage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsawn",
	Short: "",
	Long:  ` `,
	Args:  cobra.ArbitraryArgs,
	Run:   runJsonFormatter,
}

func runJsonFormatter(cmd *cobra.Command, args []string) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("failed to check stdin: %v", err)
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		jsonData, err := jsonFormatter.GetJsonDataFromPipe()
		if err != nil {
			log.Fatalln(err)
		}
		m := viewPage.InitModel(*jsonData)
		p := tea.NewProgram(
			m,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			fmt.Println("could not run program:", err)
			os.Exit(1)
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// write the flags for CLI here
}
