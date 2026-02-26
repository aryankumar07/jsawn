package tree

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type NodeType int

const (
	TypeObject NodeType = iota
	TypeArray
	TypeString
	TypeNumber
	TypeBool
	TypeNull
)

type Node struct {
	Type      NodeType
	Key       string // non-empty if child of object
	Value     any    // for primitives
	Children  []*Node
	Depth     int
	Collapsed bool
	Parent    *Node
}

type VisibleEntry struct {
	Node      *Node
	IsClosing bool // true = closing bracket line for this container
}

func Parse(data []byte) (*Node, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	root := buildNode(raw, "", 0, nil)
	return root, nil
}

func buildNode(v any, key string, depth int, parent *Node) *Node {
	n := &Node{
		Key:    key,
		Depth:  depth,
		Parent: parent,
	}

	switch val := v.(type) {
	case map[string]any:
		n.Type = TypeObject
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := buildNode(val[k], k, depth+1, n)
			n.Children = append(n.Children, child)
		}
	case []any:
		n.Type = TypeArray
		for i, item := range val {
			child := buildNode(item, "", depth+1, n)
			child.Key = fmt.Sprintf("%d", i)
			n.Children = append(n.Children, child)
		}
	case string:
		n.Type = TypeString
		n.Value = val
	case float64:
		n.Type = TypeNumber
		n.Value = val
	case bool:
		n.Type = TypeBool
		n.Value = val
	case nil:
		n.Type = TypeNull
	}

	return n
}

func (n *Node) IsContainer() bool {
	return n.Type == TypeObject || n.Type == TypeArray
}

func (n *Node) Toggle() {
	if n.IsContainer() {
		n.Collapsed = !n.Collapsed
	}
}

func (n *Node) ExpandAll() {
	n.Collapsed = false
	for _, c := range n.Children {
		c.ExpandAll()
	}
}

func (n *Node) CollapseAll() {
	if n.IsContainer() {
		n.Collapsed = true
	}
	for _, c := range n.Children {
		c.CollapseAll()
	}
}

func (n *Node) CollapseToDepth(d int) {
	if n.IsContainer() {
		n.Collapsed = n.Depth >= d
	}
	for _, c := range n.Children {
		c.CollapseToDepth(d)
	}
}

func VisibleNodes(root *Node) []VisibleEntry {
	var entries []VisibleEntry
	collectVisible(root, &entries)
	return entries
}

func collectVisible(n *Node, entries *[]VisibleEntry) {
	*entries = append(*entries, VisibleEntry{Node: n, IsClosing: false})

	if n.IsContainer() && !n.Collapsed && len(n.Children) > 0 {
		for _, child := range n.Children {
			collectVisible(child, entries)
		}
		*entries = append(*entries, VisibleEntry{Node: n, IsClosing: true})
	}
}

// MatchesSearch returns true if this node's key or value contains the query (case-insensitive).
func (n *Node) MatchesSearch(query string) bool {
	if query == "" {
		return false
	}
	q := strings.ToLower(query)
	if n.Key != "" && strings.Contains(strings.ToLower(n.Key), q) {
		return true
	}
	switch n.Type {
	case TypeString:
		if strings.Contains(strings.ToLower(n.Value.(string)), q) {
			return true
		}
	case TypeNumber:
		if strings.Contains(fmt.Sprintf("%g", n.Value.(float64)), q) {
			return true
		}
	case TypeBool:
		if strings.Contains(fmt.Sprintf("%t", n.Value.(bool)), q) {
			return true
		}
	case TypeNull:
		if strings.Contains("null", q) {
			return true
		}
	}
	return false
}

// FindMatchingNodes returns all nodes in the tree that match the search query.
func FindMatchingNodes(root *Node, query string) []*Node {
	var matches []*Node
	findMatches(root, query, &matches)
	return matches
}

func findMatches(n *Node, query string, matches *[]*Node) {
	if n.MatchesSearch(query) {
		*matches = append(*matches, n)
	}
	for _, c := range n.Children {
		findMatches(c, query, matches)
	}
}

// ExpandToNode expands all ancestors of the given node so it becomes visible.
func ExpandToNode(n *Node) {
	for p := n.Parent; p != nil; p = p.Parent {
		p.Collapsed = false
	}
}

// NodePath returns the jq-style dot notation path from root to this node.
// e.g. ".users[0].name" for an object child inside an array inside "users".
// Returns "." for the root node.
func NodePath(n *Node) string {
	var parts []string
	for cur := n; cur != nil; cur = cur.Parent {
		if cur.Parent == nil {
			break
		}
		if cur.Parent.Type == TypeArray {
			parts = append(parts, "["+cur.Key+"]")
		} else {
			parts = append(parts, "."+cur.Key)
		}
	}
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	if len(parts) == 0 {
		return "."
	}
	result := strings.Join(parts, "")
	if result[0] == '[' {
		return "." + result
	}
	return result
}

// FlatEntry represents a leaf node with its full jq-style path, used in flatten view.
type FlatEntry struct {
	Node *Node
	Path string
}

// FlattenLeaves walks the entire tree and returns all leaf (primitive) nodes
// with their full paths. Containers are excluded.
func FlattenLeaves(root *Node) []FlatEntry {
	var entries []FlatEntry
	flattenLeaves(root, &entries)
	return entries
}

func flattenLeaves(n *Node, entries *[]FlatEntry) {
	if !n.IsContainer() {
		*entries = append(*entries, FlatEntry{
			Node: n,
			Path: NodePath(n),
		})
		return
	}
	for _, c := range n.Children {
		flattenLeaves(c, entries)
	}
}

// MatchesSearch returns true if the flat entry's path or value contains the query.
func (e FlatEntry) MatchesSearch(query string) bool {
	if query == "" {
		return false
	}
	q := strings.ToLower(query)
	if strings.Contains(strings.ToLower(e.Path), q) {
		return true
	}
	return e.Node.MatchesSearch(query)
}
