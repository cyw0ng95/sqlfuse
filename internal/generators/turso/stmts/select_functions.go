package stmts

import (
	"database/sql"
	"fmt"
	"strings"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
)

// GenSelectWithScalarFunction generates a SELECT statement with scalar SQL functions
func GenSelectWithScalarFunction(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		// No tables available, use literal values
		return genSelectScalarFunctionLiteral(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]
	
	// Select a scalar function to test
	scalarFuncs := []func(*common.LCG, []helper.TableInfo) string{
		genAbsFunction,
		genCharFunction,
		genCoalesceFunction,
		genConcatFunction,
		genConcatWsFunction,
		genHexFunction,
		genIfnullFunction,
		genIifFunction,
		genInstrFunction,
		genLengthFunction,
		genLikeFunction,
		genLowerFunction,
		genUpperFunction,
		genLtrimFunction,
		genRtrimFunction,
		genTrimFunction,
		genMaxMinFunction,
		genNullifFunction,
		genOctetLengthFunction,
		genQuoteFunction,
		genRandomFunction,
		genRandomBlobFunction,
		genReplaceFunction,
		genRoundFunction,
		genSignFunction,
		genSoundexFunction,
		genSubstrFunction,
		genSubstringFunction,
		genTypeofFunction,
		genUnhexFunction,
		genUnicodeFunction,
		genZeroblobFunction,
	}

	funcIdx := rnd(len(scalarFuncs))
	funcExpr := scalarFuncs[funcIdx](lcg, tables)

	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", funcExpr, quoteIdent(tbl.Name), 1+rnd(10))
	return SelectStmt{sql: sql}, nil
}

// genSelectScalarFunctionLiteral generates a SELECT with scalar function using literal values
func genSelectScalarFunctionLiteral(lcg *common.LCG) SelectStmt {
	scalarFuncs := []string{
		"abs(-42)",
		"char(65, 66, 67)",
		"coalesce(NULL, 'default')",
		"concat('Hello', ' ', 'World')",
		"concat_ws(',', 'a', 'b', 'c')",
		"hex('ABC')",
		"ifnull(NULL, 'value')",
		"iif(1 > 0, 'yes', 'no')",
		"instr('Hello World', 'World')",
		"length('Hello')",
		"lower('UPPER')",
		"upper('lower')",
		"ltrim('  spaces  ')",
		"rtrim('  spaces  ')",
		"trim('  spaces  ')",
		"max(1, 2, 3)",
		"min(1, 2, 3)",
		"nullif(1, 2)",
		"octet_length('ABC')",
		"quote('text')",
		"random()",
		"randomblob(10)",
		"replace('Hello World', 'World', 'SQLite')",
		"round(3.14159, 2)",
		"sign(-42)",
		"soundex('Hello')",
		"substr('Hello', 1, 3)",
		"substring('Hello', 2, 2)",
		"typeof(123)",
		"unhex('414243')",
		"unicode('A')",
		"zeroblob(10)",
		"sqlite_version()",
		"sqlite_source_id()",
		"last_insert_rowid()",
	}

	funcIdx := lcg.Intn(len(scalarFuncs))
	sql := fmt.Sprintf("SELECT %s;", scalarFuncs[funcIdx])
	return SelectStmt{sql: sql}
}

// Scalar function generators

func genAbsFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return fmt.Sprintf("abs(%d)", -1-lcg.Intn(100))
	}
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("abs(%s)", col)
	}
	return "abs(-42)"
}

func genCharFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	// Try to use integer columns for char codes if available
	if len(tbls) > 0 && lcg.Intn(3) == 0 {
		col := findNumericColumn(tbls, lcg)
		if col != "" {
			return fmt.Sprintf("char(%s)", col)
		}
	}
	numChars := 1 + lcg.Intn(5)
	chars := make([]string, numChars)
	for i := 0; i < numChars; i++ {
		chars[i] = fmt.Sprintf("%d", 65+lcg.Intn(26)) // A-Z
	}
	return fmt.Sprintf("char(%s)", joinStrings(chars, ", "))
}

func genCoalesceFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "coalesce(NULL, 'default', 'fallback')"
	}
	tbl := tbls[lcg.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "coalesce(NULL, 'default')"
	}
	col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
	defaultVal := types.ValueForType(col.Type, lcg, col.Name)
	return fmt.Sprintf("coalesce(%s, %s)", quoteIdent(col.Name), defaultVal)
}

func genConcatFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	numArgs := 2 + lcg.Intn(3)
	args := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		if len(tbls) > 0 && lcg.Intn(2) == 0 {
			tbl := tbls[lcg.Intn(len(tbls))]
			if len(tbl.Cols) > 0 {
				col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
				args[i] = quoteIdent(col.Name)
				continue
			}
		}
		args[i] = fmt.Sprintf("'part%d'", i)
	}
	return fmt.Sprintf("concat(%s)", joinStrings(args, ", "))
}

func genConcatWsFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	separators := []string{"','", "' '", "'-'", "'|'"}
	sep := separators[lcg.Intn(len(separators))]
	numArgs := 2 + lcg.Intn(3)
	args := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		if len(tbls) > 0 && lcg.Intn(2) == 0 {
			col := findTextColumn(tbls, lcg)
			if col != "" {
				args[i] = col
				continue
			}
		}
		args[i] = fmt.Sprintf("'val%d'", i)
	}
	return fmt.Sprintf("concat_ws(%s, %s)", sep, joinStrings(args, ", "))
}

func genHexFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "hex('ABC123')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("hex(%s)", col)
	}
	return "hex('test')"
}

func genIfnullFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "ifnull(NULL, 'default')"
	}
	tbl := tbls[lcg.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "ifnull(NULL, 'default')"
	}
	col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
	defaultVal := types.ValueForType(col.Type, lcg, col.Name)
	return fmt.Sprintf("ifnull(%s, %s)", quoteIdent(col.Name), defaultVal)
}

func genIifFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	conditions := []string{"1 > 0", "1 = 1", "0 < 1"}
	cond := conditions[lcg.Intn(len(conditions))]
	
	// Try to use columns for the result values when available
	if len(tbls) > 0 && lcg.Intn(2) == 0 {
		col1 := findAnyColumn(tbls, lcg)
		col2 := findAnyColumn(tbls, lcg)
		if col1 != "" && col2 != "" {
			return fmt.Sprintf("iif(%s, %s, %s)", cond, col1, col2)
		}
	}
	return fmt.Sprintf("iif(%s, 'true', 'false')", cond)
}

func genInstrFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "instr('Hello World', 'World')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("instr(%s, 'test')", col)
	}
	return "instr('Hello', 'l')"
}

func genLengthFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "length('Hello')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("length(%s)", col)
	}
	return "length('test')"
}

func genLikeFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "like('Hello', 'H%')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("like(%s, '%%test%%')", col)
	}
	return "like('test', 't%')"
}

func genLowerFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "lower('UPPER')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("lower(%s)", col)
	}
	return "lower('TEST')"
}

func genUpperFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "upper('lower')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("upper(%s)", col)
	}
	return "upper('test')"
}

func genLtrimFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) > 0 && lcg.Intn(3) == 0 {
		col := findTextColumn(tbls, lcg)
		if col != "" {
			if lcg.Intn(2) == 0 {
				return fmt.Sprintf("ltrim(%s)", col)
			}
			return fmt.Sprintf("ltrim(%s, ' ')", col)
		}
	}
	if lcg.Intn(2) == 0 {
		return "ltrim('  spaces  ')"
	}
	return "ltrim('  spaces  ', ' ')"
}

func genRtrimFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) > 0 && lcg.Intn(3) == 0 {
		col := findTextColumn(tbls, lcg)
		if col != "" {
			if lcg.Intn(2) == 0 {
				return fmt.Sprintf("rtrim(%s)", col)
			}
			return fmt.Sprintf("rtrim(%s, ' ')", col)
		}
	}
	if lcg.Intn(2) == 0 {
		return "rtrim('  spaces  ')"
	}
	return "rtrim('  spaces  ', ' ')"
}

func genTrimFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) > 0 && lcg.Intn(3) == 0 {
		col := findTextColumn(tbls, lcg)
		if col != "" {
			if lcg.Intn(2) == 0 {
				return fmt.Sprintf("trim(%s)", col)
			}
			return fmt.Sprintf("trim(%s, ' ')", col)
		}
	}
	if lcg.Intn(2) == 0 {
		return "trim('  spaces  ')"
	}
	return "trim('  spaces  ', ' ')"
}

func genMaxMinFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	fn := "max"
	if lcg.Intn(2) == 0 {
		fn = "min"
	}
	
	// Try to use numeric columns when available
	if len(tbls) > 0 && lcg.Intn(2) == 0 {
		numArgs := 2 + lcg.Intn(3)
		args := make([]string, numArgs)
		col := findNumericColumn(tbls, lcg)
		if col != "" {
			for i := 0; i < numArgs; i++ {
				if lcg.Intn(2) == 0 {
					args[i] = col
				} else {
					args[i] = fmt.Sprintf("%d", lcg.Intn(100))
				}
			}
			return fmt.Sprintf("%s(%s)", fn, joinStrings(args, ", "))
		}
	}
	
	numArgs := 2 + lcg.Intn(3)
	args := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = fmt.Sprintf("%d", lcg.Intn(100))
	}
	return fmt.Sprintf("%s(%s)", fn, joinStrings(args, ", "))
}

func genNullifFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	// Try to use actual columns when available
	if len(tbls) > 0 && lcg.Intn(2) == 0 {
		col := findAnyColumn(tbls, lcg)
		if col != "" {
			val2 := fmt.Sprintf("%d", lcg.Intn(100))
			return fmt.Sprintf("nullif(%s, %s)", col, val2)
		}
	}
	val1 := lcg.Intn(100)
	val2 := lcg.Intn(100)
	return fmt.Sprintf("nullif(%d, %d)", val1, val2)
}

func genOctetLengthFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) > 0 && lcg.Intn(2) == 0 {
		// Try text columns first, then blob columns
		col := findTextColumn(tbls, lcg)
		if col == "" {
			col = findBlobColumn(tbls, lcg)
		}
		if col != "" {
			return fmt.Sprintf("octet_length(%s)", col)
		}
	}
	return "octet_length('ABC')"
}

func genQuoteFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "quote('text')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("quote(%s)", col)
	}
	return "quote('test')"
}

func genRandomFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return "random()"
}

func genRandomBlobFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	size := 1 + lcg.Intn(100)
	return fmt.Sprintf("randomblob(%d)", size)
}

func genReplaceFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "replace('Hello World', 'World', 'SQLite')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("replace(%s, 'test', 'new')", col)
	}
	return "replace('test', 't', 'T')"
}

func genRoundFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if lcg.Intn(2) == 0 {
		return fmt.Sprintf("round(%f)", 3.14159+float64(lcg.Intn(100)))
	}
	digits := lcg.Intn(5)
	return fmt.Sprintf("round(%f, %d)", 3.14159+float64(lcg.Intn(100)), digits)
}

func genSignFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	val := lcg.Intn(200) - 100 // -100 to 99
	return fmt.Sprintf("sign(%d)", val)
}

func genSoundexFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return "soundex('Hello')"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("soundex(%s)", col)
	}
	return "soundex('test')"
}

func genSubstrFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		if lcg.Intn(2) == 0 {
			return "substr('Hello', 1, 3)"
		}
		return "substr('Hello', 2)"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		start := 1 + lcg.Intn(5)
		if lcg.Intn(2) == 0 {
			length := 1 + lcg.Intn(10)
			return fmt.Sprintf("substr(%s, %d, %d)", col, start, length)
		}
		return fmt.Sprintf("substr(%s, %d)", col, start)
	}
	return "substr('test', 1, 2)"
}

func genSubstringFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		if lcg.Intn(2) == 0 {
			return "substring('Hello', 1, 3)"
		}
		return "substring('Hello', 2)"
	}
	col := findTextColumn(tbls, lcg)
	if col != "" {
		start := 1 + lcg.Intn(5)
		if lcg.Intn(2) == 0 {
			length := 1 + lcg.Intn(10)
			return fmt.Sprintf("substring(%s, %d, %d)", col, start, length)
		}
		return fmt.Sprintf("substring(%s, %d)", col, start)
	}
	return "substring('test', 1, 2)"
}

func genTypeofFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		values := []string{"123", "'text'", "3.14", "NULL", "X'0102'"}
		return fmt.Sprintf("typeof(%s)", values[lcg.Intn(len(values))])
	}
	tbl := tbls[lcg.Intn(len(tbls))]
	if len(tbl.Cols) > 0 {
		col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
		return fmt.Sprintf("typeof(%s)", quoteIdent(col.Name))
	}
	return "typeof(123)"
}

func genUnhexFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) > 0 && lcg.Intn(3) == 0 {
		col := findTextColumn(tbls, lcg)
		if col != "" {
			if lcg.Intn(2) == 0 {
				return fmt.Sprintf("unhex(%s)", col)
			}
			return fmt.Sprintf("unhex(%s, ' ')", col)
		}
	}
	if lcg.Intn(2) == 0 {
		return "unhex('414243')"
	}
	return "unhex('414243', ' ')"
}

func genUnicodeFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) > 0 && lcg.Intn(3) == 0 {
		col := findTextColumn(tbls, lcg)
		if col != "" {
			return fmt.Sprintf("unicode(%s)", col)
		}
	}
	chars := []string{"'A'", "'B'", "'Z'", "'0'", "'9'"}
	return fmt.Sprintf("unicode(%s)", chars[lcg.Intn(len(chars))])
}

func genZeroblobFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	size := 1 + lcg.Intn(100)
	return fmt.Sprintf("zeroblob(%d)", size)
}

// GenSelectWithMathFunction generates a SELECT statement with mathematical SQL functions
func GenSelectWithMathFunction(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		return genSelectMathFunctionLiteral(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]
	
	mathFuncs := []func(*common.LCG, []helper.TableInfo) string{
		genAcosFunction,
		genAcoshFunction,
		genAsinFunction,
		genAsinhFunction,
		genAtanFunction,
		genAtan2Function,
		genAtanhFunction,
		genCeilFunction,
		genCeilingFunction,
		genCosFunction,
		genCoshFunction,
		genDegreesFunction,
		genExpFunction,
		genFloorFunction,
		genLnFunction,
		genLogFunction,
		genLog10Function,
		genLog2Function,
		genModFunction,
		genPiFunction,
		genPowFunction,
		genPowerFunction,
		genRadiansFunction,
		genSinFunction,
		genSinhFunction,
		genSqrtFunction,
		genTanFunction,
		genTanhFunction,
		genTruncFunction,
	}

	funcIdx := rnd(len(mathFuncs))
	funcExpr := mathFuncs[funcIdx](lcg, tables)

	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", funcExpr, quoteIdent(tbl.Name), 1+rnd(10))
	return SelectStmt{sql: sql}, nil
}

func genSelectMathFunctionLiteral(lcg *common.LCG) SelectStmt {
	mathFuncs := []string{
		"acos(0.5)",
		"acosh(1.5)",
		"asin(0.5)",
		"asinh(0.5)",
		"atan(1.0)",
		"atan2(1.0, 1.0)",
		"atanh(0.5)",
		"ceil(3.14)",
		"ceiling(3.14)",
		"cos(0.0)",
		"cosh(0.0)",
		"degrees(3.14159)",
		"exp(1.0)",
		"floor(3.14)",
		"ln(2.71828)",
		"log(10, 100)",
		"log(2.71828)",
		"log10(100)",
		"log2(8)",
		"mod(10, 3)",
		"pi()",
		"pow(2, 3)",
		"power(2, 3)",
		"radians(180)",
		"sin(0.0)",
		"sinh(0.0)",
		"sqrt(16)",
		"tan(0.0)",
		"tanh(0.0)",
		"trunc(3.14)",
	}

	funcIdx := lcg.Intn(len(mathFuncs))
	sql := fmt.Sprintf("SELECT %s;", mathFuncs[funcIdx])
	return SelectStmt{sql: sql}
}

// Mathematical function generators

func genAcosFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("acos(%f)", float64(lcg.Intn(100))/100.0)
}

func genAcoshFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("acosh(%f)", 1.0+float64(lcg.Intn(100))/10.0)
}

func genAsinFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("asin(%f)", float64(lcg.Intn(100))/100.0)
}

func genAsinhFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("asinh(%f)", float64(lcg.Intn(100))/10.0)
}

func genAtanFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("atan(%f)", float64(lcg.Intn(100))/10.0)
}

func genAtan2Function(lcg *common.LCG, tbls []helper.TableInfo) string {
	y := float64(lcg.Intn(100)) / 10.0
	x := float64(lcg.Intn(100)) / 10.0
	return fmt.Sprintf("atan2(%f, %f)", y, x)
}

func genAtanhFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("atanh(%f)", float64(lcg.Intn(99))/100.0)
}

func genCeilFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return fmt.Sprintf("ceil(%f)", 3.14+float64(lcg.Intn(100))/10.0)
	}
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("ceil(%s)", col)
	}
	return "ceil(3.14)"
}

func genCeilingFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("ceiling(%f)", 3.14+float64(lcg.Intn(100))/10.0)
}

func genCosFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("cos(%f)", float64(lcg.Intn(360))*3.14159/180.0)
}

func genCoshFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("cosh(%f)", float64(lcg.Intn(100))/10.0)
}

func genDegreesFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("degrees(%f)", float64(lcg.Intn(628))/100.0) // 0 to 2*pi
}

func genExpFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("exp(%f)", float64(lcg.Intn(100))/10.0)
}

func genFloorFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return fmt.Sprintf("floor(%f)", 3.14+float64(lcg.Intn(100))/10.0)
	}
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("floor(%s)", col)
	}
	return "floor(3.14)"
}

func genLnFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("ln(%f)", 1.0+float64(lcg.Intn(100)))
}

func genLogFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if lcg.Intn(2) == 0 {
		// log(X) - natural log
		return fmt.Sprintf("log(%f)", 1.0+float64(lcg.Intn(100)))
	}
	// log(B, X) - log base B of X
	base := 2 + lcg.Intn(8)
	val := 1.0 + float64(lcg.Intn(100))
	return fmt.Sprintf("log(%d, %f)", base, val)
}

func genLog10Function(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("log10(%f)", 1.0+float64(lcg.Intn(100)))
}

func genLog2Function(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("log2(%f)", 1.0+float64(lcg.Intn(100)))
}

func genModFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	x := 1 + lcg.Intn(100)
	y := 1 + lcg.Intn(20)
	return fmt.Sprintf("mod(%d, %d)", x, y)
}

func genPiFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return "pi()"
}

func genPowFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	x := 1 + lcg.Intn(10)
	y := 1 + lcg.Intn(5)
	return fmt.Sprintf("pow(%d, %d)", x, y)
}

func genPowerFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	x := 1 + lcg.Intn(10)
	y := 1 + lcg.Intn(5)
	return fmt.Sprintf("power(%d, %d)", x, y)
}

func genRadiansFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("radians(%d)", lcg.Intn(360))
}

func genSinFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("sin(%f)", float64(lcg.Intn(360))*3.14159/180.0)
}

func genSinhFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("sinh(%f)", float64(lcg.Intn(100))/10.0)
}

func genSqrtFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 || lcg.Intn(2) == 0 {
		return fmt.Sprintf("sqrt(%d)", 1+lcg.Intn(100))
	}
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("sqrt(abs(%s))", col) // Use abs to ensure positive
	}
	return "sqrt(16)"
}

func genTanFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("tan(%f)", float64(lcg.Intn(360))*3.14159/180.0)
}

func genTanhFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("tanh(%f)", float64(lcg.Intn(100))/10.0)
}

func genTruncFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return fmt.Sprintf("trunc(%f)", 3.14+float64(lcg.Intn(100))/10.0)
}

// GenSelectWithAggregateFunction generates a SELECT statement with aggregate SQL functions
func GenSelectWithAggregateFunction(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		return genSelectAggregateFunctionLiteral(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]
	
	aggFuncs := []func(*common.LCG, []helper.TableInfo) string{
		genAvgFunction,
		genCountFunction,
		genCountStarFunction,
		genGroupConcatFunction,
		genStringAggFunction,
		genMaxAggFunction,
		genMinAggFunction,
		genSumFunction,
		genTotalFunction,
	}

	funcIdx := rnd(len(aggFuncs))
	funcExpr := aggFuncs[funcIdx](lcg, tables)

	sql := fmt.Sprintf("SELECT %s FROM %s;", funcExpr, quoteIdent(tbl.Name))
	return SelectStmt{sql: sql}, nil
}

