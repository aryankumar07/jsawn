/*
Copyright © 2025 Aryan Kumar aryan.pageme@gmail.com
*/
package main

import "github.com/aryankumar07/jsawn/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
