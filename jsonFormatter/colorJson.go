package jsonFormatter

import (
	"regexp"

	"github.com/fatih/color"
)

func ColorizeJSONWithLib(jsonStr string) string {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	re := regexp.MustCompile(`"([^"]+)"\s*:`)
	jsonStr = re.ReplaceAllStringFunc(jsonStr, func(s string) string {
		return cyan(s[:len(s)-1]) + ":"
	})

	re = regexp.MustCompile(`:\s*"([^"]*)"`)
	jsonStr = re.ReplaceAllStringFunc(jsonStr, func(s string) string {
		return ": " + green(s[2:])
	})

	re = regexp.MustCompile(`:\s*(-?\d+\.?\d*)`)
	jsonStr = re.ReplaceAllStringFunc(jsonStr, func(s string) string {
		return ": " + yellow(s[2:])
	})

	re = regexp.MustCompile(`:\s*(true|false)`)
	jsonStr = re.ReplaceAllStringFunc(jsonStr, func(s string) string {
		return ": " + magenta(s[2:])
	})

	re = regexp.MustCompile(`:\s*(null)`)
	jsonStr = re.ReplaceAllStringFunc(jsonStr, func(s string) string {
		return ": " + red(s[2:])
	})

	return jsonStr
}