func genSelectAggregateFunctionLiteral(lcg *common.LCG) SelectStmt {
	aggFuncs := []string{
		"count(*)",
		"count(1)",
	}

	funcIdx := lcg.Intn(len(aggFuncs))
	sql := fmt.Sprintf("SELECT %s;", aggFuncs[funcIdx])
	return SelectStmt{sql: sql}
}

// Aggregate function generators

func genAvgFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("avg(%s)", col)
	}
	return "avg(1)"
}

func genCountFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "count(1)"
	}
	tbl := tbls[lcg.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "count(1)"
	}
	col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
	return fmt.Sprintf("count(%s)", quoteIdent(col.Name))
}

func genCountStarFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	return "count(*)"
}

func genGroupConcatFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findTextColumn(tbls, lcg)
	if col == "" {
		return "group_concat('test')"
	}
	if lcg.Intn(2) == 0 {
		return fmt.Sprintf("group_concat(%s)", col)
	}
	separators := []string{",", "|", ";", " "}
	sep := separators[lcg.Intn(len(separators))]
	return fmt.Sprintf("group_concat(%s, '%s')", col, sep)
}

func genStringAggFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findTextColumn(tbls, lcg)
	if col == "" {
		return "string_agg('test', ',')"
	}
	separators := []string{",", "|", ";", " "}
	sep := separators[lcg.Intn(len(separators))]
	return fmt.Sprintf("string_agg(%s, '%s')", col, sep)
}

func genMaxAggFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("max(%s)", col)
	}
	return "max(1)"
}

func genMinAggFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("min(%s)", col)
	}
	return "min(1)"
}

func genSumFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("sum(%s)", col)
	}
	return "sum(1)"
}

func genTotalFunction(lcg *common.LCG, tbls []helper.TableInfo) string {
	col := findNumericColumn(tbls, lcg)
	if col != "" {
		return fmt.Sprintf("total(%s)", col)
	}
	return "total(1)"
}

// GenSelectWithDateTimeFunction generates a SELECT statement with date/time SQL functions
func GenSelectWithDateTimeFunction(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	datetimeFuncs := []func(*common.LCG) string{
		genDateFunction,
		genTimeFunction,
		genDatetimeFunction,
		genJuliandayFunction,
		genUnixepochFunction,
		genStrftimeFunction,
		genTimediffFunction,
	}

	rnd := lcg.Intn
	funcIdx := rnd(len(datetimeFuncs))
	funcExpr := datetimeFuncs[funcIdx](lcg)

	sql := fmt.Sprintf("SELECT %s;", funcExpr)
	return SelectStmt{sql: sql}, nil
}

// Date/time function generators

func genDateFunction(lcg *common.LCG) string {
	if lcg.Intn(3) == 0 {
		return "date('now')"
	}
	if lcg.Intn(2) == 0 {
		return "date('now', '+1 day')"
	}
	return "date('2024-01-01')"
}

func genTimeFunction(lcg *common.LCG) string {
	if lcg.Intn(3) == 0 {
		return "time('now')"
	}
	if lcg.Intn(2) == 0 {
		return "time('now', '+1 hour')"
	}
	return "time('12:00:00')"
}

func genDatetimeFunction(lcg *common.LCG) string {
	if lcg.Intn(3) == 0 {
		return "datetime('now')"
	}
	if lcg.Intn(2) == 0 {
		return "datetime('now', '+1 day')"
	}
	return "datetime('2024-01-01 12:00:00')"
}

func genJuliandayFunction(lcg *common.LCG) string {
	if lcg.Intn(2) == 0 {
		return "julianday('now')"
	}
	return "julianday('2024-01-01')"
}

func genUnixepochFunction(lcg *common.LCG) string {
	if lcg.Intn(2) == 0 {
		return "unixepoch('now')"
	}
	return "unixepoch('2024-01-01')"
}

