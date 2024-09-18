package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	db "github.com/pageton/miftahDB-go/db"
	"github.com/pageton/miftahDB-go/types"
)

const (
	QueryCount  = 100000
	KeyLength   = 10
	ValueLength = 50
)

func RandomString(length int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func FormatTime(ms float64) string {
	return fmt.Sprintf("%.2f", ms/1000)
}

func RunBenchmark(db *db.BaseMiftahDB, operation string, action func() error) (map[string]string, error) {
	start := time.Now()
	err := action()
	if err != nil {
		return nil, err
	}
	duration := time.Since(start).Milliseconds()

	averageLatency := float64(duration) / QueryCount
	operationsPerSecond := int(float64(QueryCount) / (float64(duration) / 1000))

	result := map[string]string{
		"Operation":        operation,
		"Total Time (s)":   FormatTime(float64(duration)),
		"Avg Latency (ms)": fmt.Sprintf("%.4f", averageLatency),
		"Ops/Second":       fmt.Sprintf("%d", operationsPerSecond),
	}
	return result, nil
}

func Benchmark(db *db.BaseMiftahDB) error {
	keys := make([]string, QueryCount)
	values := make([]string, QueryCount)
	for i := 0; i < QueryCount; i++ {
		keys[i] = RandomString(KeyLength)
		values[i] = RandomString(ValueLength)
	}

	entries := make([]types.Entry, QueryCount)
	for i := 0; i < QueryCount; i++ {
		entries[i] = types.Entry{
			Key:   keys[i],
			Value: values[i],
		}
	}

	benchmarks := []struct {
		name   string
		action func() error
	}{
		{"Set", func() error {
			for i := 0; i < QueryCount; i++ {
				if err := db.Set(keys[i], values[i], nil); err != nil {
					return err
				}
			}
			return nil
		}},
		{"Multi Set", func() error { return db.MultiSet(entries) }},
		{"Exists", func() error {
			for i := 0; i < QueryCount; i++ {
				db.Exists(keys[i])
			}
			return nil
		}},
		{"Get Expire", func() error {
			for i := 0; i < QueryCount; i++ {
				db.GetExpire(keys[i])
			}
			return nil
		}},
		{"Set Expire", func() error {
			expiration := time.Now().Add(24 * time.Hour)
			for i := 0; i < QueryCount; i++ {
				if err := db.SetExpire(keys[i], expiration); err != nil {
					return err
				}
			}
			return nil
		}},
		{"Get", func() error {
			for i := 0; i < QueryCount; i++ {
				db.Get(keys[i])
			}
			return nil
		}},
		{"Multi Get", func() error { db.MultiGet(keys); return nil }},
		{"Delete", func() error {
			for i := 0; i < QueryCount; i++ {
				db.Delete(keys[i])
			}
			return nil
		}},
		{"Multi Delete", func() error { db.MultiDelete(keys); return nil }},
	}

	results := make([][]string, 0)
	for _, bm := range benchmarks {
		result, err := RunBenchmark(db, bm.name, bm.action)
		if err != nil {
			return err
		}
		results = append(results, []string{
			result["Operation"],
			result["Total Time (s)"],
			result["Avg Latency (ms)"],
			result["Ops/Second"],
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Operation", "Total Time (s)", "Avg Latency (ms)", "Ops/Second"})
	for _, v := range results {
		table.Append(v)
	}
	table.Render()

	return db.Cleanup()
}

func main() {
	db, err := db.NewBaseMiftahDB(":memory:")
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	err = Benchmark(db)
	if err != nil {
		log.Fatalf("Benchmark failed: %v", err)
	}
}
