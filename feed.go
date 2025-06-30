// Package beam provides a Go implementation of the BEAM (Blog Entry Aggregation Method) protocol
// Compatible with Go 1.18+
package beam

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	// Version defines the current BEAM protocol version
	Version = "1.0"

	// ContentTypeJSON is the recommended content type for BEAM feeds
	ContentTypeJSON = "application/json; charset=utf-8"

	// DefaultCacheControl is the recommended cache control header
	DefaultCacheControl = "public, max-age=3600"
)

// Author represents author information for feeds and entries
type Author struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Feed represents a complete BEAM feed
type Feed struct {
	Version     string     `json:"version"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	HomePageURL string     `json:"home_page_url,omitempty"`
	FeedURL     string     `json:"feed_url"`
	Language    string     `json:"language,omitempty"`
	Author      *Author    `json:"author,omitempty"`
	LastUpdated *time.Time `json:"last_updated,omitempty"`
	Items       []Entry    `json:"items"`
}

// NewFeed creates a new BEAM feed with required fields
func NewFeed(title, feedURL string) *Feed {
	now := time.Now().UTC()
	return &Feed{
		Version:     Version,
		Title:       title,
		FeedURL:     feedURL,
		LastUpdated: &now,
		Items:       make([]Entry, 0),
	}
}

// AddEntry adds an entry to the feed
func (f *Feed) AddEntry(entry Entry) {
	f.Items = append(f.Items, entry)
	now := time.Now().UTC()
	f.LastUpdated = &now
}

// SetAuthor sets the feed author
func (f *Feed) SetAuthor(name, email, url string) {
	f.Author = &Author{
		Name:  name,
		Email: email,
		URL:   url,
	}
}

// SetDescription sets the feed description
func (f *Feed) SetDescription(description string) {
	f.Description = description
}

// SetHomePageURL sets the home page URL
func (f *Feed) SetHomePageURL(url string) {
	f.HomePageURL = url
}

// SetLanguage sets the feed language
func (f *Feed) SetLanguage(language string) {
	f.Language = language
}

// Validate validates the feed structure
func (f *Feed) Validate() error {
	if f.Version != Version {
		return NewError("version", fmt.Sprintf("unsupported version: %s, expected: %s", f.Version, Version))
	}

	if strings.TrimSpace(f.Title) == "" {
		return NewError("title", "title is required")
	}

	if !isValidURL(f.FeedURL) {
		return NewError("feed_url", "feed_url must be a valid URL")
	}

	if f.HomePageURL != "" && !isValidURL(f.HomePageURL) {
		return NewError("home_page_url", "home_page_url must be a valid URL")
	}

	entryIDs := make(map[string]bool)
	for i, entry := range f.Items {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry %d: %w", i, err)
		}
		if entryIDs[entry.ID] {
			return NewError("items", fmt.Sprintf("duplicate entry ID: %s", entry.ID))
		}
		entryIDs[entry.ID] = true
	}

	return nil
}

// ToJSON serializes the feed to JSON
func (f *Feed) ToJSON() ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// FromJSON deserializes a feed from JSON
func FromJSON(data []byte) (*Feed, error) {
	var feed Feed
	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := feed.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &feed, nil
}

// FetchFeed fetches a BEAM feed from a URL
func FetchFeed(url string) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return FromJSON(body)
}

// CalculateReadingTime estimates reading time in minutes based on word count
// Assumes average reading speed of 200 words per minute
func CalculateReadingTime(content string) int {
	if content == "" {
		return 0
	}

	// Remove HTML tags for word counting
	re := regexp.MustCompile(`<[^>]*>`)
	plainText := re.ReplaceAllString(content, " ")

	// Count words
	words := strings.Fields(plainText)
	wordCount := len(words)

	// Calculate reading time (200 words per minute, minimum 1 minute)
	readingTime := wordCount / 200
	return max(readingTime, 1)
}

// isValidURL checks if a string is a valid HTTP/HTTPS URL
func isValidURL(str string) bool {
	if str == "" {
		return false
	}
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// FilterByTag returns entries that contain the specified tag
func (f *Feed) FilterByTag(tag string) (filtered []Entry) {
	for _, entry := range f.Items {
		for _, entryTag := range entry.Tags {
			if strings.EqualFold(entryTag, tag) {
				filtered = append(filtered, entry)
				break
			}
		}
	}
	return
}

// FilterByCategory returns entries that match the specified category
func (f *Feed) FilterByCategory(category string) (filtered []Entry) {
	for _, entry := range f.Items {
		if strings.EqualFold(entry.Category, category) {
			filtered = append(filtered, entry)
		}
	}
	return
}

// FilterByDateRange returns entries published within the specified date range
func (f *Feed) FilterByDateRange(start, end time.Time) (filtered []Entry) {
	for _, entry := range f.Items {
		if (entry.Published.Equal(start) || entry.Published.After(start)) &&
			(entry.Published.Equal(end) || entry.Published.Before(end)) {
			filtered = append(filtered, entry)
		}
	}
	return
}

// GetTags returns all unique tags from all entries
func (f *Feed) GetTags() []string {
	tagSet := make(map[string]bool)
	for _, entry := range f.Items {
		for _, tag := range entry.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

// GetCategories returns all unique categories from all entries
func (f *Feed) GetCategories() []string {
	categorySet := make(map[string]bool)
	for _, entry := range f.Items {
		if entry.Category != "" {
			categorySet[entry.Category] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}

	return categories
}

// ServeHTTP implements http.Handler for serving BEAM feeds
func (f *Feed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.Header().Set("Cache-Control", DefaultCacheControl)

	if f.LastUpdated != nil {
		w.Header().Set("Last-Modified", f.LastUpdated.Format(http.TimeFormat))
	}

	// Generate ETag based on last updated time and item count
	etag := fmt.Sprintf(`"beam-%d-%d"`, f.LastUpdated.Unix(), len(f.Items))
	w.Header().Set("ETag", etag)

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	data, err := f.ToJSON()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// HomePage ...
func (f *Feed) HomePage(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>BEAM Feed Example</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; }
				.feed-info { margin: 20px 0; }
				.entry { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
				.tags { color: #666; font-size: 0.9em; }
			</style>
		</head>
		<body>
			<h1>BEAM Feed Example</h1>
			<p>This is a demonstration of the BEAM (Blog Entry Aggregation Method) protocol.</p>
			<p><strong>Feed URL:</strong> <a href="/feed.json">/feed.json</a></p> <hr/></br>
			
			<div class="feed-info">
				<h2>` + f.Title + `</h2>
				<p>` + f.Description + `</p>
				<p><strong>Entries:</strong> ` + fmt.Sprintf("%d", len(f.Items)) + `</p>
			</div>
			
			<h3>Recent Entries</h3>`

	for _, entry := range f.Items {
		html += fmt.Sprintf(`<div class="entry"><h4><a href="%s">%s</a></h4><p>%s</p>
			<div class="tags"><strong>Category:</strong> %s<br><strong>Tags:</strong> %v<br>
			<strong>Reading time:</strong> %d minutes<br><strong>Published:</strong> %s</div></div>`,
			entry.URL, entry.Title, entry.Summary, entry.Category, entry.Tags,
			entry.ReadingTime, entry.Published.Local().Format("2006-01-02 15:04"))
	}

	html += `</body></html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ComputeFeedHash calculates the SHA-256 hash of the serialized feed content.
// It can be used to verify the integrity and trustworthiness of the data.
func (f *Feed) ComputeFeedHash() (string, error) {
	data, err := json.Marshal(f.Items)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// ValidateFeedHash verifies the computed hash of the feed against an expected hash.
// If the hashes do not match, it indicates that the feed data may not be trustworthy.
func (f *Feed) ValidateFeedHash(expectedHash string) error {
	actualHash, err := f.ComputeFeedHash()
	if err != nil {
		return err
	}
	if actualHash != expectedHash {
		return errors.New("feed hash mismatch: data may not be trustworthy")
	}
	return nil
}
