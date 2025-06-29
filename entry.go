package beam

import (
	"strings"
	"time"
)

// Entry represents a single blog post entry
type Entry struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content,omitempty"`
	Summary     string     `json:"summary,omitempty"`
	URL         string     `json:"url"`
	Published   time.Time  `json:"published"`
	Updated     *time.Time `json:"updated,omitempty"`
	Author      *Author    `json:"author,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Category    string     `json:"category,omitempty"`
	Image       string     `json:"image,omitempty"`
	ReadingTime int        `json:"reading_time,omitempty"`
}

// NewEntry creates a new entry with required fields
func NewEntry(id, title, url string, published time.Time) Entry {
	return Entry{
		ID:        id,
		Title:     title,
		URL:       url,
		Published: published.UTC(),
		Tags:      make([]string, 0),
	}
}

// SetAuthor sets the entry author
func (e *Entry) SetAuthor(name, email, url string) {
	e.Author = &Author{
		Name:  name,
		Email: email,
		URL:   url,
	}
}

// SetContent sets the entry content and automatically calculates reading time
func (e *Entry) SetContent(content string) {
	e.Content, e.ReadingTime = content, CalculateReadingTime(content)
}

// SetSummary sets the entry summary
func (e *Entry) SetSummary(summary string) {
	e.Summary = summary
}

// SetTags sets the entry tags
func (e *Entry) SetTags(tags ...string) {
	e.Tags = tags
}

// SetCategory sets the entry category
func (e *Entry) SetCategory(category string) {
	e.Category = category
}

// SetImage sets the entry featured image
func (e *Entry) SetImage(imageURL string) {
	e.Image = imageURL
}

// SetUpdated sets the entry updated timestamp
func (e *Entry) SetUpdated(updated time.Time) {
	utcTime := updated.UTC()
	e.Updated = &utcTime
}

// Validate validates the entry structure
func (e *Entry) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return NewError("id", "id is required")
	}

	if strings.TrimSpace(e.Title) == "" {
		return NewError("title", "title is required")
	}

	if !isValidURL(e.URL) {
		return NewError("url", "url must be a valid URL")
	}

	if e.Published.IsZero() {
		return NewError("published", "published timestamp is required")
	}

	if e.Image != "" && !isValidURL(e.Image) {
		return NewError("image", "image must be a valid URL")
	}

	return nil
}