func genStrftimeFunction(lcg *common.LCG) string {
	formats := []string{"%Y-%m-%d", "%H:%M:%S", "%Y-%m-%d %H:%M:%S", "%s", "%j"}
	format := formats[lcg.Intn(len(formats))]
	if lcg.Intn(2) == 0 {
		return fmt.Sprintf("strftime('%s', 'now')", format)
	}
	return fmt.Sprintf("strftime('%s', '2024-01-01')", format)
}

func genTimediffFunction(lcg *common.LCG) string {
	return "timediff('2024-01-02', '2024-01-01')"
}

// GenSelectWithJSONFunction generates a SELECT statement with JSON SQL functions
func GenSelectWithJSONFunction(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	jsonFuncs := []func(*common.LCG) string{
		genJSONFunction,
		genJSONBFunction,
		genJSONArrayFunction,
		genJSONBArrayFunction,
		genJSONArrayLengthFunction,
		genJSONExtractFunction,
		genJSONBExtractFunction,
		genJSONInsertFunction,
		genJSONObjectFunction,
		genJSONBObjectFunction,
		genJSONPatchFunction,
		genJSONPrettyFunction,
		genJSONRemoveFunction,
		genJSONReplaceFunction,
		genJSONSetFunction,
		genJSONTypeFunction,
		genJSONValidFunction,
		genJSONQuoteFunction,
	}

	rnd := lcg.Intn
	funcIdx := rnd(len(jsonFuncs))
	funcExpr := jsonFuncs[funcIdx](lcg)

	sql := fmt.Sprintf("SELECT %s;", funcExpr)
	return SelectStmt{sql: sql}, nil
}

// JSON function generators

func genJSONFunction(lcg *common.LCG) string {
	jsonObjects := []string{
		"'{\"name\":\"John\",\"age\":30}'",
		"'{\"id\":1,\"value\":\"test\"}'",
		"'[1,2,3]'",
		"'{}'",
	}
	return fmt.Sprintf("json(%s)", jsonObjects[lcg.Intn(len(jsonObjects))])
}

func genJSONBFunction(lcg *common.LCG) string {
	jsonObjects := []string{
		"'{\"name\":\"John\",\"age\":30}'",
		"'{\"id\":1,\"value\":\"test\"}'",
		"'[1,2,3]'",
	}
	return fmt.Sprintf("jsonb(%s)", jsonObjects[lcg.Intn(len(jsonObjects))])
}

func genJSONArrayFunction(lcg *common.LCG) string {
	numValues := 1 + lcg.Intn(4)
	values := make([]string, numValues)
	for i := 0; i < numValues; i++ {
		if lcg.Intn(2) == 0 {
			values[i] = fmt.Sprintf("%d", lcg.Intn(100))
		} else {
			values[i] = fmt.Sprintf("'value%d'", i)
		}
	}
	return fmt.Sprintf("json_array(%s)", joinStrings(values, ", "))
}

func genJSONBArrayFunction(lcg *common.LCG) string {
	numValues := 1 + lcg.Intn(4)
	values := make([]string, numValues)
	for i := 0; i < numValues; i++ {
		if lcg.Intn(2) == 0 {
			values[i] = fmt.Sprintf("%d", lcg.Intn(100))
		} else {
			values[i] = fmt.Sprintf("'value%d'", i)
		}
	}
	return fmt.Sprintf("jsonb_array(%s)", joinStrings(values, ", "))
}

func genJSONArrayLengthFunction(lcg *common.LCG) string {
	if lcg.Intn(2) == 0 {
		return "json_array_length('[1,2,3,4,5]')"
	}
	return "json_array_length('{\"items\":[1,2,3]}', '$.items')"
}

func genJSONExtractFunction(lcg *common.LCG) string {
	jsonObj := "'{\"name\":\"John\",\"age\":30,\"city\":\"NYC\"}'"
	paths := []string{"'$.name'", "'$.age'", "'$.city'"}
	path := paths[lcg.Intn(len(paths))]
	return fmt.Sprintf("json_extract(%s, %s)", jsonObj, path)
}

func genJSONBExtractFunction(lcg *common.LCG) string {
	jsonObj := "'{\"name\":\"John\",\"age\":30}'"
	return fmt.Sprintf("jsonb_extract(%s, '$.name')", jsonObj)
}

func genJSONInsertFunction(lcg *common.LCG) string {
	return "json_insert('{\"a\":1}', '$.b', 2)"
}

