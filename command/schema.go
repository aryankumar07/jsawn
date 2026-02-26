package command

import (
	"github.com/aryankumar07/jsawn/schema"
)

func init() {
	Register("schema", schemaHandler)
}

func schemaHandler(ctx Context) Result {
	if len(ctx.Args) == 0 {
		return Result{Error: "usage: schema <go|ts|zod>"}
	}

	st := schema.Infer(ctx.Root)

	switch ctx.Args[0] {
	case "go":
		return Result{Content: schema.GenerateGo(st)}
	case "ts":
		return Result{Content: schema.GenerateTypeScript(st)}
	case "zod":
		return Result{Content: schema.GenerateZod(st)}
	default:
		return Result{Error: "unknown format: " + ctx.Args[0] + " (use go, ts, or zod)"}
	}
}
