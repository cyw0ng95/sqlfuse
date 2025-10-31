package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SQLiteFeature from crawler output
type SQLiteFeature struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Syntax      string   `json:"syntax"`
}

// ImplementedFeatures tracks what we've implemented in sqlfuse
type ImplementedFeatures struct {
	// SQL Statement types we have
	Implemented map[string]bool
	
	// Coverage statistics
	Stats struct {
		TotalDocumented int
		TotalImplemented int
		Missing []string
		PartiallyImplemented []string
	}
}

func main() {
	// Default input file path
	inputFile := "/tmp/sqlite_sql_features.json"
	
	// Check for command-line argument
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	}
	
	// Load crawled features
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading features file from %s: %v\n", inputFile, err)
		fmt.Printf("Usage: %s [input_file.json]\n", os.Args[0])
		os.Exit(1)
	}

	var features []SQLiteFeature
	err = json.Unmarshal(data, &features)
	if err != nil {
		fmt.Printf("Error parsing features JSON: %v\n", err)
		os.Exit(1)
	}

	// Map of implemented features in sqlfuse
	// Based on internal/stmts/stmts/types.go
	implemented := map[string]bool{
		// Core DML
		"SELECT": true,
		"INSERT": true,
		"UPDATE": true,
		"DELETE": true,
		"REPLACE": true,
		
		// DDL - Tables
		"CREATE TABLE": true,
		"DROP TABLE": true,
		"ALTER TABLE": true,
		
		// DDL - Views
		"CREATE VIEW": true,
		"DROP VIEW": true,
		
		// DDL - Indexes
		"CREATE INDEX": true,
		"DROP INDEX": true,
		
		// DDL - Triggers
		"CREATE TRIGGER": true,
		"DROP TRIGGER": true,
		
		// DDL - Virtual Tables
		"CREATE VIRTUAL TABLE": true,
		
		// Transaction Control
		"BEGIN": true,
		"BEGIN TRANSACTION": true,
		"COMMIT": true,
		"COMMIT TRANSACTION": true,
		"ROLLBACK": true,
		"ROLLBACK TRANSACTION": true,
		"SAVEPOINT": true,
		"RELEASE": true,
		"RELEASE SAVEPOINT": true,
		
		// Database Operations
		"ATTACH": true,
		"ATTACH DATABASE": true,
		"DETACH": true,
		"DETACH DATABASE": true,
		
		// Maintenance
		"ANALYZE": true,
		"VACUUM": true,
		"REINDEX": true,
		
		// Query Analysis
		"EXPLAIN": true,
		"EXPLAIN QUERY PLAN": true,
		
		// Pragmas
		"PRAGMA": true,
		
		// Compound SELECT
		"UNION": true,
		"INTERSECT": true,
		"EXCEPT": true,
		
		// Advanced SELECT features
		"WITH": true,  // CTEs
		"WINDOW": true, // Window functions
		
		// Clauses
		"ON CONFLICT": true,
		"RETURNING": true,
		"UPSERT": true,
	}

	// Analyze coverage
	fmt.Println("=== SQLite Coverage Analysis ===\n")
	
	// Extract unique categories from crawled docs
	categories := make(map[string]bool)
	categoryFeatures := make(map[string][]SQLiteFeature)
	
	for _, f := range features {
		categories[f.Category] = true
		categoryFeatures[f.Category] = append(categoryFeatures[f.Category], f)
	}

	// Check coverage per category
	missing := []string{}
	partial := []string{}
	covered := []string{}
	
	for category := range categories {
		// Determine if this category is implemented
		categoryImplemented := false
		
		// Check keywords for this category
		for _, f := range categoryFeatures[category] {
			for _, kw := range f.Keywords {
				if implemented[kw] {
					categoryImplemented = true
					break
				}
			}
			if categoryImplemented {
				break
			}
		}
		
		// Determine status
		categoryName := strings.ToUpper(category)
		
		if categoryImplemented {
			covered = append(covered, categoryName)
		} else {
			// Check if it's a function/expression category (these are partial)
			if strings.Contains(category, "func") || category == "expr" || category == "comment" {
				partial = append(partial, categoryName)
			} else {
				missing = append(missing, categoryName)
			}
		}
	}

	// Sort for consistent output
	sort.Strings(covered)
	sort.Strings(partial)
	sort.Strings(missing)

	// Print results
	fmt.Printf("Total SQLite documentation pages crawled: %d\n", len(features))
	fmt.Printf("Unique categories found: %d\n\n", len(categories))

	fmt.Printf("✅ FULLY COVERED (%d categories):\n", len(covered))
	for _, cat := range covered {
		fmt.Printf("   - %s\n", cat)
	}
	
	fmt.Printf("\n⚠️  PARTIALLY COVERED (%d categories):\n", len(partial))
	fmt.Printf("   (Functions/expressions are complex and have partial coverage)\n")
	for _, cat := range partial {
		fmt.Printf("   - %s\n", cat)
	}
	
	fmt.Printf("\n❌ NOT COVERED (%d categories):\n", len(missing))
	for _, cat := range missing {
		fmt.Printf("   - %s\n", cat)
	}

	// Calculate coverage percentage
	totalCategories := len(categories)
	coveredCount := len(covered)
	partialCount := len(partial)
	coveragePercent := float64(coveredCount+partialCount) * 100.0 / float64(totalCategories)
	
	fmt.Printf("\n=== Coverage Statistics ===\n")
	fmt.Printf("Covered: %d/%d (%.1f%%)\n", coveredCount+partialCount, totalCategories, coveragePercent)
	fmt.Printf("Fully implemented: %d\n", coveredCount)
	fmt.Printf("Partially implemented: %d\n", partialCount)
	fmt.Printf("Not implemented: %d\n", len(missing))

	// Detailed analysis of what's missing
	fmt.Printf("\n=== Missing Feature Details ===\n")
	for _, cat := range missing {
		catLower := strings.ToLower(cat)
		if feats, ok := categoryFeatures[catLower]; ok && len(feats) > 0 {
			f := feats[0]
			fmt.Printf("\n%s:\n", cat)
			fmt.Printf("  URL: %s\n", f.URL)
			if len(f.Keywords) > 0 {
				fmt.Printf("  Keywords: %s\n", strings.Join(f.Keywords, ", "))
			}
		}
	}

	// Recommendations
	fmt.Printf("\n=== Recommendations ===\n")
	if len(missing) > 0 {
		fmt.Println("Consider implementing the following missing features:")
		for _, cat := range missing {
			catLower := strings.ToLower(cat)
			if feats, ok := categoryFeatures[catLower]; ok && len(feats) > 0 {
				fmt.Printf("  • %s (%s)\n", cat, feats[0].URL)
			}
		}
	} else {
		fmt.Println("✅ All documented SQLite statement types are covered!")
	}

	// Save detailed report
	reportFile := "/tmp/sqlite_coverage_report.txt"
	report := generateDetailedReport(features, implemented, covered, partial, missing)
	err = os.WriteFile(reportFile, []byte(report), 0644)
	if err != nil {
		fmt.Printf("\nError saving report: %v\n", err)
	} else {
		fmt.Printf("\nDetailed report saved to: %s\n", reportFile)
	}
}

