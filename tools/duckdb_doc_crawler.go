package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// SQLFeature represents a documented SQL feature in DuckDB
type SQLFeature struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

// DocCrawler crawls DuckDB documentation
type DocCrawler struct {
	baseURL      string
	visited      map[string]bool
	features     []SQLFeature
	maxDepth     int
	currentDepth int
	client       *http.Client
}

// NewDocCrawler creates a new documentation crawler
func NewDocCrawler(baseURL string, maxDepth int) *DocCrawler {
	return &DocCrawler{
		baseURL:  baseURL,
		visited:  make(map[string]bool),
		features: []SQLFeature{},
		maxDepth: maxDepth,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Crawl recursively crawls documentation starting from the base URL
func (dc *DocCrawler) Crawl(url string, depth int) error {
	if depth > dc.maxDepth {
		return nil
	}

	if dc.visited[url] {
		return nil
	}
	dc.visited[url] = true

	fmt.Printf("Crawling [depth=%d]: %s\n", depth, url)

	// Fetch page content
	resp, err := dc.client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non-OK status for %s: %d\n", url, resp.StatusCode)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	content := string(body)

	// Extract SQL features from the page
	dc.extractFeatures(url, content)

	// Extract links to other documentation pages
	links := dc.extractDocLinks(url, content)

	// Recursively crawl linked pages
	for _, link := range links {
		time.Sleep(100 * time.Millisecond) // Be nice to the server
		dc.Crawl(link, depth+1)
	}

	return nil
}

// extractFeatures extracts SQL features from page content
func (dc *DocCrawler) extractFeatures(url string, content string) {
	// Extract category from URL path
	category := dc.extractCategory(url)

	// Look for SQL statement patterns
	// Common patterns: CREATE, ALTER, DROP, SELECT, INSERT, UPDATE, DELETE, etc.
	sqlKeywords := []string{
		"CREATE", "ALTER", "DROP", "SELECT", "INSERT", "UPDATE", "DELETE",
		"COPY", "SET", "RESET", "SHOW", "DESCRIBE", "SUMMARIZE", "EXPLAIN",
		"WITH", "UNION", "INTERSECT", "EXCEPT", "WINDOW", "PARTITION",
		"PIVOT", "UNPIVOT", "QUALIFY", "SAMPLE", "USING SAMPLE",
		"ATTACH", "DETACH", "USE", "CALL", "EXPORT", "IMPORT",
		"PREPARE", "EXECUTE", "DEALLOCATE",
	}

	// Extract headings and code blocks
	headingRe := regexp.MustCompile(`<h[1-4][^>]*>(.*?)</h[1-4]>`)
	codeRe := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)

	headings := headingRe.FindAllStringSubmatch(content, -1)
	codes := codeRe.FindAllStringSubmatch(content, -1)

	for _, heading := range headings {
		headingText := dc.stripHTML(heading[1])
		
		// Check if heading contains SQL keywords
		for _, keyword := range sqlKeywords {
			if strings.Contains(strings.ToUpper(headingText), keyword) {
				feature := SQLFeature{
					Name:        headingText,
					URL:         url,
					Category:    category,
					Description: fmt.Sprintf("DuckDB %s feature", keyword),
					Keywords:    []string{keyword},
				}
				dc.features = append(dc.features, feature)
				break
			}
		}
	}

	// Extract SQL syntax from code blocks
	for _, code := range codes {
		codeText := dc.stripHTML(code[1])
		codeUpper := strings.ToUpper(codeText)
		
		for _, keyword := range sqlKeywords {
			if strings.Contains(codeUpper, keyword) {
				// Check if this is a new feature
				exists := false
				for _, f := range dc.features {
					if f.URL == url && strings.Contains(f.Name, keyword) {
						exists = true
						break
					}
				}
				
				if !exists && len(codeText) < 200 {
					feature := SQLFeature{
						Name:        fmt.Sprintf("%s statement", keyword),
						URL:         url,
						Category:    category,
						Description: codeText,
						Keywords:    []string{keyword},
					}
					dc.features = append(dc.features, feature)
				}
			}
		}
	}
}

// extractDocLinks extracts documentation links from the page
func (dc *DocCrawler) extractDocLinks(currentURL string, content string) []string {
	links := []string{}
	
	// Match links to /docs/stable/sql/ pages
	linkRe := regexp.MustCompile(`href="(/docs/[^"]+)"`)
	matches := linkRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		path := match[1]
		
		// Only follow SQL-related documentation
		if !strings.Contains(path, "/sql/") {
			continue
		}

		// Build full URL
		fullURL := "https://duckdb.org" + path
		
		// Avoid duplicates
		if !dc.visited[fullURL] {
			links = append(links, fullURL)
		}
	}

	return links
}

// extractCategory extracts the category from the URL path
func (dc *DocCrawler) extractCategory(url string) string {
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "sql" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "unknown"
}

// stripHTML removes HTML tags from text
func (dc *DocCrawler) stripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}

// GetFeatures returns all discovered features
func (dc *DocCrawler) GetFeatures() []SQLFeature {
	return dc.features
}

// SaveToFile saves features to a JSON file
func (dc *DocCrawler) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(dc.features, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func main() {
	startURL := "https://duckdb.org/docs/stable/sql/introduction"
	maxDepth := 3 // Limit crawl depth to avoid excessive requests

	crawler := NewDocCrawler(startURL, maxDepth)
	
	fmt.Printf("Starting DuckDB documentation crawl from: %s\n", startURL)
	fmt.Printf("Max depth: %d\n\n", maxDepth)

	err := crawler.Crawl(startURL, 0)
	if err != nil {
		fmt.Printf("Crawl error: %v\n", err)
	}

	features := crawler.GetFeatures()
	fmt.Printf("\n\nTotal features discovered: %d\n", len(features))
	fmt.Printf("Total pages visited: %d\n", len(crawler.visited))

	// Save to file
	outputFile := "/tmp/duckdb_sql_features.json"
	err = crawler.SaveToFile(outputFile)
	if err != nil {
		fmt.Printf("Error saving to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFeatures saved to: %s\n", outputFile)

	// Print summary by category
	categoryCount := make(map[string]int)
	for _, f := range features {
		categoryCount[f.Category]++
	}

	fmt.Println("\nFeatures by category:")
	for cat, count := range categoryCount {
		fmt.Printf("  %s: %d\n", cat, count)
	}

	// Print unique SQL keywords found
	keywordSet := make(map[string]bool)
	for _, f := range features {
		for _, kw := range f.Keywords {
			keywordSet[kw] = true
		}
	}

	fmt.Println("\nUnique SQL keywords found:")
	for kw := range keywordSet {
		fmt.Printf("  %s\n", kw)
	}
}
