// Package main demonstrates the SQL flavor design pattern.
// This example shows how different SQL flavors (Turso vs SQLite) 
// generate different SQL based on their capabilities.
package main

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/sqlite/stmts"
	"sqlsmith-go/internal/generators/turso"
)

func main() {
	fmt.Println("=== SQL Flavor Design Pattern Demo ===\n")

	lcg := common.NewLCG(42)

	// Example 1: Default SQLite flavor (supports all features)
	fmt.Println("1. Default SQLite Flavor (supports window functions):")
	sqliteCtx := stmts.NewGenContext(nil, lcg, 2)
	sqliteStmt, _ := stmts.GenSelectWithWindowFunction(nil, lcg)
	fmt.Printf("   Flavor: %s\n", sqliteCtx.Flavor.Name())
	fmt.Printf("   Supports window_functions: %v\n", sqliteCtx.SupportsFeature("window_functions"))
	fmt.Printf("   Generated SQL: %s\n\n", sqliteStmt.SQL())

	// Example 2: Turso flavor (restricted features)
	fmt.Println("2. Turso LibSQL Flavor (does NOT support window functions):")
	tursoFlavor := turso.NewTursoFlavorConfig()
	tursoCtx := stmts.NewGenContextWithFlavor(nil, lcg, 2, tursoFlavor)
	fmt.Printf("   Flavor: %s\n", tursoCtx.Flavor.Name())
	fmt.Printf("   Supports window_functions: %v\n", tursoCtx.SupportsFeature("window_functions"))
	fmt.Printf("   Note: Old API (backward compat) still generates literals.\n")
	fmt.Printf("         New flavor-aware generators would use ROWID instead of OVER.\n\n")

	// Example 3: Recursive CTEs
	fmt.Println("3. Recursive CTE Support Comparison:")
	
	// SQLite supports RECURSIVE
	fmt.Println("   SQLite (supports cte_recursive):")
	sqliteCTEStmt, _ := stmts.GenSelectWithRecursiveCTE(nil, lcg)
	fmt.Printf("   Generated: %s\n", sqliteCTEStmt.SQL())
	
	// Turso does NOT support RECURSIVE
	fmt.Println("   Turso (does NOT support cte_recursive):")
	fmt.Printf("   Supports cte_recursive: %v\n", tursoCtx.SupportsFeature("cte_recursive"))
	fmt.Printf("   (Would generate non-recursive CTE as alternative)\n\n")

	// Example 4: Feature compatibility check
	fmt.Println("4. Feature Compatibility Matrix:")
	features := []string{
		"window_functions",
		"cte_recursive",
		"exists_subquery",
		"in_subquery",
		"regexp",
		"filter_clause",
	}
	
	fmt.Printf("   %-20s %-10s %-10s\n", "Feature", "SQLite", "Turso")
	fmt.Printf("   %-20s %-10s %-10s\n", "-------", "------", "-----")
	for _, feature := range features {
		sqliteSupport := sqliteCtx.SupportsFeature(feature)
		tursoSupport := tursoCtx.SupportsFeature(feature)
		fmt.Printf("   %-20s %-10v %-10v\n", feature, sqliteSupport, tursoSupport)
	}

	fmt.Println("\n=== Key Benefits ===")
	fmt.Println("✓ Single codebase supports multiple SQL flavors")
	fmt.Println("✓ Generators automatically adapt to database capabilities")
	fmt.Println("✓ Prevents generation of unsupported SQL")
	fmt.Println("✓ Easy to add new flavors (PostgreSQL, MySQL, etc.)")
	fmt.Println("✓ Maintains backward compatibility")
}