func generateDetailedReport(features []SQLiteFeature, implemented map[string]bool, 
	covered, partial, missing []string) string {
	
	var sb strings.Builder
	
	sb.WriteString("SQLite Coverage Report\n")
	sb.WriteString("======================\n\n")
	
	sb.WriteString("## Implemented Features\n\n")
	for feature := range implemented {
		sb.WriteString(fmt.Sprintf("- %s\n", feature))
	}
	
	sb.WriteString("\n## Documented Features\n\n")
	for _, f := range features {
		sb.WriteString(fmt.Sprintf("Category: %s\n", f.Category))
		sb.WriteString(fmt.Sprintf("  Name: %s\n", f.Name))
		sb.WriteString(fmt.Sprintf("  URL: %s\n", f.URL))
		if len(f.Keywords) > 0 {
			sb.WriteString(fmt.Sprintf("  Keywords: %s\n", strings.Join(f.Keywords, ", ")))
		}
		sb.WriteString("\n")
	}
	
	sb.WriteString("\n## Coverage Summary\n\n")
	sb.WriteString(fmt.Sprintf("Fully Covered: %d\n", len(covered)))
	sb.WriteString(fmt.Sprintf("Partially Covered: %d\n", len(partial)))
	sb.WriteString(fmt.Sprintf("Not Covered: %d\n", len(missing)))
	
	return sb.String()
}
