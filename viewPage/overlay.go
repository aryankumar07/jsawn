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

type copiedExpiredMsg struct {
	tabIndex int
}

func (m *Model) initOverlay(content string) {
	t := m.active()
	t.overlayActive = true
	t.overlayContent = content
	t.overlayLines = strings.Split(content, "\n")
	t.overlayScroll = 0
	t.overlayCopied = false
}

func (m *Model) dismissOverlay() {
	t := m.active()
	t.overlayActive = false
	t.overlayContent = ""
	t.overlayLines = nil
	t.overlayScroll = 0
	t.overlayCopied = false
}

func (m *Model) handleOverlayKey(key string) tea.Cmd {
	t := m.active()
	overlayH := m.overlayHeight()
	maxScroll := len(t.overlayLines) - overlayH
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch key {
	case "q", "esc":
		m.dismissOverlay()
	case "j", "down":
		t.overlayScroll++
		if t.overlayScroll > maxScroll {
			t.overlayScroll = maxScroll
		}
	case "k", "up":
		t.overlayScroll--
		if t.overlayScroll < 0 {
			t.overlayScroll = 0
		}
	case "d":
		t.overlayScroll += overlayH / 2
		if t.overlayScroll > maxScroll {
			t.overlayScroll = maxScroll
		}
	case "u":
		t.overlayScroll -= overlayH / 2
		if t.overlayScroll < 0 {
			t.overlayScroll = 0
		}
	case "G":
		t.overlayScroll = maxScroll
	case "g":
		t.overlayScroll = 0
	case "y":
		copyToClipboard(t.overlayContent)
		t.overlayCopied = true
		tabIdx := m.activeTab
		return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return copiedExpiredMsg{tabIndex: tabIdx}
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
	t := m.tabs[m.activeTab]
	w := m.overlayWidth()
	h := m.overlayHeight()

	// Title bar
	title := overlayTitleStyle.Render("Schema")
	hint := " j/k:scroll  y:copy  q:close"
	if t.overlayCopied {
		hint = " " + copiedStyle.Render("Copied!")
	}
	hintStr := overlayHintStyle.Render(hint)
	titleLine := title + hintStr

	// Content area (h - 1 for title)
	contentH := h - 1
	if contentH < 1 {
		contentH = 1
	}

	end := t.overlayScroll + contentH
	if end > len(t.overlayLines) {
		end = len(t.overlayLines)
	}
	start := t.overlayScroll
	if start > len(t.overlayLines) {
		start = len(t.overlayLines)
	}

	visibleLines := t.overlayLines[start:end]
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
