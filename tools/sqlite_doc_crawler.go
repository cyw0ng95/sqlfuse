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

// SQLiteFeature represents a documented SQL feature in SQLite
type SQLiteFeature struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Syntax      string   `json:"syntax"`
}

// SQLiteDocCrawler crawls SQLite documentation
type SQLiteDocCrawler struct {
	baseURL      string
	visited      map[string]bool
	features     []SQLiteFeature
	maxDepth     int
	currentDepth int
	client       *http.Client
}

// NewSQLiteDocCrawler creates a new SQLite documentation crawler
func NewSQLiteDocCrawler(baseURL string, maxDepth int) *SQLiteDocCrawler {
	return &SQLiteDocCrawler{
		baseURL:  baseURL,
		visited:  make(map[string]bool),
		features: []SQLiteFeature{},
		maxDepth: maxDepth,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Crawl recursively crawls documentation starting from the base URL
func (dc *SQLiteDocCrawler) Crawl(url string, depth int) error {
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
		time.Sleep(200 * time.Millisecond) // Be nice to the server
		err := dc.Crawl(link, depth+1)
		if err != nil {
			fmt.Printf("Error crawling %s: %v\n", link, err)
		}
	}

	return nil
}

// extractFeatures extracts SQL features from page content
func (dc *SQLiteDocCrawler) extractFeatures(url string, content string) {
	// Extract category from URL
	category := dc.extractCategory(url)

	// Extract the page title/heading to use as the feature name
	titleRe := regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`)
	titleMatches := titleRe.FindStringSubmatch(content)
	
	pageTitle := ""
	if len(titleMatches) > 1 {
		pageTitle = dc.stripHTML(titleMatches[1])
	}

	// If this is a SQL documentation page (lang_*.html, pragma.html, etc.)
	// Then it documents a specific SQL feature
	if strings.Contains(url, "lang_") || strings.Contains(url, "pragma.html") || 
	   strings.Contains(url, "windowfunctions.html") || strings.Contains(url, "json1.html") {
		
		// Extract syntax diagrams and code examples
		syntax := ""
		codeRe := regexp.MustCompile(`<pre[^>]*>(.*?)</pre>`)
		codes := codeRe.FindAllStringSubmatch(content, -1)
		
		if len(codes) > 0 {
			// Get first code block as syntax example
			syntax = dc.stripHTML(codes[0][1])
			if len(syntax) > 800 {
				syntax = syntax[:800] + "..."
			}
		}

		// Extract keywords from the title and URL
		keywords := dc.extractKeywords(pageTitle, url)
		
		if len(keywords) > 0 || pageTitle != "" {
			feature := SQLiteFeature{
				Name:        pageTitle,
				URL:         url,
				Category:    category,
				Description: fmt.Sprintf("SQLite %s documentation", category),
				Keywords:    keywords,
				Syntax:      syntax,
			}
			dc.features = append(dc.features, feature)
		}
	}
}

// extractKeywords extracts SQL keywords from title and URL
func (dc *SQLiteDocCrawler) extractKeywords(title string, url string) []string {
	keywords := []string{}
	
	titleUpper := strings.ToUpper(title)
	urlUpper := strings.ToUpper(url)
	
	// Common SQL keywords to look for
	keywordPatterns := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "REPLACE",
		"CREATE", "DROP", "ALTER", "ANALYZE", "ATTACH", "DETACH",
		"BEGIN", "COMMIT", "ROLLBACK", "SAVEPOINT", "RELEASE",
		"EXPLAIN", "PRAGMA", "REINDEX", "VACUUM",
		"TABLE", "VIEW", "INDEX", "TRIGGER", "VIRTUAL TABLE",
		"TRANSACTION",
		"UNION", "INTERSECT", "EXCEPT",
		"ON CONFLICT", "UPSERT", "RETURNING",
		"WITH", "WINDOW", "JSON",
	}
	
	for _, kw := range keywordPatterns {
		if strings.Contains(titleUpper, kw) || strings.Contains(urlUpper, strings.ReplaceAll(kw, " ", "")) {
			keywords = append(keywords, kw)
		}
	}
	
	return keywords
}

// looksLikeSQLSyntax checks if text looks like SQL syntax documentation
func (dc *SQLiteDocCrawler) looksLikeSQLSyntax(text string) bool {
	upper := strings.ToUpper(text)
	// Look for common SQL syntax patterns
	patterns := []string{
		"SELECT", "FROM", "WHERE", "CREATE", "INSERT", "UPDATE", "DELETE",
		"TABLE", "INDEX", "VIEW", "TRIGGER", ";",
	}
	
	matchCount := 0
	for _, pattern := range patterns {
		if strings.Contains(upper, pattern) {
			matchCount++
		}
	}
	
	// If it contains at least 2 SQL keywords, it's likely syntax
	return matchCount >= 2
}

// extractDocLinks extracts documentation links from the page
func (dc *SQLiteDocCrawler) extractDocLinks(currentURL string, content string) []string {
	links := []string{}
	
	// Match links to lang*.html and related SQL documentation pages
	// Also include pragma.html, windowfunctions.html, json1.html
	linkRe := regexp.MustCompile(`href=['"]([^'"]*(?:lang_[^'"]*\.html|pragma\.html|windowfunctions\.html|json1\.html|syntaxdiagrams\.html))['"]`)
	matches := linkRe.FindAllStringSubmatch(content, -1)

	baseURL := "https://sqlite.org/"
	
	for _, match := range matches {
		path := match[1]
		
		// Skip fragments
		if strings.Contains(path, "#") {
			path = strings.Split(path, "#")[0]
		}
		
		if path == "" {
			continue
		}
		
		// Build full URL
		fullURL := baseURL + path
		
		// Avoid duplicates
		if !dc.visited[fullURL] {
			links = append(links, fullURL)
		}
	}

	return links
}

