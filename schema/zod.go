package schema

import (
	"fmt"
	"strings"
)

func GenerateZod(st *SchemaType) string {
	var b strings.Builder
	generateZod(&b, st, "Root", 0)
	return b.String()
}

func generateZod(b *strings.Builder, st *SchemaType, name string, depth int) {
	switch st.Kind {
	case KindObject:
		schemaName := name + "Schema"
		if depth > 0 {
			schemaName = st.Name + "Schema"
		}
		fmt.Fprintf(b, "const %s = z.object({\n", schemaName)
		for _, f := range st.Fields {
			b.WriteString(strings.Repeat("  ", depth+1))
			fmt.Fprintf(b, "%s: ", f.JSONName)
			writeZodType(b, f.Type, depth+1)
			b.WriteString(",\n")
		}
		b.WriteString(strings.Repeat("  ", depth))
		b.WriteString("})")
		if depth == 0 {
			b.WriteString(";\n")
		}

	case KindArray:
		if depth == 0 {
			fmt.Fprintf(b, "const %sSchema = ", name)
		}
		b.WriteString("z.array(")
		writeZodType(b, st.ElemType, depth)
		b.WriteString(")")
		if depth == 0 {
			b.WriteString(";\n")
		}

	default:
		if depth == 0 {
			fmt.Fprintf(b, "const %sSchema = %s;\n", name, zodScalar(st.Kind))
		} else {
			b.WriteString(zodScalar(st.Kind))
		}
	}
}

func writeZodType(b *strings.Builder, st *SchemaType, depth int) {
	switch st.Kind {
	case KindObject:
		generateZod(b, st, st.Name, depth)
	case KindArray:
		b.WriteString("z.array(")
		writeZodType(b, st.ElemType, depth)
		b.WriteString(")")
	default:
		b.WriteString(zodScalar(st.Kind))
	}
}

func zodScalar(k Kind) string {
	switch k {
	case KindString:
		return "z.string()"
	case KindNumber:
		return "z.number()"
	case KindInt:
		return "z.number().int()"
	case KindBool:
		return "z.boolean()"
	case KindNull:
		return "z.null()"
	case KindAny:
		return "z.any()"
	default:
		return "z.any()"
	}
}
