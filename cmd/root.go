/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aryankumar07/jsawn/jsonFormatter"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsawn",
	Short: "",
	Long:  ` `,
	Args:  cobra.ArbitraryArgs,
	Run:   runJsonFormatter,
}

func runJsonFormatter(cmd *cobra.Command, args []string) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("failed to check stdin: %v", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		jsonData, err := jsonFormatter.GetJsonDataFromPipe()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(*jsonData)
		return
	}

	fmt.Println("No JSON input detected. Use:")
	fmt.Println("  curl https://dummyjson.com/products | jsawn")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// write the flags for CLI here
}