func genJSONObjectFunction(lcg *common.LCG) string {
	numPairs := 1 + lcg.Intn(3)
	args := make([]string, 0, numPairs*2)
	for i := 0; i < numPairs; i++ {
		args = append(args, fmt.Sprintf("'key%d'", i))
		if lcg.Intn(2) == 0 {
			args = append(args, fmt.Sprintf("%d", lcg.Intn(100)))
		} else {
			args = append(args, fmt.Sprintf("'value%d'", i))
		}
	}
	return fmt.Sprintf("json_object(%s)", joinStrings(args, ", "))
}

func genJSONBObjectFunction(lcg *common.LCG) string {
	return "jsonb_object('name', 'John', 'age', 30)"
}

func genJSONPatchFunction(lcg *common.LCG) string {
	return "json_patch('{\"a\":1}', '{\"b\":2}')"
}

func genJSONPrettyFunction(lcg *common.LCG) string {
	return "json_pretty('{\"name\":\"John\",\"age\":30}')"
}

func genJSONRemoveFunction(lcg *common.LCG) string {
	return "json_remove('{\"a\":1,\"b\":2,\"c\":3}', '$.b')"
}

func genJSONReplaceFunction(lcg *common.LCG) string {
	return "json_replace('{\"a\":1,\"b\":2}', '$.b', 3)"
}

func genJSONSetFunction(lcg *common.LCG) string {
	return "json_set('{\"a\":1}', '$.b', 2)"
}

func genJSONTypeFunction(lcg *common.LCG) string {
	if lcg.Intn(2) == 0 {
		values := []string{"'null'", "'123'", "'\"text\"'", "'[1,2,3]'", "'{\"a\":1}'"}
		return fmt.Sprintf("json_type(%s)", values[lcg.Intn(len(values))])
	}
	return "json_type('{\"a\":1,\"b\":[1,2]}', '$.b')"
}

func genJSONValidFunction(lcg *common.LCG) string {
	if lcg.Intn(2) == 0 {
		return "json_valid('{\"valid\":true}')"
	}
	return "json_valid('not valid json')"
}

func genJSONQuoteFunction(lcg *common.LCG) string {
	values := []string{"'text'", "'123'", "'true'", "'null'"}
	return fmt.Sprintf("json_quote(%s)", values[lcg.Intn(len(values))])
}

// Helper functions

func findNumericColumn(tbls []helper.TableInfo, lcg *common.LCG) string {
	var numericCols []string
	for _, tbl := range tbls {
		for _, col := range tbl.Cols {
			if isNumericType(col.Type) {
				numericCols = append(numericCols, quoteIdent(col.Name))
			}
		}
	}
	if len(numericCols) == 0 {
		return ""
	}
	return numericCols[lcg.Intn(len(numericCols))]
}

func findTextColumn(tbls []helper.TableInfo, lcg *common.LCG) string {
	var textCols []string
	for _, tbl := range tbls {
		for _, col := range tbl.Cols {
			if isTextType(col.Type) {
				textCols = append(textCols, quoteIdent(col.Name))
			}
		}
	}
	if len(textCols) == 0 {
		return ""
	}
	return textCols[lcg.Intn(len(textCols))]
}

func findBlobColumn(tbls []helper.TableInfo, lcg *common.LCG) string {
	var blobCols []string
	for _, tbl := range tbls {
		for _, col := range tbl.Cols {
			if isBlobType(col.Type) {
				blobCols = append(blobCols, quoteIdent(col.Name))
			}
		}
	}
	if len(blobCols) == 0 {
		return ""
	}
	return blobCols[lcg.Intn(len(blobCols))]
}

func findAnyColumn(tbls []helper.TableInfo, lcg *common.LCG) string {
	var allCols []string
	for _, tbl := range tbls {
		for _, col := range tbl.Cols {
			allCols = append(allCols, quoteIdent(col.Name))
		}
	}
	if len(allCols) == 0 {
		return ""
	}
	return allCols[lcg.Intn(len(allCols))]
}

func isBlobType(t string) bool {
	if t == "" {
		return false
	}
	up := strings.ToUpper(t)
	return strings.Contains(up, "BLOB")
}
