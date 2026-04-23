// Produces test events into the outbox table for local development and test
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/open-outbox/relay/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Bound to flags
	dbURL     string
	table     string
	topic     string
	batchSize int
	interval  time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "producer",
	Short: "Open Outbox Benchmark Producer",
}

func init() {
	_ = godotenv.Load()

	_ = godotenv.Load()

	// Flag for Database
	rootCmd.PersistentFlags().
		StringVar(
			&dbURL,
			"storage-url",
			os.Getenv("STORAGE_URL"),
			"Database URL (Default: STORAGE_URL)",
		)
	rootCmd.PersistentFlags().
		StringVar(
			&table,
			"table-name",
			getEnv("STORAGE_TABLE_NAME",
				config.DefaultTableName),
			"Table Name",
		)

	// Producer specific flags
	rootCmd.PersistentFlags().
		StringVar(
			&topic,
			"topic",
			getEnv("LOCAL_TEST_TOPIC", "openoutbox.events.v1"),
			"Target topic",
		)
	rootCmd.PersistentFlags().
		IntVar(
			&batchSize,
			"batch-size",
			getEnvInt("LOCAL_PRODUCER_BATCH_SIZE", 100),
			"Batch size",
		)
	rootCmd.PersistentFlags().
		DurationVar(
			&interval,
			"interval",
			getEnvDuration("LOCAL_PRODUCER_INTERVAL", 1*time.Second),
			"Interval between batches",
		)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}

// --- COMMAND: SEED ---
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Inject a fixed number of events and exit",
	RunE: func(cmd *cobra.Command, _ []string) error {
		count, _ := cmd.Flags().GetInt("count")
		pool, err := pgxpool.New(cmd.Context(), dbURL)
		if err != nil {
			return err
		}
		defer pool.Close()

		fmt.Printf("🚀 Seeding %d events into %s...\n", count, table)
		start := time.Now()

		for produced := 0; produced < count; {
			// Calculate if we should send a full batch or just what's left
			toSend := batchSize
			if remaining := count - produced; remaining < batchSize {
				toSend = remaining
			}

			if err := sendBatch(cmd.Context(), pool, toSend); err != nil {
				return err
			}

			produced += toSend
			fmt.Printf("\rInserted: %d/%d", produced, count)
		}

		fmt.Printf(
			"\n✅ Done in %v (Avg: %.0f eps)\n",
			time.Since(start),
			float64(count)/time.Since(start).Seconds(),
		)
		return nil
	},
}

func init() {
	seedCmd.Flags().IntP("count", "c", 100000, "Total events to produce")
	rootCmd.AddCommand(seedCmd)
}

// --- COMMAND: BENCH ---
var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Produce events continuously at a set interval",
	RunE: func(cmd *cobra.Command, _ []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")
		pool, err := pgxpool.New(cmd.Context(), dbURL)
		if err != nil {
			return err
		}
		defer pool.Close()

		// Handle OS interrupts
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Printf("🔥 Benchmark mode active: %d events every %v\n", batchSize, interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nStopping benchmark mode...")
				return nil
			case <-ticker.C:
				if err := sendBatch(ctx, pool, batchSize); err != nil { // Always uses full batchSize
					log.Printf("Batch error: %v", err)
				}
				fmt.Printf("%d events inserted\n", batchSize)
			}
		}
	},
}

func init() {
	benchCmd.Flags().DurationP("interval", "i", time.Second, "Interval between batches")
	rootCmd.AddCommand(benchCmd)
}

func sendBatch(
	ctx context.Context,
	pool *pgxpool.Pool,
	currentSize int,
) error { // Added currentSize
	query := fmt.Sprintf(`
        INSERT INTO %s (event_id, event_type, partition_key, payload, headers, status)
        VALUES ($1, $2, $3, $4, $5, 'PENDING')`, table)

	batch := &pgx.Batch{}
	for i := 0; i < currentSize; i++ { // Use the parameter, not the global
		userID := rand.Intn(100000)
		payload := fmt.Sprintf(`{"user_id": %d, "email": "user-%d@example.com"}`, userID, userID)

		batch.Queue(
			query,
			uuid.New(),
			topic,
			fmt.Sprintf("user-%d", userID),
			[]byte(payload),
			map[string]any{"trace_id": uuid.New().String()},
		)
	}

	return pool.SendBatch(ctx, batch).Close()
}
