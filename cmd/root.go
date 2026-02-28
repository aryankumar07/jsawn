/*
Copyright © 2025 Aryan Kumar aryan.pageme@gmail.com
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aryankumar07/jsawn/source"
	"github.com/aryankumar07/jsawn/tree"
	"github.com/aryankumar07/jsawn/viewPage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	magenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	bold    = lipgloss.NewStyle().Bold(true)
	dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func helpText() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Render("jsawn")
	b.WriteString(fmt.Sprintf("\n  %s — interactive JSON viewer for the terminal\n\n", title))

	// Usage
	b.WriteString(bold.Render("  USAGE") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	b.WriteString(fmt.Sprintf("    %s                          %s\n", cyan.Render("jsawn data.json"), dim.Render("open a local file")))
	b.WriteString(fmt.Sprintf("    %s                %s\n", cyan.Render("jsawn <url>"), dim.Render("fetch JSON from a URL")))
	b.WriteString(fmt.Sprintf("    %s   %s\n", cyan.Render("jsawn a.json,b.json,<url>"), dim.Render("multiple sources as tabs")))
	b.WriteString(fmt.Sprintf("    %s                             %s\n", cyan.Render("cat data.json | jsawn"), dim.Render("pipe from stdin")))
	b.WriteString("\n")

	// Flags
	b.WriteString(bold.Render("  HTTP OPTIONS") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	b.WriteString(fmt.Sprintf("    %s       %s\n", cyan.Render("-H 'Key: Value'"), dim.Render("add HTTP header (repeatable)")))
	b.WriteString(fmt.Sprintf("    %s           %s\n", cyan.Render("-X METHOD"), dim.Render("HTTP method (default: GET)")))
	b.WriteString(fmt.Sprintf("    %s         %s\n", cyan.Render("-d '{...}'"), dim.Render("request body (auto-sets Content-Type: application/json)")))
	b.WriteString("\n")

	// Examples
	b.WriteString(bold.Render("  EXAMPLES") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	b.WriteString(fmt.Sprintf("    %s\n", dim.Render("explore a local file")))
	b.WriteString(fmt.Sprintf("    %s\n\n", cyan.Render("jsawn package.json")))
	b.WriteString(fmt.Sprintf("    %s\n", dim.Render("pipe from any command")))
	b.WriteString(fmt.Sprintf("    %s\n\n", cyan.Render("kubectl get pods -o json | jsawn")))
	b.WriteString(fmt.Sprintf("    %s\n", dim.Render("fetch an API with auth")))
	b.WriteString(fmt.Sprintf("    %s\n\n", cyan.Render("jsawn -H 'Authorization: Bearer token' https://api.github.com/user")))
	b.WriteString(fmt.Sprintf("    %s\n", dim.Render("POST and inspect the response")))
	b.WriteString(fmt.Sprintf("    %s\n\n", cyan.Render("jsawn -X POST -d '{\"query\":\"users\"}' https://api.example.com/graphql")))
	b.WriteString(fmt.Sprintf("    %s\n", dim.Render("compare local and remote JSON")))
	b.WriteString(fmt.Sprintf("    %s\n", cyan.Render("jsawn config.json,https://api.example.com/config")))
	b.WriteString("\n")

	// Navigation
	b.WriteString(bold.Render("  NAVIGATION") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	writeKey(&b, "j / ↓", "move down")
	writeKey(&b, "k / ↑", "move up")
	writeKey(&b, "h / ←", "collapse node / go to parent")
	writeKey(&b, "l / →", "expand node / enter first child")
	writeKey(&b, "space", "toggle collapse/expand")
	writeKey(&b, "gg", "jump to top")
	writeKey(&b, "G", "jump to bottom")
	b.WriteString("\n")

	// Folding
	b.WriteString(bold.Render("  FOLDING") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	writeKey(&b, "e", "expand all")
	writeKey(&b, "E", "collapse all")
	writeKey(&b, "1-9", "fold to depth N")
	b.WriteString("\n")

	// Search
	b.WriteString(bold.Render("  SEARCH") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	writeKey(&b, "/", "start search")
	writeKey(&b, "enter", "execute search")
	writeKey(&b, "n / N", "next / previous match")
	writeKey(&b, "esc", "clear search")
	writeKey(&b, "ctrl+u", "clear search query")
	b.WriteString("\n")

	// Views & Commands
	b.WriteString(bold.Render("  VIEWS & COMMANDS") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	writeKey(&b, "f", "toggle flat view (leaf paths)")
	writeKey(&b, ":", "enter command mode")
	writeKey(&b, "tab / shift+tab", "switch tabs (multi-source)")
	b.WriteString("\n")

	// Schema commands
	b.WriteString(bold.Render("  COMMANDS") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	writeKey(&b, ":schema go", "generate Go struct")
	writeKey(&b, ":schema ts", "generate TypeScript interface")
	writeKey(&b, ":schema zod", "generate Zod schema")
	b.WriteString("\n")

	// Overlay
	b.WriteString(bold.Render("  SCHEMA OVERLAY") + "\n")
	b.WriteString(dim.Render("  ─────────────────────────────────────────────") + "\n")
	writeKey(&b, "j / k", "scroll down / up")
	writeKey(&b, "d / u", "page down / up")
	writeKey(&b, "y", "copy to clipboard")
	writeKey(&b, "q / esc", "close overlay")
	b.WriteString("\n")

	// Quit
	writeKey(&b, "q / esc / ctrl+c", "quit")
	b.WriteString("\n")

	return b.String()
}

func writeKey(b *strings.Builder, key, desc string) {
	b.WriteString(fmt.Sprintf("    %s %s\n",
		green.Render(fmt.Sprintf("%-20s", key)),
		dim.Render(desc),
	))
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf("jsawn %s (commit: %s, built: %s)\n", version, commit, date))
}

var rootCmd = &cobra.Command{
	Use:   "jsawn [sources]",
	Short: "Interactive JSON viewer for the terminal",
	Long:  helpText(),
	Args:  cobra.MaximumNArgs(1),
	Run:   runJsonViewer,
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

	headers, _ := cmd.Flags().GetStringArray("header")
	method, _ := cmd.Flags().GetString("method")
	data, _ := cmd.Flags().GetString("data")

	opts := source.RequestOptions{
		Method:  method,
		Headers: headers,
		Body:    data,
	}

	sources := source.ResolveAll(arg, stdinData, opts)
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
	rootCmd.Flags().StringArrayP("header", "H", nil, "HTTP header in 'Key: Value' format (repeatable)")
	rootCmd.Flags().StringP("method", "X", "GET", "HTTP method")
	rootCmd.Flags().StringP("data", "d", "", "Request body")

	rootCmd.SetHelpTemplate(`{{.Long}}{{if .HasAvailableFlags}}
  FLAGS
  ─────────────────────────────────────────────
{{.Flags.FlagUsages}}{{end}}
`)
}
