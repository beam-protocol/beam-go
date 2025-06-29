package beam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// FeedSource represents a source feed with metadata
type FeedSource struct {
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LastFetch   time.Time `json:"last_fetch"`
	Status      string    `json:"status"` // "active", "error", "timeout"
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// Aggregator manages multiple BEAM feeds and creates aggregated content
type Aggregator struct {
	Sources        []FeedSource `json:"sources"`
	AggregatedFeed *Feed        `json:"-"`
	mu             sync.RWMutex
	fetchTimeout   time.Duration
}

// NewAggregator creates a new feed aggregator
func NewAggregator(title, feedURL string) *Aggregator {
	return &Aggregator{
		Sources:        make([]FeedSource, 0),
		AggregatedFeed: NewFeed(title, feedURL),
		fetchTimeout:   30 * time.Second,
	}
}

// AddSource adds a new feed source to the aggregator
func (a *Aggregator) AddSource(url, name, description string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.Sources = append(a.Sources, FeedSource{
		URL:         url,
		Name:        name,
		Description: description,
		Status:      "new",
	})

	fmt.Printf("Added source: %s (%s)\n", name, url)
}

// FetchAllFeeds fetches all source feeds concurrently
func (a *Aggregator) FetchAllFeeds() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	fmt.Printf("Fetching %d source feeds...\n", len(a.Sources))

	type fetchResult struct {
		index int
		feed  *Feed
		err   error
	}

	results := make(chan fetchResult, len(a.Sources))
	for i, source := range a.Sources {
		go func(index int, src FeedSource) {
			feed, err := a.fetchFeedWithTimeout(src.URL)
			results <- fetchResult{index: index, feed: feed, err: err}
		}(i, source)
	}

	var allEntries []Entry
	successCount := 0

	// Wait for all goroutines to complete
	for i := 0; i < len(a.Sources); i++ {
		result := <-results
		a.Sources[result.index].LastFetch = time.Now()
		if result.err != nil {
			a.Sources[result.index].Status = "error"
			a.Sources[result.index].ErrorMsg = result.err.Error()
			fmt.Printf("Failed to fetch %s: %v\n", a.Sources[result.index].Name, result.err)
			continue
		}

		a.Sources[result.index].Status = "active"
		a.Sources[result.index].ErrorMsg = ""
		successCount++

		// Add source information to entries and collect them
		for _, entry := range result.feed.Items {
			// Create a copy of the entry with source information
			enrichedEntry := entry

			// Add source information as custom fields (using underscore prefix for extensions)
			// Note: In a real implementation, you might want to extend the Entry struct
			// For now, we'll add source info to the summary if it exists
			if enrichedEntry.Summary != "" {
				enrichedEntry.Summary = fmt.Sprintf("[%s] %s", a.Sources[result.index].Name, enrichedEntry.Summary)
			} else {
				enrichedEntry.Summary = fmt.Sprintf("From %s", a.Sources[result.index].Name)
			}
			allEntries = append(allEntries, enrichedEntry)
		}

		fmt.Printf("✓ Fetched %d entries from %s\n", len(result.feed.Items), a.Sources[result.index].Name)
	}

	fmt.Printf("Successfully fetched %d/%d feeds\n", successCount, len(a.Sources))

	// Sort entries by publication date (newest first)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Published.After(allEntries[j].Published)
	})

	// Limit to most recent entries (e.g., last 100)
	maxEntries := 100
	if len(allEntries) > maxEntries {
		allEntries = allEntries[:maxEntries]
	}

	// Create new aggregated feed
	aggregatedFeed := NewFeed(a.AggregatedFeed.Title, a.AggregatedFeed.FeedURL)
	aggregatedFeed.SetDescription(fmt.Sprintf("Aggregated content from %d sources", len(a.Sources)))
	aggregatedFeed.SetLanguage("en-US")

	// Add all entries to the aggregated feed
	for _, entry := range allEntries {
		aggregatedFeed.AddEntry(entry)
	}

	a.AggregatedFeed = aggregatedFeed
	return nil
}

