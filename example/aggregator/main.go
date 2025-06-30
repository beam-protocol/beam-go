package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beam-protocol/beam-go"
)

func main() {
	aggregator := beam.NewAggregator("Tech News Aggregator", "http://localhost:8181/feed.json")
	aggregator.AddSource("Feed 01", "Latest technology news and startup coverage", "http://localhost:8081/feed.json")
	aggregator.AddSource("Feed 02", "Social news for hackers and entrepreneurs", "http://localhost:8082/feed.json")

	fmt.Println("\n=== Initial Feed Fetch ===")
	if err := aggregator.FetchAllFeeds(); err != nil {
		log.Printf("Error during initial fetch: %v", err)
	}

	fmt.Println("\n=== Aggregation Statistics ===")
	for key, value := range aggregator.GetStats() {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Display some sample entries from aggregated feed
	fmt.Println("\n=== Sample Aggregated Entries ===")
	if aggregator.AggregatedFeed != nil {
		count := min(5, len(aggregator.AggregatedFeed.Items))
		for i, entry := range aggregator.AggregatedFeed.Items[:count] {
			fmt.Printf("%d. %s\n", i+1, entry.Title)
			fmt.Printf("   Published: %s\n", entry.Published.Format("2006-01-02 15:04"))
			fmt.Printf("   Summary: %s\n", entry.Summary)
			fmt.Printf("   Tags: %v\n", entry.Tags)
			fmt.Printf("   URL: %s\n\n", entry.URL)
		}
	}

	// Start auto-refresh every 5 seconds
	aggregator.StartAutoRefresh(5 * time.Second)

	http.HandleFunc("/", aggregator.HomePage)
	http.Handle("/feed.json", aggregator)

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aggregator.GetStats())
	})

	http.HandleFunc("/sources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aggregator.Sources)
	})

	log.Fatal(http.ListenAndServe(":8181", nil))
}
