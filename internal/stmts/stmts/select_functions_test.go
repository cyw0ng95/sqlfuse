package stmts

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"

	_ "github.com/tursodatabase/turso-go"
)

// TestGenSelectWithScalarFunction tests scalar function SQL generation
func TestGenSelectWithScalarFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(1000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithScalarFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithScalarFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithScalarFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid scalar function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			// Some functions may not be supported, log but don't fail
			t.Logf("Execution failed (may be expected) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithMathFunction tests mathematical function SQL generation
func TestGenSelectWithMathFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithMathFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithMathFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithMathFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid math function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Execution failed (may be expected) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithAggregateFunction tests aggregate function SQL generation
func TestGenSelectWithAggregateFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(3000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithAggregateFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithAggregateFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithAggregateFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid aggregate function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Execution failed (may be expected) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithDateTimeFunction tests date/time function SQL generation
func TestGenSelectWithDateTimeFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(4000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithDateTimeFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithDateTimeFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithDateTimeFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid date/time function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Execution failed (may be expected) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestSpecificScalarFunctions tests specific scalar functions individually
func TestSpecificScalarFunctions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(5000)

	tables, _ := helper.GetAllTablesAndCols(db)
	if tables == nil {
		tables = []helper.TableInfo{}
	}

	tests := []struct {
		name string
		gen  func(*common.LCG, []helper.TableInfo) string
	}{
		{"abs", genAbsFunction},
		{"char", genCharFunction},
		{"coalesce", genCoalesceFunction},
		{"concat", genConcatFunction},
		{"concat_ws", genConcatWsFunction},
		{"glob", genGlobFunction},
		{"hex", genHexFunction},
		{"ifnull", genIfnullFunction},
		{"iif", genIifFunction},
		{"instr", genInstrFunction},
		{"last_insert_rowid", genLastInsertRowidFunction},
		{"length", genLengthFunction},
		{"like", genLikeFunction},
		{"likelihood", genLikelihoodFunction},
		{"likely", genLikelyFunction},
		{"lower", genLowerFunction},
		{"upper", genUpperFunction},
		{"ltrim", genLtrimFunction},
		{"rtrim", genRtrimFunction},
		{"trim", genTrimFunction},
		{"max_min", genMaxMinFunction},
		{"nullif", genNullifFunction},
		{"octet_length", genOctetLengthFunction},
		{"printf", genPrintfFunction},
		{"quote", genQuoteFunction},
		{"random", genRandomFunction},
		{"randomblob", genRandomBlobFunction},
		{"replace", genReplaceFunction},
		{"round", genRoundFunction},
		{"sign", genSignFunction},
		{"soundex", genSoundexFunction},
		{"sqlite_source_id", genSqliteSourceIdFunction},
		{"sqlite_version", genSqliteVersionFunction},
		{"substr", genSubstrFunction},
		{"substring", genSubstringFunction},
		{"typeof", genTypeofFunction},
		{"unhex", genUnhexFunction},
		{"unicode", genUnicodeFunction},
		{"unlikely", genUnlikelyFunction},
		{"zeroblob", genZeroblobFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg, tables)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			sql := fmt.Sprintf("SELECT %s;", funcExpr)
			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			// Try to execute
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Execution of %s failed (may be expected): %v\nSQL: %s", tt.name, err, sql)
			}
			if rows != nil {
				rows.Close()
			}
		})
	}
}

// TestSpecificMathFunctions tests specific mathematical functions individually
func TestSpecificMathFunctions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(6000)

	tables, _ := helper.GetAllTablesAndCols(db)
	if tables == nil {
		tables = []helper.TableInfo{}
	}

	tests := []struct {
		name string
		gen  func(*common.LCG, []helper.TableInfo) string
	}{
		{"acos", genAcosFunction},
		{"acosh", genAcoshFunction},
		{"asin", genAsinFunction},
		{"asinh", genAsinhFunction},
		{"atan", genAtanFunction},
		{"atan2", genAtan2Function},
		{"atanh", genAtanhFunction},
		{"ceil", genCeilFunction},
		{"ceiling", genCeilingFunction},
		{"cos", genCosFunction},
		{"cosh", genCoshFunction},
		{"degrees", genDegreesFunction},
		{"exp", genExpFunction},
		{"floor", genFloorFunction},
		{"ln", genLnFunction},
		{"log", genLogFunction},
		{"log10", genLog10Function},
		{"log2", genLog2Function},
		{"mod", genModFunction},
		{"pi", genPiFunction},
		{"pow", genPowFunction},
		{"power", genPowerFunction},
		{"radians", genRadiansFunction},
		{"sin", genSinFunction},
		{"sinh", genSinhFunction},
		{"sqrt", genSqrtFunction},
		{"tan", genTanFunction},
		{"tanh", genTanhFunction},
		{"trunc", genTruncFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg, tables)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			sql := fmt.Sprintf("SELECT %s;", funcExpr)
			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			// Try to execute
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Execution of %s failed (may be expected): %v\nSQL: %s", tt.name, err, sql)
			}
			if rows != nil {
				rows.Close()
			}
		})
	}
}

