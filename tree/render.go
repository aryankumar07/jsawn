package tree

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	keyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))  // cyan
	stringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	numberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	boolStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // magenta
	nullStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	bracketStyle = lipgloss.NewStyle()
	cursorStyle  = lipgloss.NewStyle().Reverse(true)
	matchStyle   = lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")) // yellow bg, black fg
)

const indent = "  "

func RenderEntry(e VisibleEntry, cursorOn bool, width int, searchQuery string) string {
	n := e.Node
	depth := n.Depth
	prefix := strings.Repeat(indent, depth)

	var line string

	if e.IsClosing {
		bracket := "}"
		if n.Type == TypeArray {
			bracket = "]"
		}
		line = prefix + bracketStyle.Render(bracket)
	} else if n.IsContainer() {
		line = renderContainer(n, prefix, searchQuery)
	} else {
		line = renderPrimitive(n, prefix, searchQuery)
	}

	if cursorOn {
		plain := lipgloss.NewStyle().Width(width).Render(line)
		return cursorStyle.Render(plain)
	}
	return line
}

func renderContainer(n *Node, prefix string, searchQuery string) string {
	var keyPart string
	if n.Parent != nil && n.Parent.Type == TypeObject {
		keyPart = highlightText(fmt.Sprintf("%q", n.Key), searchQuery, keyStyle) + ": "
	} else if n.Parent != nil && n.Parent.Type == TypeArray {
		keyPart = highlightText(n.Key, searchQuery, keyStyle) + ": "
	}

	openBracket := "{"
	closeBracket := "}"
	if n.Type == TypeArray {
		openBracket = "["
		closeBracket = "]"
	}

	if n.Collapsed {
		count := len(n.Children)
		label := "keys"
		if n.Type == TypeArray {
			label = "items"
		}
		if count == 1 {
			if n.Type == TypeArray {
				label = "item"
			} else {
				label = "key"
			}
		}
		summary := fmt.Sprintf("%s... %d %s%s", openBracket, count, label, closeBracket)
		return prefix + keyPart + bracketStyle.Render(summary)
	}

	if len(n.Children) == 0 {
		return prefix + keyPart + bracketStyle.Render(openBracket+closeBracket)
	}

	return prefix + keyPart + bracketStyle.Render(openBracket)
}

func renderPrimitive(n *Node, prefix string, searchQuery string) string {
	var keyPart string
	if n.Parent != nil && n.Parent.Type == TypeObject {
		keyPart = highlightText(fmt.Sprintf("%q", n.Key), searchQuery, keyStyle) + ": "
	} else if n.Parent != nil && n.Parent.Type == TypeArray {
		keyPart = highlightText(n.Key, searchQuery, keyStyle) + ": "
	}

	var valStr string
	switch n.Type {
	case TypeString:
		valStr = highlightText(fmt.Sprintf("%q", n.Value.(string)), searchQuery, stringStyle)
	case TypeNumber:
		valStr = highlightText(formatNumber(n.Value.(float64)), searchQuery, numberStyle)
	case TypeBool:
		valStr = highlightText(fmt.Sprintf("%t", n.Value.(bool)), searchQuery, boolStyle)
	case TypeNull:
		valStr = highlightText("null", searchQuery, nullStyle)
	}

	return prefix + keyPart + valStr
}

func highlightText(text, query string, baseStyle lipgloss.Style) string {
	if query == "" {
		return baseStyle.Render(text)
	}

	lower := strings.ToLower(text)
	lowerQ := strings.ToLower(query)

	if !strings.Contains(lower, lowerQ) {
		return baseStyle.Render(text)
	}

	var b strings.Builder
	i := 0
	for i < len(text) {
		idx := strings.Index(lower[i:], lowerQ)
		if idx == -1 {
			b.WriteString(baseStyle.Render(text[i:]))
			break
		}
		if idx > 0 {
			b.WriteString(baseStyle.Render(text[i : i+idx]))
		}
		matchEnd := i + idx + len(lowerQ)
		if matchEnd > len(text) {
			matchEnd = len(text)
		}
		b.WriteString(matchStyle.Render(text[i+idx : matchEnd]))
		i = matchEnd
	}
	return b.String()
}

func formatNumber(f float64) string {
	if f == math.Trunc(f) && !math.IsInf(f, 0) {
		return fmt.Sprintf("%g", f)
	}
	return fmt.Sprintf("%g", f)
}
