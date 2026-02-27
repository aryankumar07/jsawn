package viewPage

import (
	"fmt"
	"strings"

	"github.com/aryankumar07/jsawn/command"
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

	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1)
	errorTabStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

// TabData is the input type passed from cmd/root.go to build tabs.
type TabData struct {
	Label string
	Root  *tree.Node
	Error string
}

// TabState holds all per-document state for one tab.
type TabState struct {
	label   string
	root    *tree.Node
	errMsg  string
	visible []tree.VisibleEntry
	cursor  int
	offset  int

	gPressed      bool
	searching     bool
	searchQuery   string
	searchMatches []int
	matchIndex    int

	flatMode    bool
	flatEntries []tree.FlatEntry

	commandMode  bool
	commandInput string

	overlayActive  bool
	overlayContent string
	overlayLines   []string
	overlayScroll  int
	overlayCopied  bool
}

type Model struct {
	tabs      []TabState
	activeTab int
	width     int
	height    int
	ready     bool
}

func InitModel(tabsData []TabData) Model {
	tabs := make([]TabState, len(tabsData))
	for i, td := range tabsData {
		t := TabState{
			label:      td.Label,
			matchIndex: -1,
		}
		if td.Error != "" {
			t.errMsg = td.Error
		} else {
			t.root = td.Root
			t.visible = tree.VisibleNodes(td.Root)
		}
		tabs[i] = t
	}
	return Model{
		tabs: tabs,
	}
}

func (m *Model) active() *TabState {
	return &m.tabs[m.activeTab]
}

func (m Model) Init() tea.Cmd {
	return nil
}

// listLen returns the length of the current tab's data.
func (t *TabState) listLen() int {
	if t.flatMode {
		return len(t.flatEntries)
	}
	return len(t.visible)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle copied flash expiry
	if ce, ok := msg.(copiedExpiredMsg); ok {
		if ce.tabIndex >= 0 && ce.tabIndex < len(m.tabs) {
			m.tabs[ce.tabIndex].overlayCopied = false
		}
		return m, nil
	}

	t := m.active()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Overlay active — route all keys to overlay handler
		if t.overlayActive {
			if key == "ctrl+c" {
				return m, tea.Quit
			}
			cmd := m.handleOverlayKey(key)
			return m, cmd
		}

		// Tab switching is always available (except in text input modes)
		if !t.searching && !t.commandMode && len(m.tabs) > 1 {
			switch key {
			case "tab":
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, nil
			case "shift+tab":
				m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
				return m, nil
			}
		}

		// If active tab has an error, only allow quit/esc/tab switching
		if t.errMsg != "" {
			switch key {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				return m, tea.Quit
			}
			return m, nil
		}

		// Command mode input
		if t.commandMode {
			switch key {
			case "enter":
				t.commandMode = false
				input := t.commandInput
				t.commandInput = ""
				result := command.Execute(input, t.root)
				if result.Error != "" {
					m.initOverlay("Error: " + result.Error)
				} else if result.Content != "" {
					m.initOverlay(result.Content)
				}
				return m, nil
			case "esc":
				t.commandMode = false
				t.commandInput = ""
			case "ctrl+c":
				return m, tea.Quit
			case "backspace":
				if len(t.commandInput) > 0 {
					runes := []rune(t.commandInput)
					t.commandInput = string(runes[:len(runes)-1])
				}
			case "ctrl+u":
				t.commandInput = ""
			default:
				if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
					t.commandInput += key
				}
			}
			return m, nil
		}

		// Search mode input
		if t.searching {
			switch key {
			case "enter":
				t.searching = false
				m.executeSearch()
			case "esc":
				t.searching = false
				t.searchQuery = ""
				t.searchMatches = nil
				t.matchIndex = -1
			case "ctrl+c":
				return m, tea.Quit
			case "backspace":
				if len(t.searchQuery) > 0 {
					runes := []rune(t.searchQuery)
					t.searchQuery = string(runes[:len(runes)-1])
				}
			case "ctrl+u":
				t.searchQuery = ""
			default:
				if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
					t.searchQuery += key
				}
			}
			return m, nil
		}

		// Normal mode
		switch {
		case key == "ctrl+c" || key == "q":
			return m, tea.Quit

		case key == "esc":
			if t.searchQuery != "" {
				t.searchQuery = ""
				t.searchMatches = nil
				t.matchIndex = -1
			} else {
				return m, tea.Quit
			}

		case key == "/":
			t.searching = true
			t.searchQuery = ""
			t.searchMatches = nil
			t.matchIndex = -1
			t.gPressed = false
			return m, nil

		case key == ":":
			t.commandMode = true
			t.commandInput = ""
			t.gPressed = false
			return m, nil

		case key == "n":
			m.nextMatch()

		case key == "N":
			m.prevMatch()

		case key == "f":
			t.flatMode = !t.flatMode
			if t.flatMode {
				t.flatEntries = tree.FlattenLeaves(t.root)
			}
			t.cursor = 0
			t.offset = 0
			m.rebuildSearchMatches()

		case key == "j" || key == "down":
			m.moveCursor(1)

		case key == "k" || key == "up":
			m.moveCursor(-1)

		case key == "l" || key == "right":
			if !t.flatMode {
				m.expandOrEnter()
			}

		case key == "h" || key == "left":
			if !t.flatMode {
				m.collapseOrParent()
			}

		case key == " ":
			if !t.flatMode {
				m.toggleCurrent()
			}

		case key == "e":
			if !t.flatMode {
				t.root.ExpandAll()
				m.rebuildVisible()
			}

		case key == "E":
			if !t.flatMode {
				t.root.CollapseAll()
				m.rebuildVisible()
				t.cursor = 0
			}

		case key >= "1" && key <= "9":
			if !t.flatMode {
				depth := int(key[0] - '0')
				t.root.CollapseToDepth(depth)
				m.rebuildVisible()
				m.clampCursor()
			}

		case key == "G":
			t.gPressed = false
			t.cursor = t.listLen() - 1

		case key == "g":
			if t.gPressed {
				t.cursor = 0
				t.gPressed = false
			} else {
				t.gPressed = true
				return m, nil
			}

		default:
			t.gPressed = false
		}

		if key != "g" {
			t.gPressed = false
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
		m.width = msg.Width
		m.height = msg.Height - m.chromeHeight()
		m.ready = true
		m.clampOffset()
	}

	return m, nil
}

