package viewPage

import (
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	overlayBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(0, 1)

	overlayTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")).
		Bold(true)

	overlayHintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	copiedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Bold(true)
)

type copiedExpiredMsg struct{}

func (m *Model) initOverlay(content string) {
	m.overlayActive = true
	m.overlayContent = content
	m.overlayLines = strings.Split(content, "\n")
	m.overlayScroll = 0
	m.overlayCopied = false
}

func (m *Model) dismissOverlay() {
	m.overlayActive = false
	m.overlayContent = ""
	m.overlayLines = nil
	m.overlayScroll = 0
	m.overlayCopied = false
}

func (m *Model) handleOverlayKey(key string) tea.Cmd {
	overlayH := m.overlayHeight()
	maxScroll := len(m.overlayLines) - overlayH
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch key {
	case "q", "esc":
		m.dismissOverlay()
	case "j", "down":
		m.overlayScroll++
		if m.overlayScroll > maxScroll {
			m.overlayScroll = maxScroll
		}
	case "k", "up":
		m.overlayScroll--
		if m.overlayScroll < 0 {
			m.overlayScroll = 0
		}
	case "d":
		m.overlayScroll += overlayH / 2
		if m.overlayScroll > maxScroll {
			m.overlayScroll = maxScroll
		}
	case "u":
		m.overlayScroll -= overlayH / 2
		if m.overlayScroll < 0 {
			m.overlayScroll = 0
		}
	case "G":
		m.overlayScroll = maxScroll
	case "g":
		m.overlayScroll = 0
	case "y":
		copyToClipboard(m.overlayContent)
		m.overlayCopied = true
		return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return copiedExpiredMsg{}
		})
	}
	return nil
}

func (m *Model) overlayHeight() int {
	h := m.height * 80 / 100
	if h < 3 {
		h = 3
	}
	return h
}

func (m *Model) overlayWidth() int {
	w := m.width * 80 / 100
	if w < 20 {
		w = 20
	}
	return w
}

func (m Model) renderOverlay() string {
	w := m.overlayWidth()
	h := m.overlayHeight()

	// Title bar
	title := overlayTitleStyle.Render("Schema")
	hint := " j/k:scroll  y:copy  q:close"
	if m.overlayCopied {
		hint = " " + copiedStyle.Render("Copied!")
	}
	hintStr := overlayHintStyle.Render(hint)
	titleLine := title + hintStr

	// Content area (h - 1 for title)
	contentH := h - 1
	if contentH < 1 {
		contentH = 1
	}

	end := m.overlayScroll + contentH
	if end > len(m.overlayLines) {
		end = len(m.overlayLines)
	}
	start := m.overlayScroll
	if start > len(m.overlayLines) {
		start = len(m.overlayLines)
	}

	visibleLines := m.overlayLines[start:end]
	// Pad to fill height
	for len(visibleLines) < contentH {
		visibleLines = append(visibleLines, "")
	}

	// Truncate lines to fit width
	innerW := w - 4 // account for border + padding
	if innerW < 1 {
		innerW = 1
	}
	for i, line := range visibleLines {
		runes := []rune(line)
		if len(runes) > innerW {
			visibleLines[i] = string(runes[:innerW])
		}
	}

	body := titleLine + "\n" + strings.Join(visibleLines, "\n")

	box := overlayBorderStyle.
		Width(w).
		Render(body)

	return lipgloss.Place(
		m.width, m.height+2, // +2 for header/footer
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		// Linux: try xclip, then xsel
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}
