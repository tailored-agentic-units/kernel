package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/tailored-agentic-units/kernel/kernel"
)

func main() {
	var (
		configFile    = flag.String("config", "", "Path to kernel config JSON file (required)")
		prompt        = flag.String("prompt", "", "Prompt to send to the agent (required)")
		systemPrompt  = flag.String("system-prompt", "", "System prmopt (overrides config)")
		memoryPath    = flag.String("memory", "", "Path to memory directory (overrides config)")
		maxIterations = flag.Int("max-iterations", -1, "Maximum loop iterations; 0 for unlimited (overrides config)")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging to stderr")
	)
	flag.Parse()

	if *configFile == "" || *prompt == "" {
		fmt.Fprintln(os.Stderr, "Usage: kernel -config <file> -prompt <text>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := kernel.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *systemPrompt != "" {
		cfg.SystemPrompt = *systemPrompt
	}
	if *memoryPath != "" {
		cfg.Memory.Path = *memoryPath
	}
	if *maxIterations >= 0 {
		cfg.MaxIterations = *maxIterations
	}

	var logger *slog.Logger
	if *verbose {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	registerBuiltinTools()

	runtime, err := kernel.New(cfg, kernel.WithLogger(logger))
	if err != nil {
		log.Fatalf("Failed to create kernel runtime: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := runtime.Run(ctx, *prompt)
	if err != nil {
		log.Fatalf("Kernel run failed: %v", err)
	}

	fmt.Printf("Response: %s\n", result.Response)

	if len(result.ToolCalls) > 0 {
		fmt.Println("\nTool Calls:")
		for i, tc := range result.ToolCalls {
			fmt.Printf("  [%d] %s(%s)\n", i+1, tc.Function.Name, tc.Function.Arguments)
			if tc.IsError {
				fmt.Printf("    error: %s\n", tc.Result)
			} else if len(tc.Result) > 200 {
				fmt.Printf("    -> %s...\n", tc.Result[:200])
			} else {
				fmt.Printf("    -> %s\n", tc.Result)
			}
		}
	}

	fmt.Printf("\nIterations: %d\n", result.Iterations)
}
