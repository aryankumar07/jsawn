package viewPage

import (
	"fmt"
	"strings"

	"github.com/aryankumar07/jsawn/tree"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()

	searchStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	noMatchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	breadcrumbStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
)

type Model struct {
	root          *tree.Node
	visible       []tree.VisibleEntry
	cursor        int
	offset        int
	width         int
	height        int
	ready         bool
	gPressed      bool
	searching     bool
	searchQuery   string
	searchMatches []int
	matchIndex    int
	flatMode      bool
	flatEntries   []tree.FlatEntry
}

func InitModel(root *tree.Node) Model {
	visible := tree.VisibleNodes(root)
	return Model{
		root:       root,
		visible:    visible,
		matchIndex: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// listLen returns the length of the current view's data (flat entries or visible tree nodes).
func (m *Model) listLen() int {
	if m.flatMode {
		return len(m.flatEntries)
	}
	return len(m.visible)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Search mode input
		if m.searching {
			switch key {
			case "enter":
				m.searching = false
				m.executeSearch()
			case "esc":
				m.searching = false
				m.searchQuery = ""
				m.searchMatches = nil
				m.matchIndex = -1
			case "ctrl+c":
				return m, tea.Quit
			case "backspace":
				if len(m.searchQuery) > 0 {
					runes := []rune(m.searchQuery)
					m.searchQuery = string(runes[:len(runes)-1])
				}
			case "ctrl+u":
				m.searchQuery = ""
			default:
				if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
					m.searchQuery += key
				}
			}
			return m, nil
		}

		// Normal mode
		switch {
		case key == "ctrl+c" || key == "q":
			return m, tea.Quit

		case key == "esc":
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.searchMatches = nil
				m.matchIndex = -1
			} else {
				return m, tea.Quit
			}

		case key == "/":
			m.searching = true
			m.searchQuery = ""
			m.searchMatches = nil
			m.matchIndex = -1
			m.gPressed = false
			return m, nil

		case key == "n":
			m.nextMatch()

		case key == "N":
			m.prevMatch()

		case key == "f":
			m.flatMode = !m.flatMode
			if m.flatMode {
				m.flatEntries = tree.FlattenLeaves(m.root)
			}
			m.cursor = 0
			m.offset = 0
			m.rebuildSearchMatches()

		case key == "j" || key == "down":
			m.moveCursor(1)

		case key == "k" || key == "up":
			m.moveCursor(-1)

		case key == "l" || key == "right":
			if !m.flatMode {
				m.expandOrEnter()
			}

		case key == "h" || key == "left":
			if !m.flatMode {
				m.collapseOrParent()
			}

		case key == " ":
			if !m.flatMode {
				m.toggleCurrent()
			}

		case key == "e":
			if !m.flatMode {
				m.root.ExpandAll()
				m.rebuildVisible()
			}

		case key == "E":
			if !m.flatMode {
				m.root.CollapseAll()
				m.rebuildVisible()
				m.cursor = 0
			}

		case key >= "1" && key <= "9":
			if !m.flatMode {
				depth := int(key[0] - '0')
				m.root.CollapseToDepth(depth)
				m.rebuildVisible()
				m.clampCursor()
			}

		case key == "G":
			m.gPressed = false
			m.cursor = m.listLen() - 1

		case key == "g":
			if m.gPressed {
				m.cursor = 0
				m.gPressed = false
			} else {
				m.gPressed = true
				return m, nil
			}

		default:
			m.gPressed = false
		}

		if key != "g" {
			m.gPressed = false
		}
		m.clampOffset()

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.moveCursor(-3)
			m.clampOffset()
		case tea.MouseButtonWheelDown:
			m.moveCursor(3)
			m.clampOffset()
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		m.width = msg.Width
		m.height = msg.Height - headerHeight - footerHeight
		m.ready = true
		m.clampOffset()
	}

	return m, nil
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
}

func (m *Model) clampCursor() {
	if m.cursor < 0 {
		m.cursor = 0
	}
	n := m.listLen()
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) clampOffset() {
	if m.height <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
	maxOffset := m.listLen() - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) expandOrEnter() {
	if m.cursor >= len(m.visible) {
		return
	}
	entry := m.visible[m.cursor]
	if entry.IsClosing {
		return
	}
	n := entry.Node
	if !n.IsContainer() {
		return
	}
	if n.Collapsed {
		n.Toggle()
		m.rebuildVisible()
	} else if len(n.Children) > 0 {
		// enter first child
		m.cursor++
		m.clampCursor()
	}
}

func (m *Model) collapseOrParent() {
	if m.cursor >= len(m.visible) {
		return
	}
	entry := m.visible[m.cursor]
	n := entry.Node

	if entry.IsClosing {
		// jump to the opening line of this container
		m.jumpToNode(n)
		return
	}

	if n.IsContainer() && !n.Collapsed {
		n.Toggle()
		m.rebuildVisible()
		return
	}

	// jump to parent
	if n.Parent != nil {
		m.jumpToNode(n.Parent)
	}
}

func (m *Model) jumpToNode(target *tree.Node) {
	for i, e := range m.visible {
		if e.Node == target && !e.IsClosing {
			m.cursor = i
			break
		}
	}
	m.clampOffset()
}

func (m *Model) toggleCurrent() {
	if m.cursor >= len(m.visible) {
		return
	}
	entry := m.visible[m.cursor]
	if entry.IsClosing {
		return
	}
	n := entry.Node
	if n.IsContainer() {
		n.Toggle()
		m.rebuildVisible()
		m.clampCursor()
	}
}

func (m *Model) rebuildVisible() {
	m.visible = tree.VisibleNodes(m.root)
	m.rebuildSearchMatches()
}