// TestSpecificAggregateFunctions tests specific aggregate functions individually
func TestSpecificAggregateFunctions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(7000)

	tables, _ := helper.GetAllTablesAndCols(db)
	if tables == nil {
		tables = []helper.TableInfo{}
	}

	tests := []struct {
		name string
		gen  func(*common.LCG, []helper.TableInfo) string
	}{
		{"avg", genAvgFunction},
		{"count", genCountFunction},
		{"count_star", genCountStarFunction},
		{"group_concat", genGroupConcatFunction},
		{"string_agg", genStringAggFunction},
		{"max", genMaxAggFunction},
		{"min", genMinAggFunction},
		{"sum", genSumFunction},
		{"total", genTotalFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg, tables)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			// Aggregate functions need FROM clause with actual table
			var sql string
			if len(tables) > 0 && len(tables[0].Cols) > 0 {
				sql = fmt.Sprintf("SELECT %s FROM %s;", funcExpr, quoteIdent(tables[0].Name))
			} else {
				sql = fmt.Sprintf("SELECT %s;", funcExpr)
			}

			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			// Try to execute
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Execution of %s failed (may be expected): %v\nSQL: %s", tt.name, err, sql)
			}
			if rows != nil {
				rows.Close()
			}
		})
	}
}

// TestSpecificDateTimeFunctions tests specific date/time functions individually
func TestSpecificDateTimeFunctions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(8000)

	tests := []struct {
		name string
		gen  func(*common.LCG) string
	}{
		{"date", genDateFunction},
		{"time", genTimeFunction},
		{"datetime", genDatetimeFunction},
		{"julianday", genJuliandayFunction},
		{"unixepoch", genUnixepochFunction},
		{"strftime", genStrftimeFunction},
		{"timediff", genTimediffFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			sql := fmt.Sprintf("SELECT %s;", funcExpr)
			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			// Try to execute
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Execution of %s failed (may be expected): %v\nSQL: %s", tt.name, err, sql)
			}
			if rows != nil {
				rows.Close()
			}
		})
	}
}

// TestFunctionDeterminism tests that same seed produces same function SQL
func TestFunctionDeterminism(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		gen  func(*sql.DB, *common.LCG) (SelectStmt, error)
	}{
		{"ScalarFunction", GenSelectWithScalarFunction},
		{"MathFunction", GenSelectWithMathFunction},
		{"AggregateFunction", GenSelectWithAggregateFunction},
		{"DateTimeFunction", GenSelectWithDateTimeFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lcg1 := common.NewLCG(99999)
			stmt1, err1 := tt.gen(db, lcg1)
			if err1 != nil {
				t.Fatalf("First generation failed: %v", err1)
			}

			lcg2 := common.NewLCG(99999)
			stmt2, err2 := tt.gen(db, lcg2)
			if err2 != nil {
				t.Fatalf("Second generation failed: %v", err2)
			}

			if stmt1.SQL() != stmt2.SQL() {
				t.Errorf("Same seed produced different SQL:\n  First:  %s\n  Second: %s",
					stmt1.SQL(), stmt2.SQL())
			}
		})
	}
}

// TestFunctionCoverage ensures we generate different functions across iterations
func TestFunctionCoverage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(10000)

	// Track which functions we've seen
	seenFunctions := make(map[string]bool)

	for i := 0; i < 100; i++ {
		stmt, err := GenSelectWithScalarFunction(db, lcg)
		if err != nil {
			t.Fatalf("Generation failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		// Extract function name (rough heuristic)
		if idx := strings.Index(sql, "("); idx > 0 {
			funcName := sql[strings.LastIndex(sql[:idx], " ")+1 : idx]
			seenFunctions[funcName] = true
		}
	}

	// We should see multiple different functions
	if len(seenFunctions) < 5 {
		t.Errorf("Expected to see at least 5 different functions, got %d: %v",
			len(seenFunctions), seenFunctions)
	}
}

// TestExecuteScalarFunctionsWithRealData tests execution with actual database
func TestExecuteScalarFunctionsWithRealData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test specific functions that should always work
	testCases := []struct {
		name string
		sql  string
	}{
		{"abs", "SELECT abs(-42);"},
		{"length", "SELECT length('Hello');"},
		{"lower", "SELECT lower('UPPER');"},
		{"upper", "SELECT upper('lower');"},
		{"trim", "SELECT trim('  spaces  ');"},
		{"round", "SELECT round(3.14159, 2);"},
		{"typeof", "SELECT typeof(123);"},
		{"quote", "SELECT quote('text');"},
		{"hex", "SELECT hex('ABC');"},
		{"substr", "SELECT substr('Hello', 1, 3);"},
		{"replace", "SELECT replace('Hello', 'l', 'L');"},
		{"coalesce", "SELECT coalesce(NULL, 'default');"},
		{"ifnull", "SELECT ifnull(NULL, 'value');"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(tc.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v\nSQL: %s", tc.name, err, tc.sql)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Errorf("No rows returned for %s", tc.name)
				return
			}

			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tc.name, err)
			}

			t.Logf("%s result: %v", tc.name, result)
		})
	}
}

