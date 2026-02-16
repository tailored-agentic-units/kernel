package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tailored-agentic-units/kernel/core/protocol"
	"github.com/tailored-agentic-units/kernel/tools"
)

func registerBuiltinTools() {
	must(tools.Register(protocol.Tool{
		Name:        "datetime",
		Description: "Returns the current date and time in RFC3339 format.",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, handleDatetime))

	must(tools.Register(protocol.Tool{
		Name:        "read_file",
		Description: "Reads the contents of a file at the given path.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Absolute or relative path to the file to read.",
				},
			},
			"required": []string{"path"},
		},
	}, handleReadFile))

	must(tools.Register(protocol.Tool{
		Name:        "list_directory",
		Description: "Lists files and directories at the given path.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Abssolute or relative path to the directory to list.",
				},
			},
			"required": []string{"path"},
		},
	}, handleListDirectory))
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("failed to register tool: %v", err))
	}
}

func handleDatetime(_ context.Context, _ json.RawMessage) (tools.Result, error) {
	return tools.Result{Content: time.Now().Format(time.RFC3339)}, nil
}

func handleReadFile(_ context.Context, raw json.RawMessage) (tools.Result, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return tools.Result{Content: "invalid arguments: " + err.Error(), IsError: true}, nil
	}
	if args.Path == "" {
		return tools.Result{Content: "path is required", IsError: true}, nil
	}

	data, err := os.ReadFile(args.Path)
	if err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}
	return tools.Result{Content: string(data)}, nil
}

func handleListDirectory(_ context.Context, raw json.RawMessage) (tools.Result, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return tools.Result{Content: "invalid arguments: " + err.Error(), IsError: true}, nil
	}
	if args.Path == "" {
		args.Path = "."
	}

	entries, err := os.ReadDir(args.Path)
	if err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}

	var b strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		b.WriteString(name)
		b.WriteByte('\n')
	}
	return tools.Result{Content: b.String()}, nil
}
