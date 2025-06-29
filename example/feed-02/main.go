package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beam-protocol/beam-go"
)

func main() {
	feed := beam.NewFeed("Feed 02", "http://localhost:8082/feed.json")
	feed.SetDescription("Articles about web development and technology")
	feed.SetHomePageURL("http://localhost:8082")
	feed.SetLanguage("en-US")
	feed.SetAuthor("John Doe", "john@localhost:8082", "https://johndoe.dev")

	entry3 := beam.NewEntry("ai-trends-2025", "AI Trends to Watch in 2025", "https://localhost:8082/ai-trends-2025", time.Now().Add(-24*time.Hour))
	entry3.SetContent("<p>Artificial Intelligence continues to evolve rapidly. Discover the top trends shaping AI in 2025, from generative models to ethical AI frameworks.</p>")
	entry3.SetSummary("Key AI trends and predictions for 2025")
	entry3.SetAuthor("Alice Johnson", "alice@localhost:8082", "")
	entry3.SetTags("ai", "machine-learning", "trends", "technology")
	entry3.SetCategory("Artificial Intelligence")
	entry3.SetImage("https://localhost:8082/images/ai-trends-2025.jpg")

	entry4 := beam.NewEntry("css-in-2025", "Modern CSS Techniques in 2025", "https://localhost:8082/css-in-2025", time.Now().Add(-36*time.Hour))
	entry4.SetContent("<p>CSS has come a long way. Learn about container queries, new color spaces, and advanced layout techniques available in 2025.</p>")
	entry4.SetSummary("A look at the latest CSS features and how to use them")
	entry4.SetAuthor("John Doe", "john@localhost:8082", "")
	entry4.SetTags("css", "frontend", "web-design", "styles")
	entry4.SetCategory("Web Design")
	entry4.SetImage("https://localhost:8082/images/css-2025.jpg")

	entry5 := beam.NewEntry("cloud-native-security", "Cloud-Native Security Essentials", "https://localhost:8082/cloud-native-security", time.Now().Add(-48*time.Hour))
	entry5.SetContent("<p>Security is critical in cloud-native environments. This article covers best practices for securing containers, orchestrators, and cloud workloads.</p>")
	entry5.SetSummary("How to secure your cloud-native applications in 2025")
	entry5.SetAuthor("Jane Smith", "jane@localhost:8082", "")
	entry5.SetTags("cloud", "security", "devops", "containers")
	entry5.SetCategory("Cloud & DevOps")
	entry5.SetImage("https://localhost:8082/images/cloud-security.jpg")

	feed.AddEntry(entry3)
	feed.AddEntry(entry4)
	feed.AddEntry(entry5)

	if err := feed.Validate(); err != nil {
		log.Fatalf("Feed validation failed: %v", err)
	}

	fmt.Printf("Created feed with %d entries\n", len(feed.Items))
	fmt.Printf("Feed title: %s\n", feed.Title)
	fmt.Printf("Feed URL: %s\n", feed.FeedURL)

	fmt.Println("\n=== Starting HTTP Server ===")
	fmt.Println("Feed available at: http://localhost:8082/feed.json")
	fmt.Println("Press Ctrl+C to stop the server")

	http.Handle("/feed.json", feed)
	http.HandleFunc("/", feed.HomePage)

	http.ListenAndServe(":8082", nil)
}