// extractCategory extracts the category from the URL or page title
func (dc *SQLiteDocCrawler) extractCategory(url string) string {
	// Extract from filename
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		filename = strings.TrimSuffix(filename, ".html")
		
		// Clean up the category name
		if strings.HasPrefix(filename, "lang_") {
			return strings.TrimPrefix(filename, "lang_")
		}
		return filename
	}
	return "unknown"
}

// stripHTML removes HTML tags from text
func (dc *SQLiteDocCrawler) stripHTML(s string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	
	// Decode common HTML entities
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	
	// Clean up whitespace
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	s = strings.Join(cleanedLines, "\n")
	
	return s
}

// GetFeatures returns all discovered features
func (dc *SQLiteDocCrawler) GetFeatures() []SQLiteFeature {
	return dc.features
}

// SaveToFile saves features to a JSON file
func (dc *SQLiteDocCrawler) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(dc.features, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// PrintSummary prints a summary of discovered features
func (dc *SQLiteDocCrawler) PrintSummary() {
	features := dc.GetFeatures()
	fmt.Printf("\n\n=== SQLite Documentation Crawl Summary ===\n")
	fmt.Printf("Total features discovered: %d\n", len(features))
	fmt.Printf("Total pages visited: %d\n\n", len(dc.visited))

	// Count by category
	categoryCount := make(map[string]int)
	for _, f := range features {
		categoryCount[f.Category]++
	}

	fmt.Println("Features by category:")
	for cat, count := range categoryCount {
		fmt.Printf("  %-30s: %d\n", cat, count)
	}

	// Unique keywords
	keywordSet := make(map[string]bool)
	for _, f := range features {
		for _, kw := range f.Keywords {
			keywordSet[kw] = true
		}
	}

	fmt.Printf("\nUnique SQL keywords found: %d\n", len(keywordSet))
	fmt.Println("Keywords:")
	for kw := range keywordSet {
		fmt.Printf("  %s\n", kw)
	}

	// Print sample features
	fmt.Println("\nSample features discovered:")
	sampleCount := 5
	if len(features) < sampleCount {
		sampleCount = len(features)
	}
	for i := 0; i < sampleCount; i++ {
		f := features[i]
		fmt.Printf("\n  Name: %s\n", f.Name)
		fmt.Printf("  Category: %s\n", f.Category)
		fmt.Printf("  URL: %s\n", f.URL)
		if f.Syntax != "" {
			syntaxPreview := f.Syntax
			if len(syntaxPreview) > 100 {
				syntaxPreview = syntaxPreview[:100] + "..."
			}
			fmt.Printf("  Syntax: %s\n", syntaxPreview)
		}
	}
}

func main() {
	startURL := "https://sqlite.org/lang.html"
	maxDepth := 2 // Limit crawl depth to avoid excessive requests

	crawler := NewSQLiteDocCrawler(startURL, maxDepth)
	
	fmt.Printf("=== SQLite Documentation Crawler ===\n")
	fmt.Printf("Starting URL: %s\n", startURL)
	fmt.Printf("Max depth: %d\n\n", maxDepth)

	err := crawler.Crawl(startURL, 0)
	if err != nil {
		fmt.Printf("Crawl error: %v\n", err)
	}

	// Print summary
	crawler.PrintSummary()

	// Save to file
	outputFile := "/tmp/sqlite_sql_features.json"
	err = crawler.SaveToFile(outputFile)
	if err != nil {
		fmt.Printf("\nError saving to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n\nFeatures saved to: %s\n", outputFile)
}