// TestExecuteMathFunctionsWithRealData tests math function execution
func TestExecuteMathFunctionsWithRealData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name string
		sql  string
	}{
		{"abs", "SELECT abs(-42);"},
		{"ceil", "SELECT ceil(3.14);"},
		{"floor", "SELECT floor(3.14);"},
		{"round", "SELECT round(3.14159);"},
		{"sqrt", "SELECT sqrt(16);"},
		{"power", "SELECT power(2, 3);"},
		{"mod", "SELECT mod(10, 3);"},
		{"pi", "SELECT pi();"},
		{"sin", "SELECT sin(0);"},
		{"cos", "SELECT cos(0);"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(tc.sql)
			if err != nil {
				t.Logf("Math function %s not supported (expected): %v", tc.name, err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Errorf("No rows returned for %s", tc.name)
				return
			}

			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tc.name, err)
			}

			t.Logf("%s result: %v", tc.name, result)
		})
	}
}

// TestExecuteAggregateFunctionsWithRealData tests aggregate function execution
func TestExecuteAggregateFunctionsWithRealData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name string
		sql  string
	}{
		{"count_star", "SELECT count(*) FROM users;"},
		{"count", "SELECT count(id) FROM users;"},
		{"avg", "SELECT avg(age) FROM users;"},
		{"sum", "SELECT sum(age) FROM users;"},
		{"min", "SELECT min(age) FROM users;"},
		{"max", "SELECT max(age) FROM users;"},
		{"total", "SELECT total(age) FROM users;"},
		{"group_concat", "SELECT group_concat(name) FROM users;"},
		{"group_concat_sep", "SELECT group_concat(name, ',') FROM users;"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(tc.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v\nSQL: %s", tc.name, err, tc.sql)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Errorf("No rows returned for %s", tc.name)
				return
			}

			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tc.name, err)
			}

			t.Logf("%s result: %v", tc.name, result)
		})
	}
}

// TestExecuteDateTimeFunctionsWithRealData tests date/time function execution
func TestExecuteDateTimeFunctionsWithRealData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name string
		sql  string
	}{
		{"date_now", "SELECT date('now');"},
		{"date_literal", "SELECT date('2024-01-01');"},
		{"time_now", "SELECT time('now');"},
		{"datetime_now", "SELECT datetime('now');"},
		{"strftime", "SELECT strftime('%Y-%m-%d', 'now');"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(tc.sql)
			if err != nil {
				t.Logf("Date/time function %s not fully supported (may be expected): %v", tc.name, err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Errorf("No rows returned for %s", tc.name)
				return
			}

			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tc.name, err)
			}

			t.Logf("%s result: %v", tc.name, result)
		})
	}
}

// TestGenSelectWithJSONFunction tests JSON function SQL generation
func TestGenSelectWithJSONFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(9000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithJSONFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithJSONFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithJSONFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid JSON function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Execution failed (may be expected) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestSpecificJSONFunctions tests specific JSON functions individually
func TestSpecificJSONFunctions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(10000)

	tests := []struct {
		name string
		gen  func(*common.LCG) string
	}{
		{"json", genJSONFunction},
		{"jsonb", genJSONBFunction},
		{"json_array", genJSONArrayFunction},
		{"jsonb_array", genJSONBArrayFunction},
		{"json_array_length", genJSONArrayLengthFunction},
		{"json_extract", genJSONExtractFunction},
		{"jsonb_extract", genJSONBExtractFunction},
		{"json_insert", genJSONInsertFunction},
		{"json_object", genJSONObjectFunction},
		{"jsonb_object", genJSONBObjectFunction},
		{"json_patch", genJSONPatchFunction},
		{"json_pretty", genJSONPrettyFunction},
		{"json_remove", genJSONRemoveFunction},
		{"json_replace", genJSONReplaceFunction},
		{"json_set", genJSONSetFunction},
		{"json_type", genJSONTypeFunction},
		{"json_valid", genJSONValidFunction},
		{"json_quote", genJSONQuoteFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			sql := fmt.Sprintf("SELECT %s;", funcExpr)
			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			// Try to execute
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Execution of %s failed (may be expected): %v\nSQL: %s", tt.name, err, sql)
			}
			if rows != nil {
				rows.Close()
			}
		})
	}
}