// fetchFeedWithTimeout fetches a feed with a timeout
func (a *Aggregator) fetchFeedWithTimeout(url string) (*Feed, error) {
	client := &http.Client{Timeout: a.fetchTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	var feed Feed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w", err)
	}

	if err := feed.Validate(); err != nil {
		return nil, fmt.Errorf("feed validation error: %w", err)
	}

	return &feed, nil
}

// GetStats returns aggregation statistics
func (a *Aggregator) GetStats() map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()

	activeCount := 0
	errorCount := 0
	totalEntries := 0

	for _, source := range a.Sources {
		switch source.Status {
		case "active":
			activeCount++
		case "error":
			errorCount++
		}
	}

	if a.AggregatedFeed != nil {
		totalEntries = len(a.AggregatedFeed.Items)
	}

	return map[string]any{
		"total_sources":  len(a.Sources),
		"active_sources": activeCount,
		"error_sources":  errorCount,
		"total_entries":  totalEntries,
		"last_updated":   a.AggregatedFeed.LastUpdated,
	}
}

// ServeHTTP implements http.Handler for the aggregator
func (a *Aggregator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.AggregatedFeed == nil {
		http.Error(w, "Aggregated feed not available", http.StatusServiceUnavailable)
		return
	}

	a.AggregatedFeed.ServeHTTP(w, r)
}

func (a *Aggregator) HomePage(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>BEAM Aggregator Example</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; background: #f9f9f9; }
				.feed-info { margin: 20px 0; }
				.entry { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; background: #fff; }
				.tags { color: #666; font-size: 0.9em; }
				.source { font-size: 0.9em; color: #888; }
			</style>
		</head>
		<body>
			<h1>BEAM Aggregator Example</h1>
			<p>This is a demonstration of the BEAM (Blog Entry Aggregation Method) protocol.</p>
			<p><strong>Feed URL:</strong> <a href="/feed.json">/feed.json</a></p> <hr/>
			<p>Last Updated: <strong>` + a.AggregatedFeed.LastUpdated.Local().Format(time.DateTime) + `</strong></p>
			<h2>Sources</h2>
			<ul>`
	for _, src := range a.Sources {
		html += fmt.Sprintf("<li><strong>%s</strong> - <a href=\"%s\">%s</a> <span class=\"source\">(%s)</span></li>", src.Name, src.URL, src.URL, src.Description)
	}
	html += `</ul><hr/>`

	html += `<h3>Recent Entries</h3>`
	if a.AggregatedFeed != nil && len(a.AggregatedFeed.Items) > 0 {
		maxEntries := 10
		if len(a.AggregatedFeed.Items) < maxEntries {
			maxEntries = len(a.AggregatedFeed.Items)
		}
		for i := 0; i < maxEntries; i++ {
			entry := a.AggregatedFeed.Items[i]
			html += `<div class="entry">`
			html += fmt.Sprintf("<h4><a href=\"%s\">%s</a></h4>", entry.URL, entry.Title)
			html += fmt.Sprintf("<div class=\"source\">Published: %s</div>", entry.Published.Format("2006-01-02 15:04"))
			if entry.Summary != "" {
				html += fmt.Sprintf("<p>%s</p>", entry.Summary)
			}
			if len(entry.Tags) > 0 {
				html += fmt.Sprintf("<div class=\"tags\">Tags: %v</div>", entry.Tags)
			}
			html += `</div>`
		}
	} else {
		html += `<p>No entries available.</p>`
	}

	html += `</body></html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// StartAutoRefresh starts automatic feed refresh in the background
func (a *Aggregator) StartAutoRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			fmt.Println("Starting automatic feed refresh...")
			if err := a.FetchAllFeeds(); err != nil {
				fmt.Printf("Auto-refresh error: %v\n", err)
			}
		}
	}()
	fmt.Printf("Auto-refresh started with interval: %v\n", interval)
}

// GetSources returns the list of feed sources
func (a *Aggregator) GetSources() []FeedSource {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Sources
}
