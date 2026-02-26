package schema

import (
	"fmt"
	"strings"
)

func GenerateTypeScript(st *SchemaType) string {
	var b strings.Builder
	generateTS(&b, st, "Root", 0)
	return b.String()
}

func generateTS(b *strings.Builder, st *SchemaType, name string, depth int) {
	switch st.Kind {
	case KindObject:
		ifaceName := st.Name
		if depth == 0 {
			ifaceName = name
		}
		fmt.Fprintf(b, "interface %s {\n", ifaceName)
		for _, f := range st.Fields {
			b.WriteString(strings.Repeat("  ", depth+1))
			fmt.Fprintf(b, "%s: ", f.JSONName)
			writeTSType(b, f.Type, depth+1)
			b.WriteString(";\n")
		}
		b.WriteString(strings.Repeat("  ", depth))
		b.WriteString("}")
		if depth == 0 {
			b.WriteString("\n")
		}

	case KindArray:
		if depth == 0 {
			fmt.Fprintf(b, "type %s = ", name)
		}
		writeTSType(b, st.ElemType, depth)
		b.WriteString("[]")
		if depth == 0 {
			b.WriteString(";\n")
		}

	default:
		if depth == 0 {
			fmt.Fprintf(b, "type %s = %s;\n", name, tsScalar(st.Kind))
		} else {
			b.WriteString(tsScalar(st.Kind))
		}
	}
}

func writeTSType(b *strings.Builder, st *SchemaType, depth int) {
	switch st.Kind {
	case KindObject:
		generateTS(b, st, st.Name, depth)
	case KindArray:
		writeTSType(b, st.ElemType, depth)
		b.WriteString("[]")
	default:
		b.WriteString(tsScalar(st.Kind))
	}
}

func tsScalar(k Kind) string {
	switch k {
	case KindString:
		return "string"
	case KindNumber, KindInt:
		return "number"
	case KindBool:
		return "boolean"
	case KindNull:
		return "null"
	case KindAny:
		return "any"
	default:
		return "any"
	}
}