// chromeHeight returns the total height of non-content chrome (header, footer, tab bar).
func (m *Model) chromeHeight() int {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	tabBarHeight := 0
	if len(m.tabs) > 1 {
		tabBarHeight = 1
	}
	return headerHeight + footerHeight + tabBarHeight
}

func (m *Model) moveCursor(delta int) {
	t := m.active()
	t.cursor += delta
	m.clampCursor()
}

func (m *Model) clampCursor() {
	t := m.active()
	if t.cursor < 0 {
		t.cursor = 0
	}
	n := t.listLen()
	if t.cursor >= n {
		t.cursor = n - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
}

func (m *Model) clampOffset() {
	t := m.active()
	if m.height <= 0 {
		return
	}
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+m.height {
		t.offset = t.cursor - m.height + 1
	}
	maxOffset := t.listLen() - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.offset > maxOffset {
		t.offset = maxOffset
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

func (m *Model) expandOrEnter() {
	t := m.active()
	if t.cursor >= len(t.visible) {
		return
	}
	entry := t.visible[t.cursor]
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
		t.cursor++
		m.clampCursor()
	}
}

func (m *Model) collapseOrParent() {
	t := m.active()
	if t.cursor >= len(t.visible) {
		return
	}
	entry := t.visible[t.cursor]
	n := entry.Node

	if entry.IsClosing {
		m.jumpToNode(n)
		return
	}

	if n.IsContainer() && !n.Collapsed {
		n.Toggle()
		m.rebuildVisible()
		return
	}

	if n.Parent != nil {
		m.jumpToNode(n.Parent)
	}
}

func (m *Model) jumpToNode(target *tree.Node) {
	t := m.active()
	for i, e := range t.visible {
		if e.Node == target && !e.IsClosing {
			t.cursor = i
			break
		}
	}
	m.clampOffset()
}

func (m *Model) toggleCurrent() {
	t := m.active()
	if t.cursor >= len(t.visible) {
		return
	}
	entry := t.visible[t.cursor]
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
	t := m.active()
	t.visible = tree.VisibleNodes(t.root)
	m.rebuildSearchMatches()
}

func (m *Model) executeSearch() {
	t := m.active()
	if t.searchQuery == "" {
		t.searchMatches = nil
		t.matchIndex = -1
		return
	}

	if t.flatMode {
		t.searchMatches = nil
		for i, entry := range t.flatEntries {
			if entry.MatchesSearch(t.searchQuery) {
				t.searchMatches = append(t.searchMatches, i)
			}
		}
		if len(t.searchMatches) > 0 {
			t.matchIndex = 0
			t.cursor = t.searchMatches[0]
			m.clampOffset()
		} else {
			t.matchIndex = -1
		}
		return
	}

	// Tree mode search
	matchingNodes := tree.FindMatchingNodes(t.root, t.searchQuery)
	if len(matchingNodes) == 0 {
		t.searchMatches = nil
		t.matchIndex = -1
		return
	}

	for _, n := range matchingNodes {
		tree.ExpandToNode(n)
	}

	t.visible = tree.VisibleNodes(t.root)

	matchSet := make(map[*tree.Node]bool)
	for _, n := range matchingNodes {
		matchSet[n] = true
	}

	t.searchMatches = nil
	for i, entry := range t.visible {
		if !entry.IsClosing && matchSet[entry.Node] {
			t.searchMatches = append(t.searchMatches, i)
		}
	}

	if len(t.searchMatches) > 0 {
		t.matchIndex = 0
		t.cursor = t.searchMatches[0]
		m.clampOffset()
	}
}

func (m *Model) nextMatch() {
	t := m.active()
	if len(t.searchMatches) == 0 {
		return
	}
	t.matchIndex = (t.matchIndex + 1) % len(t.searchMatches)
	t.cursor = t.searchMatches[t.matchIndex]
	m.clampOffset()
}

func (m *Model) prevMatch() {
	t := m.active()
	if len(t.searchMatches) == 0 {
		return
	}
	t.matchIndex--
	if t.matchIndex < 0 {
		t.matchIndex = len(t.searchMatches) - 1
	}
	t.cursor = t.searchMatches[t.matchIndex]
	m.clampOffset()
}

func (m *Model) rebuildSearchMatches() {
	t := m.active()
	if t.searchQuery == "" {
		t.searchMatches = nil
		t.matchIndex = -1
		return
	}

	t.searchMatches = nil
	if t.flatMode {
		for i, entry := range t.flatEntries {
			if entry.MatchesSearch(t.searchQuery) {
				t.searchMatches = append(t.searchMatches, i)
			}
		}
	} else {
		for i, entry := range t.visible {
			if !entry.IsClosing && entry.Node.MatchesSearch(t.searchQuery) {
				t.searchMatches = append(t.searchMatches, i)
			}
		}
	}

	if len(t.searchMatches) == 0 {
		t.matchIndex = -1
	} else if t.matchIndex >= len(t.searchMatches) {
		t.matchIndex = len(t.searchMatches) - 1
	}
}

func (m Model) tabBarView() string {
	if len(m.tabs) <= 1 {
		return ""
	}

	var parts []string
	for i, tab := range m.tabs {
		label := tab.label
		// Truncate long labels
		if len(label) > 30 {
			label = label[:27] + "..."
		}
		// Mark error tabs
		if tab.errMsg != "" {
			label = label + " " + errorTabStyle.Render("!")
		}

		if i == m.activeTab {
			parts = append(parts, activeTabStyle.Render(label))
		} else {
			parts = append(parts, inactiveTabStyle.Render(label))
		}
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	// Pad to full width
	barWidth := lipgloss.Width(bar)
	if barWidth < m.width {
		bar += strings.Repeat(" ", m.width-barWidth)
	}
	return bar
}

func (m Model) headerView() string {
	t := m.tabs[m.activeTab]
	titleText := "Mr. Jsawn"
	if t.flatMode {
		titleText = "Mr. Jsawn (flat)"
	}
	title := titleStyle.Render(titleText)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	t := m.tabs[m.activeTab]
	pct := 0.0
	n := t.listLen()
	if n > 1 {
		pct = float64(t.cursor) / float64(n-1) * 100
	}
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", pct))
	infoWidth := lipgloss.Width(info)

	// When in command mode, show command input
	if t.commandMode {
		cmdInfo := searchStyle.Render(fmt.Sprintf(":%s█", t.commandInput))
		cmdWidth := lipgloss.Width(cmdInfo)
		lineWidth := max(0, m.width-cmdWidth-infoWidth)
		line := strings.Repeat("─", lineWidth)
		return lipgloss.JoinHorizontal(lipgloss.Center, cmdInfo, line, info)
	}

	// When in search mode, show search input instead of breadcrumb
	if t.searching {
		searchInfo := searchStyle.Render(fmt.Sprintf("/%s█", t.searchQuery))
		searchWidth := lipgloss.Width(searchInfo)
		lineWidth := max(0, m.width-searchWidth-infoWidth)
		line := strings.Repeat("─", lineWidth)
		return lipgloss.JoinHorizontal(lipgloss.Center, searchInfo, line, info)
	}

	// Breadcrumb path
	path := "."
	if t.flatMode {
		if t.cursor < len(t.flatEntries) {
			path = t.flatEntries[t.cursor].Path
		}
	} else {
		if t.cursor < len(t.visible) {
			path = tree.NodePath(t.visible[t.cursor].Node)
		}
	}

	// Search status
	var searchInfo string
	if t.searchQuery != "" && len(t.searchMatches) == 0 {
		searchInfo = noMatchStyle.Render(" [no matches]")
	} else if len(t.searchMatches) > 0 {
		searchInfo = searchStyle.Render(fmt.Sprintf(" [%d/%d]", t.matchIndex+1, len(t.searchMatches)))
	}

	searchInfoWidth := lipgloss.Width(searchInfo)
	availableForPath := m.width - infoWidth - searchInfoWidth - 1
	if availableForPath < 1 {
		availableForPath = 1
	}

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

	t := m.tabs[m.activeTab]

	// Tab bar (only shown when multiple tabs)
	tabBar := m.tabBarView()

	// Error tab rendering
	if t.errMsg != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true).
			Width(m.width).
			Align(lipgloss.Center)

		errContent := errStyle.Render(fmt.Sprintf("Error: %s", t.errMsg))

		// Vertically center the error
		padding := m.height / 2
		var lines []string
		for i := 0; i < padding; i++ {
			lines = append(lines, "")
		}
		lines = append(lines, errContent)
		for len(lines) < m.height {
			lines = append(lines, "")
		}
		content := strings.Join(lines, "\n")
		base := fmt.Sprintf("%s\n%s\n%s", m.headerView(), content, m.footerView())
		if tabBar != "" {
			base = tabBar + "\n" + base
		}
		return base
	}

	var lines []string
	total := t.listLen()
	end := t.offset + m.height
	if end > total {
		end = total
	}

	for i := t.offset; i < end; i++ {
		isCursor := i == t.cursor
		var line string
		if t.flatMode {
			line = tree.RenderFlatEntry(t.flatEntries[i], isCursor, m.width, t.searchQuery)
		} else {
			line = tree.RenderEntry(t.visible[i], isCursor, m.width, t.searchQuery)
		}
		lines = append(lines, line)
	}

	for len(lines) < m.height {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	base := fmt.Sprintf("%s\n%s\n%s", m.headerView(), content, m.footerView())
	if tabBar != "" {
		base = tabBar + "\n" + base
	}

	if t.overlayActive {
		return m.renderOverlay()
	}

	return base
}
