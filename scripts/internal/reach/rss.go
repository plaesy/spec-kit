// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RSSAdapter handles RSS/Atom feed parsing.
type RSSAdapter struct {
	BaseAdapter
}

func NewRSSAdapter() *RSSAdapter {
	return &RSSAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformRSS,
			backends:     []Backend{"builtin", "feedparser"},
			requiresAuth: false,
		},
	}
}

// Execute fetches and parses RSS feed.
func (r *RSSAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "builtin":
		return r.executeBuiltin(ctx, query, opts)
	case "feedparser":
		return r.executeFeedparser(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for rss", backend)
	}
}

func (r *RSSAdapter) executeFeedparser(ctx context.Context, url string, opts map[string]string) (BackendResult, error) {
	// Use Python feedparser if available
	args := []string{"python3", "-c", `
import feedparser, json, sys
feed = feedparser.parse(sys.argv[1])
result = {
    "title": feed.feed.get("title", ""),
    "description": feed.feed.get("description", ""),
    "entries": []
}
for entry in feed.entries[:20]:
    result["entries"].append({
        "title": entry.get("title", ""),
        "link": entry.get("link", ""),
        "published": entry.get("published", ""),
        "summary": entry.get("summary", ""),
    })
print(json.dumps(result))
`, url}
	return r.runCommand(ctx, "feedparser", args...)
}

func (r *RSSAdapter) executeBuiltin(ctx context.Context, url string, opts map[string]string) (BackendResult, error) {
	// Fetch the feed
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return BackendResult{
			Backend:   "builtin",
			Platform:  PlatformRSS,
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; plaesy-reach/1.0)")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return BackendResult{
			Backend:   "builtin",
			Platform:  PlatformRSS,
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return BackendResult{
			Backend:   "builtin",
			Platform:  PlatformRSS,
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Try parsing as RSS 2.0
	feed, err := parseRSS2(body)
	if err != nil {
		// Try parsing as Atom
		feed, err = parseAtom(body)
		if err != nil {
			return BackendResult{
				Backend:   "builtin",
				Platform:  PlatformRSS,
				Success:   false,
				Error:     fmt.Sprintf("failed to parse feed: %v", err),
				Timestamp: time.Now(),
			}, nil
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Title: %s\n", feed.Title))
	b.WriteString(fmt.Sprintf("Description: %s\n\n", feed.Description))

	for i, item := range feed.Items {
		if i >= 20 {
			break
		}
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
		b.WriteString(fmt.Sprintf("   Link: %s\n", item.Link))
		if item.Published != "" {
			b.WriteString(fmt.Sprintf("   Published: %s\n", item.Published))
		}
		if item.Description != "" {
			desc := item.Description
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}
			b.WriteString(fmt.Sprintf("   Summary: %s\n", desc))
		}
		b.WriteString("\n")
	}

	return BackendResult{
		Backend:   "builtin",
		Platform:  PlatformRSS,
		Success:   true,
		Data:      b.String(),
		Timestamp: time.Now(),
	}, nil
}

// RSS2Feed represents an RSS 2.0 feed.
type RSS2Feed struct {
	Title       string     `xml:"channel>title"`
	Description string     `xml:"channel>description"`
	Items       []RSS2Item `xml:"channel>item"`
}

// RSS2Item represents an RSS 2.0 item.
type RSS2Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// AtomFeed represents an Atom feed.
type AtomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"http://www.w3.org/2005/Atom title"`
	Entries []AtomEntry `xml:"http://www.w3.org/2005/Atom entry"`
}

// AtomEntry represents an Atom entry.
type AtomEntry struct {
	Title   string   `xml:"http://www.w3.org/2005/Atom title"`
	Link    AtomLink `xml:"http://www.w3.org/2005/Atom link"`
	Summary string   `xml:"http://www.w3.org/2005/Atom summary"`
	Content string   `xml:"http://www.w3.org/2005/Atom content"`
	Updated string   `xml:"http://www.w3.org/2005/Atom updated"`
	ID      string   `xml:"http://www.w3.org/2005/Atom id"`
}

// AtomLink represents an Atom link.
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// ParsedFeed is a normalized feed representation.
type ParsedFeed struct {
	Title       string
	Description string
	Items       []ParsedItem
}

// ParsedItem is a normalized feed item.
type ParsedItem struct {
	Title       string
	Link        string
	Description string
	Published   string
}

func parseRSS2(data []byte) (*ParsedFeed, error) {
	var feed RSS2Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	if feed.Title == "" && len(feed.Items) == 0 {
		return nil, fmt.Errorf("not a valid RSS 2.0 feed")
	}

	parsed := &ParsedFeed{
		Title:       feed.Title,
		Description: feed.Description,
		Items:       make([]ParsedItem, len(feed.Items)),
	}

	for i, item := range feed.Items {
		parsed.Items[i] = ParsedItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Published:   item.PubDate,
		}
	}

	return parsed, nil
}

func parseAtom(data []byte) (*ParsedFeed, error) {
	var feed AtomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	if feed.Title == "" && len(feed.Entries) == 0 {
		return nil, fmt.Errorf("not a valid Atom feed")
	}

	parsed := &ParsedFeed{
		Title: feed.Title,
		Items: make([]ParsedItem, len(feed.Entries)),
	}

	for i, entry := range feed.Entries {
		link := entry.Link.Href
		if entry.Link.Rel == "alternate" && entry.Link.Href != "" {
			link = entry.Link.Href
		}

		description := entry.Summary
		if description == "" {
			description = entry.Content
		}

		published := entry.Updated

		parsed.Items[i] = ParsedItem{
			Title:       entry.Title,
			Link:        link,
			Description: description,
			Published:   published,
		}
	}

	return parsed, nil
}

// Health checks if backends are available.
func (r *RSSAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "builtin":
		return BackendHealth{
			Backend:     "builtin",
			Platform:    PlatformRSS,
			Available:   true,
			LastChecked: time.Now(),
			Version:     "built-in",
		}, nil
	case "feedparser":
		return r.checkCommand(ctx, backend, "python3", "-c", "import feedparser"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}
