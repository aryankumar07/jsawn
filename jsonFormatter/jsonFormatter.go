package jsonFormatter

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
)

func GetJsonDataFromPipe() (*string, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read stdin: %v", err)
		return nil, err
	}
	var stdout bytes.Buffer
	err = json.Indent(&stdout, input, "", "  ")
	if err != nil {
		return nil, err
	}
	colorString := ColorizeJSONWithLib(stdout.String())
	return &colorString, nil
}
