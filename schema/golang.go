package schema

import (
	"fmt"
	"strings"
)

func GenerateGo(st *SchemaType) string {
	var b strings.Builder
	generateGo(&b, st, 0)
	return b.String()
}

func generateGo(b *strings.Builder, st *SchemaType, depth int) {
	switch st.Kind {
	case KindObject:
		if depth == 0 {
			fmt.Fprintf(b, "type %s struct {\n", st.Name)
		} else {
			b.WriteString("struct {\n")
		}
		for _, f := range st.Fields {
			b.WriteString(strings.Repeat("\t", depth+1))
			fmt.Fprintf(b, "%s ", f.Name)
			writeGoType(b, f.Type, depth+1)
			fmt.Fprintf(b, " `json:\"%s\"`\n", f.JSONName)
		}
		b.WriteString(strings.Repeat("\t", depth))
		b.WriteString("}")
		if depth == 0 {
			b.WriteString("\n")
		}

	case KindArray:
		if depth == 0 {
			fmt.Fprintf(b, "type %s ", "Root")
		}
		b.WriteString("[]")
		writeGoType(b, st.ElemType, depth)
		if depth == 0 {
			b.WriteString("\n")
		}

	default:
		if depth == 0 {
			fmt.Fprintf(b, "type Root %s\n", goScalar(st.Kind))
		} else {
			b.WriteString(goScalar(st.Kind))
		}
	}
}

func writeGoType(b *strings.Builder, st *SchemaType, depth int) {
	switch st.Kind {
	case KindObject:
		generateGo(b, st, depth)
	case KindArray:
		b.WriteString("[]")
		writeGoType(b, st.ElemType, depth)
	default:
		b.WriteString(goScalar(st.Kind))
	}
}

func goScalar(k Kind) string {
	switch k {
	case KindString:
		return "string"
	case KindNumber:
		return "float64"
	case KindInt:
		return "int"
	case KindBool:
		return "bool"
	case KindNull:
		return "interface{}"
	case KindAny:
		return "interface{}"
	default:
		return "interface{}"
	}
}