// TestExecuteJSONFunctionsWithRealData tests JSON function execution
func TestExecuteJSONFunctionsWithRealData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name string
		sql  string
	}{
		{"json", "SELECT json('{\"name\":\"John\",\"age\":30}');"},
		{"json_array", "SELECT json_array(1, 2, 3, 'four');"},
		{"json_object", "SELECT json_object('name', 'John', 'age', 30);"},
		{"json_extract", "SELECT json_extract('{\"name\":\"John\",\"age\":30}', '$.name');"},
		{"json_type", "SELECT json_type('{\"a\":1}');"},
		{"json_valid", "SELECT json_valid('{\"valid\":true}');"},
		{"json_quote", "SELECT json_quote('text');"},
		{"json_array_length", "SELECT json_array_length('[1,2,3,4,5]');"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(tc.sql)
			if err != nil {
				t.Logf("JSON function %s not supported (may be expected): %v", tc.name, err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Errorf("No rows returned for %s", tc.name)
				return
			}

			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tc.name, err)
			}

			t.Logf("%s result: %v", tc.name, result)
		})
	}
}

// TestGenSelectWithUUIDFunction tests UUID extension function SQL generation
func TestGenSelectWithUUIDFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(8000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithUUIDFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithUUIDFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithUUIDFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid UUID function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			// UUID functions may not be supported in all SQLite builds, log but don't fail
			t.Logf("Execution failed (may be expected for UUID extension) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithRegexpFunction tests regexp extension function SQL generation
func TestGenSelectWithRegexpFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(9000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithRegexpFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithRegexpFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithRegexpFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid regexp function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			// Regexp functions may not be supported in all SQLite builds, log but don't fail
			t.Logf("Execution failed (may be expected for regexp extension) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithVectorFunction tests vector extension function SQL generation
func TestGenSelectWithVectorFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(10000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithVectorFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithVectorFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithVectorFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid vector function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			// Vector functions may not be supported in all SQLite builds, log but don't fail
			t.Logf("Execution failed (may be expected for vector extension) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithTimeFunction tests time extension function SQL generation
func TestGenSelectWithTimeFunction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(11000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithTimeFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithTimeFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithTimeFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid time function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute the SQL
		rows, err := db.Query(sql)
		if err != nil {
			// Time functions may not be supported in all SQLite builds, log but don't fail
			t.Logf("Execution failed (may be expected for time extension) on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestSpecificExtensionFunctions tests specific extension functions individually
func TestSpecificExtensionFunctions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testCases := []struct {
		name string
		sql  string
	}{
		// UUID functions
		{"uuid4", "SELECT uuid4();"},
		{"uuid4_str", "SELECT uuid4_str();"},
		{"uuid7", "SELECT uuid7();"},
		{"uuid7_with_timestamp", "SELECT uuid7(1609459200);"},
		{"uuid_str", "SELECT uuid_str(uuid4());"},

		// Regexp functions
		{"regexp", "SELECT regexp('[0-9]+', '123abc');"},
		{"regexp_like", "SELECT regexp_like('hello123', '[a-z]+');"},
		{"regexp_substr", "SELECT regexp_substr('hello world', 'w[a-z]+');"},
		{"regexp_capture", "SELECT regexp_capture('test@example.com', '([a-z]+)@([a-z]+\\.[a-z]+)');"},
		{"regexp_replace", "SELECT regexp_replace('hello world', 'world', 'universe');"},

		// Vector functions
		{"vector", "SELECT vector('[1.0,2.0,3.0]');"},
		{"vector32", "SELECT vector32('[1.0,2.0,3.0]');"},
		{"vector64", "SELECT vector64('[1.0,2.0,3.0]');"},
		{"vector_extract", "SELECT vector_extract(vector('[1.0,2.0,3.0]'));"},
		{"vector_distance_cos", "SELECT vector_distance_cos(vector('[1.0,2.0,3.0]'), vector('[4.0,5.0,6.0]'));"},
		{"vector_distance_l2", "SELECT vector_distance_l2(vector('[1.0,2.0,3.0]'), vector('[4.0,5.0,6.0]'));"},

		// Time functions
		{"time_now", "SELECT time_now();"},
		{"time_date", "SELECT time_date(2024, 1, 1);"},
		{"time_unix", "SELECT time_unix(1609459200);"},
		{"time_get_year", "SELECT time_get_year(time_now());"},
		{"time_to_unix", "SELECT time_to_unix(time_now());"},
		{"time_fmt_iso", "SELECT time_fmt_iso(time_now());"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, errors := ValidateSQL(tc.sql)
			if !valid {
				t.Logf("SQL validation failed (may be expected for extensions): %v\nSQL: %s", errors, tc.sql)
			}

			rows, err := db.Query(tc.sql)
			if err != nil {
				// Extension functions may not be supported in standard SQLite, log but don't fail
				t.Logf("Execution failed (expected for extension function %s): %v", tc.name, err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Logf("No rows returned for %s", tc.name)
				return
			}

			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tc.name, err)
			}

			t.Logf("%s result: %v", tc.name, result)
		})
	}
}

// TestGenSelectWithGoSQLite3ScalarFunction tests go-sqlite3 specific core functions
func TestGenSelectWithGoSQLite3ScalarFunction(t *testing.T) {
	// This test uses Turso which doesn't support go-sqlite3 specific functions
	// We just verify that the SQL is generated correctly, not that it executes
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(5000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithGoSQLite3ScalarFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithGoSQLite3ScalarFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithGoSQLite3ScalarFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		// Validate SQL syntax (will pass even if function not supported)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid go-sqlite3 scalar function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Note: We don't try to execute these functions with Turso as they're not supported
		// They should be tested with actual go-sqlite3 database
		t.Logf("Generated SQL (iteration %d): %s", i, sql)
	}
}

// TestGoSQLite3SpecificFunctionGenerators tests individual go-sqlite3 specific function generators
func TestGoSQLite3SpecificFunctionGenerators(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(6000)
	
	// Get tables for testing
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	tests := []struct {
		name string
		gen  func(*common.LCG, []helper.TableInfo) string
	}{
		{"changes", genChangesFunction},
		{"total_changes", genTotalChangesFunction},
		{"format", genFormatFunction},
		{"sqlite_compileoption_get", genSqliteCompileoptionGetFunction},
		{"sqlite_compileoption_used", genSqliteCompileoptionUsedFunction},
		{"sqlite_offset", genSqliteOffsetFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg, tables)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			sql := fmt.Sprintf("SELECT %s;", funcExpr)
			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			t.Logf("Generated %s: %s", tt.name, funcExpr)
		})
	}
}

// TestGenSelectWithGoSQLite3AggregateFunction tests go-sqlite3 specific aggregate functions
func TestGenSelectWithGoSQLite3AggregateFunction(t *testing.T) {
	// This test uses Turso which may not support go-sqlite3 specific JSON aggregate functions
	// We verify that the SQL is generated correctly, not that it executes
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(7000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithGoSQLite3AggregateFunction(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithGoSQLite3AggregateFunction failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithGoSQLite3AggregateFunction returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		// Validate SQL syntax (will pass even if function not supported)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid go-sqlite3 aggregate function SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Note: We don't try to execute these functions with Turso as they may not be supported
		// They should be tested with actual go-sqlite3 database
		t.Logf("Generated SQL (iteration %d): %s", i, sql)
	}
}

// TestGoSQLite3SpecificAggregateFunctionGenerators tests individual go-sqlite3 specific aggregate function generators
func TestGoSQLite3SpecificAggregateFunctionGenerators(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(8000)
	
	// Get tables for testing
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	tests := []struct {
		name string
		gen  func(*common.LCG, []helper.TableInfo) string
	}{
		{"json_group_array", genJSONGroupArrayFunction},
		{"json_group_object", genJSONGroupObjectFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcExpr := tt.gen(lcg, tables)
			if funcExpr == "" {
				t.Error("Function generator returned empty string")
				return
			}

			// Aggregate functions need FROM clause with actual table
			var sql string
			if len(tables) > 0 && len(tables[0].Cols) > 0 {
				sql = fmt.Sprintf("SELECT %s FROM %s;", funcExpr, quoteIdent(tables[0].Name))
			} else {
				sql = fmt.Sprintf("SELECT %s;", funcExpr)
			}

			valid, errors := ValidateSQL(sql)
			if !valid {
				t.Errorf("Invalid SQL for %s: %s\nErrors: %v", tt.name, sql, errors)
			}

			t.Logf("Generated %s: %s", tt.name, funcExpr)
		})
	}
}
