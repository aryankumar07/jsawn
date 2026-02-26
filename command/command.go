package command

import (
	"strings"

	"github.com/aryankumar07/jsawn/tree"
)

type Context struct {
	Root *tree.Node
	Args []string
}

type Result struct {
	Content string
	Error   string
}

type Handler func(Context) Result

var registry = map[string]Handler{}

func Register(name string, h Handler) {
	registry[name] = h
}

func Execute(raw string, root *tree.Node) Result {
	raw = strings.TrimSpace(raw)
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return Result{Error: "empty command"}
	}

	name := parts[0]
	h, ok := registry[name]
	if !ok {
		return Result{Error: "unknown command: " + name}
	}

	return h(Context{
		Root: root,
		Args: parts[1:],
	})
}
