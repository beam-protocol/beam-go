package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beam-protocol/beam-go"
)

func main() {
	feed := beam.NewFeed("Feed 01", "http://localhost:8081/feed.json")
	feed.SetDescription("Articles about web development and technology")
	feed.SetHomePageURL("https://localhost:8081")
	feed.SetLanguage("en-US")
	feed.SetAuthor("John Doe", "john@localhost:8081", "https://johndoe.dev")

	entry1 := beam.NewEntry("intro-to-react-2025", "Introduction to React in 2025", "https://localhost:8081/intro-to-react-2025", time.Now().Add(-24*time.Hour))
	entry1.SetContent("<p>React continues to evolve with new features and improvements. In this comprehensive guide, we'll explore the latest developments in React and how to get started with modern React development in 2025.</p><p>We'll cover hooks, concurrent features, and best practices for building performant React applications.</p>")
	entry1.SetSummary("A comprehensive guide to getting started with React in 2025")
	entry1.SetAuthor("John Doe", "john@localhost:8081", "")
	entry1.SetTags("react", "javascript", "frontend", "web-development")
	entry1.SetCategory("Web Development")
	entry1.SetImage("https://localhost:8081/images/react-2025.jpg")

	entry1.SetExtension(beam.KindViewsExtension, 1500)
	entry1.SetExtension(beam.KindCommentsExtension, []string{
		"Great insights on React!",
		"Looking forward to trying these features.",
		"Thanks for sharing!",
	})

	entry2 := beam.NewEntry("golang-best-practices", "Go Best Practices for 2025", "https://localhost:8081/golang-best-practices", time.Now().Add(-12*time.Hour))
	entry2.SetContent("<p>Go has become one of the most popular programming languages for backend development. Here are the best practices every Go developer should know in 2025.</p><p>We'll cover error handling, project structure, testing strategies, and performance optimization techniques.</p>")
	entry2.SetSummary("Essential Go best practices for modern development")
	entry2.SetAuthor("Jane Smith", "jane@localhost:8081", "")
	entry2.SetTags("golang", "backend", "best-practices", "programming")
	entry2.SetCategory("Backend Development")
	entry2.SetUpdated(time.Now().Add(-6 * time.Hour))

	entry3 := beam.NewEntry("devops-trends-2025", "DevOps Trends for 2025", "https://localhost:8081/devops-trends-2025", time.Now().Add(-1*time.Hour))
	entry3.SetContent("<p>DevOps practices are evolving with new tools and automation strategies. Discover what's new in the DevOps landscape for 2025, including GitOps, advanced CI/CD, and platform engineering.</p>")
	entry3.SetSummary("Emerging DevOps trends and tools for 2025")
	entry3.SetAuthor("Alice Johnson", "alice@localhost:8081", "")
	entry3.SetTags("devops", "automation", "ci/cd", "platform-engineering")
	entry3.SetCategory("DevOps")
	entry3.SetImage("https://localhost:8081/images/devops-2025.jpg")

	feed.AddEntry(entry1)
	feed.AddEntry(entry2)
	feed.AddEntry(entry3)

	if err := feed.Validate(); err != nil {
		log.Fatalf("Feed validation failed: %v", err)
	}

	fmt.Printf("Created feed with %d entries\n", len(feed.Items))
	fmt.Printf("Feed title: %s\n", feed.Title)
	fmt.Printf("Feed URL: %s\n", feed.FeedURL)

	fmt.Println("\n=== Starting HTTP Server ===")
	fmt.Println("Feed available at: http://localhost:8081/feed.json")
	fmt.Println("Press Ctrl+C to stop the server")

	http.Handle("/feed.json", feed)
	http.HandleFunc("/", feed.HomePage)

	http.ListenAndServe(":8081", nil)
}
