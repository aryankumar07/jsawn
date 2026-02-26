package schema

import (
	"math"
	"strings"
	"unicode"

	"github.com/aryankumar07/jsawn/tree"
)

type Kind int

const (
	KindString Kind = iota
	KindNumber
	KindInt
	KindBool
	KindNull
	KindObject
	KindArray
	KindAny
)

type Field struct {
	Name       string
	JSONName   string
	Type       *SchemaType
}

type SchemaType struct {
	Kind     Kind
	Name     string   // struct/interface name for objects
	Fields   []Field  // for KindObject
	ElemType *SchemaType // for KindArray
}

type inferContext struct {
	seen map[string]int
}

func Infer(root *tree.Node) *SchemaType {
	ctx := &inferContext{seen: map[string]int{}}
	return ctx.infer(root, "Root")
}

func (c *inferContext) infer(n *tree.Node, name string) *SchemaType {
	switch n.Type {
	case tree.TypeObject:
		objName := c.uniqueName(name)
		st := &SchemaType{Kind: KindObject, Name: objName}
		for _, child := range n.Children {
			fieldName := toPascalCase(child.Key)
			if fieldName == "" {
				fieldName = "Field"
			}
			ft := c.infer(child, fieldName)
			st.Fields = append(st.Fields, Field{
				Name:     fieldName,
				JSONName: child.Key,
				Type:     ft,
			})
		}
		return st

	case tree.TypeArray:
		st := &SchemaType{Kind: KindArray}
		if len(n.Children) == 0 {
			st.ElemType = &SchemaType{Kind: KindAny}
			return st
		}
		// Infer from first element
		first := c.infer(n.Children[0], name+"Item")
		// Check for heterogeneous arrays
		if len(n.Children) > 1 {
			for _, child := range n.Children[1:] {
				other := c.infer(child, name+"Item")
				if !typesCompatible(first, other) {
					st.ElemType = &SchemaType{Kind: KindAny}
					return st
				}
			}
		}
		st.ElemType = first
		return st

	case tree.TypeString:
		return &SchemaType{Kind: KindString}

	case tree.TypeNumber:
		v := n.Value.(float64)
		if v == math.Trunc(v) && !math.IsInf(v, 0) && !math.IsNaN(v) {
			return &SchemaType{Kind: KindInt}
		}
		return &SchemaType{Kind: KindNumber}

	case tree.TypeBool:
		return &SchemaType{Kind: KindBool}

	case tree.TypeNull:
		return &SchemaType{Kind: KindNull}
	}

	return &SchemaType{Kind: KindAny}
}

func (c *inferContext) uniqueName(name string) string {
	count, exists := c.seen[name]
	if !exists {
		c.seen[name] = 1
		return name
	}
	c.seen[name] = count + 1
	suffixed := name + itoa(count+1)
	return suffixed
}

func typesCompatible(a, b *SchemaType) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind == KindObject {
		if len(a.Fields) != len(b.Fields) {
			return false
		}
		for i := range a.Fields {
			if a.Fields[i].JSONName != b.Fields[i].JSONName {
				return false
			}
		}
	}
	return true
}

func toPascalCase(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