func (m *Model) executeSearch() {
	if m.searchQuery == "" {
		m.searchMatches = nil
		m.matchIndex = -1
		return
	}

	if m.flatMode {
		m.searchMatches = nil
		for i, entry := range m.flatEntries {
			if entry.MatchesSearch(m.searchQuery) {
				m.searchMatches = append(m.searchMatches, i)
			}
		}
		if len(m.searchMatches) > 0 {
			m.matchIndex = 0
			m.cursor = m.searchMatches[0]
			m.clampOffset()
		} else {
			m.matchIndex = -1
		}
		return
	}

	// Tree mode search
	matchingNodes := tree.FindMatchingNodes(m.root, m.searchQuery)
	if len(matchingNodes) == 0 {
		m.searchMatches = nil
		m.matchIndex = -1
		return
	}

	// Expand all ancestors so matches become visible
	for _, n := range matchingNodes {
		tree.ExpandToNode(n)
	}

	m.visible = tree.VisibleNodes(m.root)

	// Map matching nodes to their visible indices
	matchSet := make(map[*tree.Node]bool)
	for _, n := range matchingNodes {
		matchSet[n] = true
	}

	m.searchMatches = nil
	for i, entry := range m.visible {
		if !entry.IsClosing && matchSet[entry.Node] {
			m.searchMatches = append(m.searchMatches, i)
		}
	}

	if len(m.searchMatches) > 0 {
		m.matchIndex = 0
		m.cursor = m.searchMatches[0]
		m.clampOffset()
	}
}

func (m *Model) nextMatch() {
	if len(m.searchMatches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex + 1) % len(m.searchMatches)
	m.cursor = m.searchMatches[m.matchIndex]
	m.clampOffset()
}

func (m *Model) prevMatch() {
	if len(m.searchMatches) == 0 {
		return
	}
	m.matchIndex--
	if m.matchIndex < 0 {
		m.matchIndex = len(m.searchMatches) - 1
	}
	m.cursor = m.searchMatches[m.matchIndex]
	m.clampOffset()
}

func (m *Model) rebuildSearchMatches() {
	if m.searchQuery == "" {
		m.searchMatches = nil
		m.matchIndex = -1
		return
	}

	m.searchMatches = nil
	if m.flatMode {
		for i, entry := range m.flatEntries {
			if entry.MatchesSearch(m.searchQuery) {
				m.searchMatches = append(m.searchMatches, i)
			}
		}
	} else {
		for i, entry := range m.visible {
			if !entry.IsClosing && entry.Node.MatchesSearch(m.searchQuery) {
				m.searchMatches = append(m.searchMatches, i)
			}
		}
	}

	if len(m.searchMatches) == 0 {
		m.matchIndex = -1
	} else if m.matchIndex >= len(m.searchMatches) {
		m.matchIndex = len(m.searchMatches) - 1
	}
}

func (m Model) headerView() string {
	titleText := "Mr. Jsawn"
	if m.flatMode {
		titleText = "Mr. Jsawn (flat)"
	}
	title := titleStyle.Render(titleText)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	pct := 0.0
	n := m.listLen()
	if n > 1 {
		pct = float64(m.cursor) / float64(n-1) * 100
	}
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", pct))
	infoWidth := lipgloss.Width(info)

	// When in search mode, show search input instead of breadcrumb
	if m.searching {
		searchInfo := searchStyle.Render(fmt.Sprintf("/%s█", m.searchQuery))
		searchWidth := lipgloss.Width(searchInfo)
		lineWidth := max(0, m.width-searchWidth-infoWidth)
		line := strings.Repeat("─", lineWidth)
		return lipgloss.JoinHorizontal(lipgloss.Center, searchInfo, line, info)
	}

	// Breadcrumb path
	path := "."
	if m.flatMode {
		if m.cursor < len(m.flatEntries) {
			path = m.flatEntries[m.cursor].Path
		}
	} else {
		if m.cursor < len(m.visible) {
			path = tree.NodePath(m.visible[m.cursor].Node)
		}
	}

	// Search status (appended after breadcrumb)
	var searchInfo string
	if m.searchQuery != "" && len(m.searchMatches) == 0 {
		searchInfo = noMatchStyle.Render(" [no matches]")
	} else if len(m.searchMatches) > 0 {
		searchInfo = searchStyle.Render(fmt.Sprintf(" [%d/%d]", m.matchIndex+1, len(m.searchMatches)))
	}

	searchInfoWidth := lipgloss.Width(searchInfo)
	availableForPath := m.width - infoWidth - searchInfoWidth - 1
	if availableForPath < 1 {
		availableForPath = 1
	}

	// Horizontal scroll: crop from the left if path is too long
	pathRunes := []rune(path)
	if len(pathRunes) > availableForPath {
		path = string(pathRunes[len(pathRunes)-availableForPath:])
	}

	pathRendered := breadcrumbStyle.Render(path)
	leftContent := pathRendered + searchInfo
	leftWidth := lipgloss.Width(leftContent)
	lineWidth := max(0, m.width-leftWidth-infoWidth)
	line := strings.Repeat("─", lineWidth)

	return lipgloss.JoinHorizontal(lipgloss.Center, leftContent, line, info)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var lines []string
	total := m.listLen()
	end := m.offset + m.height
	if end > total {
		end = total
	}

	for i := m.offset; i < end; i++ {
		isCursor := i == m.cursor
		var line string
		if m.flatMode {
			line = tree.RenderFlatEntry(m.flatEntries[i], isCursor, m.width, m.searchQuery)
		} else {
			line = tree.RenderEntry(m.visible[i], isCursor, m.width, m.searchQuery)
		}
		lines = append(lines, line)
	}

	// pad with empty lines if content is shorter than viewport
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), content, m.footerView())
}
