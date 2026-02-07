package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/tailored-agentic-units/kernel/orchestrate/config"
	"github.com/tailored-agentic-units/kernel/orchestrate/observability"
	"github.com/tailored-agentic-units/kernel/orchestrate/workflows"
	"github.com/tailored-agentic-units/kernel/agent"
	agentconfig "github.com/tailored-agentic-units/kernel/core/config"
)

type ProductReview struct {
	ID      int
	Product string
	Review  string
}

type SentimentResult struct {
	ReviewID   int
	Product    string
	Review     string
	Sentiment  string
	Analysis   string
	ProcessedAt time.Time
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Product Review Sentiment Analysis - Parallel Execution Example ===")
	fmt.Println()

	// ============================================================================
	// 1. Configure Observer
	// ============================================================================
	fmt.Println("1. Configuring observability...")

	slogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogLogger := slog.New(slogHandler)
	slogObserver := observability.NewSlogObserver(slogLogger)
	observability.RegisterObserver("slog", slogObserver)

	fmt.Printf("  ✓ Registered slog observer\n")
	fmt.Println()

	// ============================================================================
	// 2. Load Agent Configuration
	// ============================================================================
	fmt.Println("2. Loading agent configuration...")

	llamaConfig, err := agentconfig.LoadAgentConfig("examples/phase-05-parallel-execution/config.llama.json")
	if err != nil {
		log.Fatalf("Failed to load llama config: %v", err)
	}

	llamaConfig.Name = "sentiment-analyst"
	llamaConfig.SystemPrompt = `You are an expert sentiment analysis system.
Analyze product reviews and classify sentiment as positive, neutral, or negative.
Respond in format: "SENTIMENT" where SENTIMENT is one word: positive, neutral, or negative.`

	sentimentAgent, err := agent.New(llamaConfig)
	if err != nil {
		log.Fatalf("Failed to create sentiment agent: %v", err)
	}

	fmt.Printf("  ✓ Created sentiment-analyst agent (llama3.2:3b)\n")
	fmt.Println()

	// ============================================================================
	// 3. Prepare Product Reviews
	// ============================================================================
	fmt.Println("3. Preparing product reviews...")

	reviews := []ProductReview{
		{ID: 1, Product: "Wireless Mouse", Review: "Excellent mouse! Great battery life and very responsive. Highly recommend!"},
		{ID: 2, Product: "USB-C Cable", Review: "Cable stopped working after 2 weeks. Very disappointed with the quality."},
		{ID: 3, Product: "Keyboard", Review: "Keys feel nice but the backlight is inconsistent. Works okay for the price."},
		{ID: 4, Product: "Monitor Stand", Review: "Perfect height adjustment. Sturdy and well-built. Worth every penny."},
		{ID: 5, Product: "Webcam", Review: "Video quality is mediocre in low light. Audio is decent but not great."},
		{ID: 6, Product: "Headphones", Review: "Amazing sound quality! Comfortable for long sessions. Best purchase this year."},
		{ID: 7, Product: "Phone Case", Review: "Flimsy material. Doesn't provide good protection. Would not buy again."},
		{ID: 8, Product: "Laptop Stand", Review: "Does the job. Nothing special but no complaints. Good for basic use."},
		{ID: 9, Product: "Charging Dock", Review: "Fast charging and multiple ports. Very convenient for desk setup."},
		{ID: 10, Product: "Screen Protector", Review: "Terrible application process. Air bubbles everywhere. Waste of money."},
		{ID: 11, Product: "Stylus Pen", Review: "Precise and responsive. Great for digital art. Battery lasts forever."},
		{ID: 12, Product: "Cable Organizer", Review: "Simple but effective. Keeps desk tidy. Exactly what I needed."},
	}

	fmt.Printf("  ✓ Loaded %d product reviews\n", len(reviews))
	fmt.Println()

	// ============================================================================
	// 4. Configure Parallel Processing
	// ============================================================================
	fmt.Println("4. Configuring parallel processing...")

	parallelConfig := config.DefaultParallelConfig()
	parallelConfig.Observer = "slog"
	failFast := false
	parallelConfig.FailFastNil = &failFast
	parallelConfig.WorkerCap = 4

	fmt.Printf("  ✓ Parallel configuration ready\n")
	fmt.Printf("    Worker cap: %d\n", parallelConfig.WorkerCap)
	fmt.Printf("    Fail-fast: %v (collect all errors)\n", parallelConfig.FailFast())
	fmt.Println()

	// ============================================================================
	// 5. Define Task Processor
	// ============================================================================
	fmt.Println("5. Defining sentiment analysis processor...")

	taskProcessor := func(ctx context.Context, review ProductReview) (SentimentResult, error) {
		prompt := fmt.Sprintf("Analyze sentiment of this review: \"%s\"", review.Review)

		response, err := sentimentAgent.Chat(ctx, prompt)
		if err != nil {
			return SentimentResult{}, fmt.Errorf("sentiment analysis failed: %w", err)
		}

		analysis := response.Content()
		sentimentWord := "unknown"

		if len(analysis) > 0 {
			sentimentWord = analysis
		}

		return SentimentResult{
			ReviewID:   review.ID,
			Product:    review.Product,
			Review:     review.Review,
			Sentiment:  sentimentWord,
			Analysis:   analysis,
			ProcessedAt: time.Now(),
		}, nil
	}

	fmt.Printf("  ✓ Task processor defined\n")
	fmt.Println()

	// ============================================================================
	// 6. Define Progress Callback
	// ============================================================================
	fmt.Println("6. Configuring progress tracking...")

	totalReviews := len(reviews)

	progressCallback := func(completed int, total int, result SentimentResult) {
		percentage := (completed * 100) / total
		fmt.Printf("\n  Progress: %d/%d reviews analyzed (%d%%) - Latest: Review #%d (%s)\n",
			completed, total, percentage, result.ReviewID, result.Sentiment)
	}

	fmt.Printf("  ✓ Progress callback configured\n")
	fmt.Println()

	// ============================================================================
	// 7. Execute Parallel Analysis
	// ============================================================================
	fmt.Println("7. Executing parallel sentiment analysis...")
	fmt.Println()

	fmt.Printf("  Processing %d reviews concurrently...\n", len(reviews))
	fmt.Println()

	startTime := time.Now()

	result, err := workflows.ProcessParallel(
		ctx,
		parallelConfig,
		reviews,
		taskProcessor,
		progressCallback,
	)

	duration := time.Since(startTime)

	fmt.Println()

	successCount := len(result.Results)
	errorCount := len(result.Errors)

	if err != nil {
		fmt.Printf("  ⚠ Parallel processing completed with errors: %v\n", err)
	} else {
		fmt.Println("  ✓ Parallel processing completed successfully")
	}
	fmt.Println()

	// ============================================================================
	// 8. Display Results
	// ============================================================================
	fmt.Println("8. Sentiment Analysis Results")
	fmt.Println()

	fmt.Printf("   Analyzed %d/%d reviews successfully\n", successCount, totalReviews)
	if errorCount > 0 {
		fmt.Printf("   Errors: %d\n", errorCount)
	}
	fmt.Println()

	fmt.Println("   Individual Results (in original order):")
	fmt.Println()

	resultMap := make(map[int]SentimentResult)
	for _, r := range result.Results {
		resultMap[r.ReviewID] = r
	}

	errorMap := make(map[int]error)
	for _, taskErr := range result.Errors {
		errorMap[taskErr.Item.ID] = taskErr.Err
	}

	for _, review := range reviews {
		fmt.Printf("   [%d] %s\n", review.ID, review.Product)
		fmt.Printf("       Review: %s\n", review.Review)

		if sentResult, exists := resultMap[review.ID]; exists {
			fmt.Printf("       ✓ Sentiment: %s\n", sentResult.Sentiment)
		} else if err, hasError := errorMap[review.ID]; hasError {
			fmt.Printf("       ✗ Error: %v\n", err)
		} else {
			fmt.Printf("       ✗ No result\n")
		}
		fmt.Println()
	}

	// ============================================================================
	// 9. Sentiment Summary
	// ============================================================================
	fmt.Println("9. Sentiment Summary")
	fmt.Println()

	positiveCount := 0
	neutralCount := 0
	negativeCount := 0

	for _, r := range result.Results {
		sentiment := r.Sentiment
		if len(sentiment) > 0 {
			firstChar := sentiment[0]
			if firstChar == 'p' || firstChar == 'P' {
				positiveCount++
			} else if firstChar == 'n' && (len(sentiment) > 2 && sentiment[2] == 'g' || sentiment[2] == 'G') {
				negativeCount++
			} else {
				neutralCount++
			}
		}
	}

	fmt.Printf("   Positive: %d (%.1f%%)\n", positiveCount, float64(positiveCount)/float64(successCount)*100)
	fmt.Printf("   Neutral:  %d (%.1f%%)\n", neutralCount, float64(neutralCount)/float64(successCount)*100)
	fmt.Printf("   Negative: %d (%.1f%%)\n", negativeCount, float64(negativeCount)/float64(successCount)*100)
	fmt.Println()

	// ============================================================================
	// 10. Error Analysis
	// ============================================================================
	if errorCount > 0 {
		fmt.Println("10. Error Analysis")
		fmt.Println()

		fmt.Printf("   Total errors: %d/%d reviews\n", errorCount, totalReviews)
		fmt.Println()

		fmt.Println("   Failed reviews:")
		for _, taskErr := range result.Errors {
			fmt.Printf("     - Review %d (%s): %v\n", taskErr.Item.ID, taskErr.Item.Product, taskErr.Err)
		}
		fmt.Println()
	}

	// ============================================================================
	// 11. Performance Metrics
	// ============================================================================
	section := 10
	if errorCount > 0 {
		section = 11
	}

	fmt.Printf("%d. Performance Metrics\n", section)
	fmt.Println()

	avgTimePerReview := duration / time.Duration(successCount)
	reviewsPerSecond := float64(successCount) / duration.Seconds()

	fmt.Printf("   Total Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("   Reviews Processed: %d/%d\n", successCount, totalReviews)
	fmt.Printf("   Success Rate: %.1f%%\n", (float64(successCount)/float64(totalReviews))*100)
	fmt.Printf("   Average Time per Review: %v\n", avgTimePerReview.Round(time.Millisecond))
	fmt.Printf("   Throughput: %.2f reviews/second\n", reviewsPerSecond)
	fmt.Println()

	fmt.Printf("   Concurrency:\n")
	fmt.Printf("     Worker Cap: %d\n", parallelConfig.WorkerCap)
	sequentialEstimate := avgTimePerReview * time.Duration(successCount)
	speedup := sequentialEstimate.Seconds() / duration.Seconds()
	fmt.Printf("     Estimated Speedup: %.1fx\n", speedup)
	fmt.Println()

	fmt.Println("=== Sentiment Analysis Complete ===")
}
