/*
Copyright © 2025 Aryan Kumar aryan.pageme@gmail.com
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aryankumar07/jsawn/source"
	"github.com/aryankumar07/jsawn/tree"
	"github.com/aryankumar07/jsawn/viewPage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsawn [sources]",
	Short: "Interactive JSON viewer for the terminal",
	Long: `jsawn is a fast, interactive JSON viewer for the terminal.

Pipe JSON from stdin or pass a file path as an argument:
  cat data.json | jsawn
  curl https://api.example.com | jsawn
  jsawn data.json

Fetch JSON from a URL:
  jsawn https://api.example.com/data

View multiple sources (comma-separated):
  jsawn file.json,https://api.com/data,other.json

Navigate with vim-style keybindings, fold/expand nodes,
and explore large JSON documents with ease.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runJsonViewer,
}

func runJsonViewer(cmd *cobra.Command, args []string) {
	var stdinData []byte

	stat, statErr := os.Stdin.Stat()
	piped := statErr == nil && (stat.Mode()&os.ModeCharDevice) == 0

	if piped {
		var err error
		stdinData, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read stdin: %v", err)
		}
	}

	var arg string
	if len(args) == 1 {
		arg = args[0]
	}

	sources := source.ResolveAll(arg, stdinData)
	if len(sources) == 0 {
		cmd.Help()
		return
	}

	tabs := make([]viewPage.TabData, len(sources))
	for i, src := range sources {
		td := viewPage.TabData{Label: src.Label}
		if src.Err != nil {
			td.Error = src.Err.Error()
		} else if len(src.Data) == 0 {
			td.Error = "empty response"
		} else {
			root, err := tree.Parse(src.Data)
			if err != nil {
				td.Error = fmt.Sprintf("invalid JSON: %v", err)
			} else {
				td.Root = root
			}
		}
		tabs[i] = td
	}

	m := viewPage.InitModel(tabs)
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
