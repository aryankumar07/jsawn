/*
Copyright © 2025 Aryan Kumar aryan.pageme@gmail.com
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aryankumar07/jsawn/tree"
	"github.com/aryankumar07/jsawn/viewPage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsawn [file]",
	Short: "Interactive JSON viewer for the terminal",
	Long: `jsawn is a fast, interactive JSON viewer for the terminal.

Pipe JSON from stdin or pass a file path as an argument:
  cat data.json | jsawn
  curl https://api.example.com | jsawn
  jsawn data.json

Navigate with vim-style keybindings, fold/expand nodes,
and explore large JSON documents with ease.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runJsonViewer,
}

func runJsonViewer(cmd *cobra.Command, args []string) {
	var data []byte
	var err error

	stat, statErr := os.Stdin.Stat()
	piped := statErr == nil && (stat.Mode()&os.ModeCharDevice) == 0

	if piped {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read stdin: %v", err)
		}
	} else if len(args) == 1 {
		data, err = os.ReadFile(args[0])
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
	} else {
		cmd.Help()
		return
	}

	if len(data) == 0 {
		log.Fatalln("JSON returned null")
	}

	root, err := tree.Parse(data)
	if err != nil {
		log.Fatalf("failed to parse JSON: %v", err)
	}

	m := viewPage.InitModel(root)
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// write the flags for CLI here
}
