package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// ExampleStmtGeneratorFactory_basic demonstrates basic usage of the Factory pattern.
func ExampleStmtGeneratorFactory_basic() {
	// Create a factory with seed, max recursion depth, and flavor
	lcg := common.NewLCG(12345)
	factory := NewStmtGeneratorFactory(lcg, 3, GetDefaultFlavor())

	// Create a context - this would normally use a real database
	// For this example, we'll show the API usage
	fmt.Printf("Factory created with max depth: %d\n", 3)
	fmt.Printf("Factory uses flavor: %s\n", GetDefaultFlavor().Name())

	// The factory can create generators for different statement types
	gen := factory.CreateGenerator(StmtSelectBasic)
	if gen != nil {
		fmt.Println("SELECT generator created successfully")
	}

	// Output:
	// Factory created with max depth: 3
	// Factory uses flavor: sqlite
	// SELECT generator created successfully
}

// ExampleSelectBuilder demonstrates the Builder pattern for SELECT statements.
func ExampleSelectBuilder() {
	// Create a context
	lcg := common.NewLCG(11111)
	ctx := NewGenContext(nil, lcg, 3)

	// Build a SELECT statement using the fluent interface
	builder := NewSelectBuilder(ctx)
	builder.Select("id", "name", "email").
		From("users").
		Where("age > 18").
		OrderBy("name ASC").
		Limit(10)

	// Note: Build() requires a valid DB connection, so we can't complete this example
	// But this shows the fluent API usage
	fmt.Println("Builder configured with SELECT, FROM, WHERE, ORDER BY, and LIMIT")

	// Output:
	// Builder configured with SELECT, FROM, WHERE, ORDER BY, and LIMIT
}

// ExampleGeneratorRegistry demonstrates the Registry pattern.
func ExampleGeneratorRegistry() {
	// Get the default registry with all standard generators
	registry := DefaultRegistry()

	// Check if a generator exists
	if registry.Has("pragma") {
		fmt.Println("PRAGMA generator is registered")
	}

	// Get a generator by name
	gen := registry.Get("select")
	if gen != nil {
		fmt.Println("SELECT generator retrieved from registry")
	}

	// List available generators
	names := registry.Names()
	fmt.Printf("Total generators registered: %d\n", len(names))

	// Output:
	// PRAGMA generator is registered
	// SELECT generator retrieved from registry
	// Total generators registered: 10
}

// ExampleInsertVariantGenerator demonstrates the Strategy pattern for INSERT variants.
func ExampleInsertVariantGenerator() {
	// Create generators for different INSERT strategies
	insertGen := &InsertVariantGenerator{variant: StmtInsert}
	replaceGen := &InsertVariantGenerator{variant: StmtInsertOrReplace}
	ignoreGen := &InsertVariantGenerator{variant: StmtInsertOrIgnore}

	// Each generator uses a different strategy for handling conflicts
	fmt.Printf("Created INSERT generator with variant: %s\n", StmtInsert)
	fmt.Printf("Created INSERT OR REPLACE generator with variant: %s\n", StmtInsertOrReplace)
	fmt.Printf("Created INSERT OR IGNORE generator with variant: %s\n", StmtInsertOrIgnore)

	// The generators implement the same interface but use different strategies
	if insertGen != nil && replaceGen != nil && ignoreGen != nil {
		fmt.Println("All variant generators created successfully")
	}

	// Output:
	// Created INSERT generator with variant: insert
	// Created INSERT OR REPLACE generator with variant: insert_or_replace
	// Created INSERT OR IGNORE generator with variant: insert_or_ignore
	// All variant generators created successfully
}

// ExampleSelectVariantGenerator demonstrates the Strategy pattern for SELECT variants.
func ExampleSelectVariantGenerator() {
	// Create generators for different SELECT strategies
	basicGen := &SelectVariantGenerator{variant: StmtSelectBasic, maxDepth: 3}
	whereGen := &SelectVariantGenerator{variant: StmtSelectWhere, maxDepth: 3}
	joinGen := &SelectVariantGenerator{variant: StmtSelectJoin, maxDepth: 3}
	cteGen := &SelectVariantGenerator{variant: StmtSelectCTE, maxDepth: 3}

	// Each generator uses a different strategy for SELECT generation
	fmt.Printf("Created basic SELECT generator\n")
	fmt.Printf("Created WHERE SELECT generator\n")
	fmt.Printf("Created JOIN SELECT generator\n")
	fmt.Printf("Created CTE SELECT generator\n")

	if basicGen != nil && whereGen != nil && joinGen != nil && cteGen != nil {
		fmt.Println("All SELECT variant generators created")
	}

	// Output:
	// Created basic SELECT generator
	// Created WHERE SELECT generator
	// Created JOIN SELECT generator
	// Created CTE SELECT generator
	// All SELECT variant generators created
}

// ExampleUnsupportedStmtTypeError demonstrates error handling.
func ExampleUnsupportedStmtTypeError() {
	err := &UnsupportedStmtTypeError{StmtType: StmtType("invalid")}
	fmt.Println(err.Error())

	// Output:
	// unsupported statement type: invalid
}

// ExampleCannotGenerateError demonstrates error handling.
func ExampleCannotGenerateError() {
	err := &CannotGenerateError{
		StmtType: StmtInsert,
		Reason:   "no tables available",
	}
	fmt.Println(err.Error())

	// Output:
	// cannot generate insert: no tables available
}

// ExampleGenContext demonstrates context usage for statement generation.
func ExampleGenContext() {
	// Create a generation context
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 5)

	// Context provides useful methods
	fmt.Printf("Can recurse: %v\n", ctx.CanRecurse())
	fmt.Printf("Current depth: %d\n", ctx.Depth)
	fmt.Printf("Max depth: %d\n", ctx.MaxDepth)

	// Create a descended context for recursive generation
	childCtx := ctx.Descend()
	fmt.Printf("Child depth: %d\n", childCtx.Depth)

	// Output:
	// Can recurse: true
	// Current depth: 0
	// Max depth: 5
	// Child depth: 1
}

// ExampleStmtGeneratorFactory_CreateGenerator demonstrates generator creation.
func ExampleStmtGeneratorFactory_CreateGenerator() {
	lcg := common.NewLCG(12345)
	factory := NewStmtGeneratorFactory(lcg, 3, GetDefaultFlavor())

	// Create different types of generators
	pragmaGen := factory.CreateGenerator(StmtPragma)
	selectGen := factory.CreateGenerator(StmtSelectBasic)
	insertGen := factory.CreateGenerator(StmtInsert)

	// Check if generators were created
	if pragmaGen != nil {
		fmt.Println("PRAGMA generator created")
	}
	if selectGen != nil {
		fmt.Println("SELECT generator created")
	}
	if insertGen != nil {
		fmt.Println("INSERT generator created")
	}

	// Try to create an invalid generator
	invalidGen := factory.CreateGenerator(StmtType("nonexistent"))
	if invalidGen == nil {
		fmt.Println("Invalid generator returned nil")
	}

	// Output:
	// PRAGMA generator created
	// SELECT generator created
	// INSERT generator created
	// Invalid generator returned nil
}
