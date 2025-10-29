package keywords

// SQLite keyword support based on https://sqlite.org/lang_keywords.html
// This package provides comprehensive keyword detection and classification
// for proper identifier handling and SQL generation.

// KeywordType represents the classification of a SQLite keyword
type KeywordType int

const (
	// Reserved keywords that must always be quoted when used as identifiers
	Reserved KeywordType = iota
	// NonReserved keywords that may need quoting in certain contexts
	NonReserved
	// TursoUnsupported keywords not supported by Turso LibSQL
	TursoUnsupported
)

// Keyword represents a SQLite keyword with its properties
type Keyword struct {
	Name         string
	Type         KeywordType
	TursoSupport bool
}

// allKeywords is the comprehensive list of all SQLite keywords from
// https://sqlite.org/lang_keywords.html (as of SQLite 3.x)
var allKeywords = []Keyword{
	{"ABORT", Reserved, true},
	{"ACTION", Reserved, true},
	{"ADD", Reserved, true},
	{"AFTER", Reserved, true},
	{"ALL", Reserved, true},
	{"ALTER", Reserved, true},
	{"ALWAYS", Reserved, true},
	{"ANALYZE", Reserved, true},
	{"AND", Reserved, true},
	{"AS", Reserved, true},
	{"ASC", Reserved, true},
	{"ATTACH", Reserved, true},
	{"AUTOINCREMENT", Reserved, true},
	{"BEFORE", Reserved, true},
	{"BEGIN", Reserved, true},
	{"BETWEEN", Reserved, true},
	{"BY", Reserved, true},
	{"CASCADE", Reserved, true},
	{"CASE", Reserved, true},
	{"CAST", Reserved, true},
	{"CHECK", Reserved, true},
	{"COLLATE", Reserved, true},
	{"COLUMN", Reserved, true},
	{"COMMIT", Reserved, true},
	{"CONFLICT", Reserved, true},
	{"CONSTRAINT", Reserved, true},
	{"CREATE", Reserved, true},
	{"CROSS", Reserved, true},
	{"CURRENT", Reserved, true},
	{"CURRENT_DATE", Reserved, true},
	{"CURRENT_TIME", Reserved, true},
	{"CURRENT_TIMESTAMP", Reserved, true},
	{"DATABASE", Reserved, true},
	{"DEFAULT", Reserved, true},
	{"DEFERRABLE", Reserved, true},
	{"DEFERRED", Reserved, true},
	{"DELETE", Reserved, true},
	{"DESC", Reserved, true},
	{"DETACH", Reserved, true},
	{"DISTINCT", Reserved, true},
	{"DO", Reserved, true},
	{"DROP", Reserved, true},
	{"EACH", Reserved, true},
	{"ELSE", Reserved, true},
	{"END", Reserved, true},
	{"ESCAPE", Reserved, true},
	{"EXCEPT", Reserved, true},
	{"EXCLUDE", Reserved, true},
	{"EXCLUSIVE", Reserved, true},
	{"EXISTS", Reserved, true},
	{"EXPLAIN", Reserved, true},
	{"FAIL", Reserved, true},
	{"FILTER", Reserved, false}, // Turso doesn't support FILTER clause
	{"FIRST", Reserved, true},
	{"FOLLOWING", Reserved, true},
	{"FOR", Reserved, true},
	{"FOREIGN", Reserved, true},
	{"FROM", Reserved, true},
	{"FULL", Reserved, true},
	{"GENERATED", Reserved, true},
	{"GLOB", Reserved, true},
	{"GROUP", Reserved, true},
	{"GROUPS", Reserved, true},
	{"HAVING", Reserved, true},
	{"IF", Reserved, true},
	{"IGNORE", Reserved, true},
	{"IMMEDIATE", Reserved, true},
	{"IN", Reserved, true},
	{"INDEX", Reserved, true},
	{"INDEXED", Reserved, true},
	{"INITIALLY", Reserved, true},
	{"INNER", Reserved, true},
	{"INSERT", Reserved, true},
	{"INSTEAD", Reserved, true},
	{"INTERSECT", Reserved, true},
	{"INTO", Reserved, true},
	{"IS", Reserved, true},
	{"ISNULL", Reserved, true},
	{"JOIN", Reserved, true},
	{"KEY", Reserved, true},
	{"LAST", Reserved, true},
	{"LEFT", Reserved, true},
	{"LIKE", Reserved, true},
	{"LIMIT", Reserved, true},
	{"MATCH", Reserved, false}, // Turso doesn't support MATCH operator
	{"MATERIALIZED", Reserved, false}, // Turso doesn't support MATERIALIZED in CTEs
	{"NATURAL", Reserved, true},
	{"NO", Reserved, true},
	{"NOT", Reserved, true},
	{"NOTHING", Reserved, true},
	{"NOTNULL", Reserved, true},
	{"NULL", Reserved, true},
	{"NULLS", Reserved, true},
	{"OF", Reserved, true},
	{"OFFSET", Reserved, true},
	{"ON", Reserved, true},
	{"OR", Reserved, true},
	{"ORDER", Reserved, true},
	{"OTHERS", Reserved, true},
	{"OUTER", Reserved, true},
	{"OVER", Reserved, false}, // Turso doesn't support window functions (OVER clause)
	{"PARTITION", Reserved, false}, // Turso doesn't support PARTITION in window functions
	{"PLAN", Reserved, true},
	{"PRAGMA", Reserved, true},
	{"PRECEDING", Reserved, true},
	{"PRIMARY", Reserved, true},
	{"QUERY", Reserved, true},
	{"RAISE", Reserved, false}, // Turso doesn't support RAISE() function
	{"RANGE", Reserved, true},
	{"RECURSIVE", Reserved, false}, // Turso doesn't support RECURSIVE keyword in CTEs
	{"REFERENCES", Reserved, true},
	{"REGEXP", Reserved, false}, // Turso doesn't support REGEXP operator
	{"REINDEX", Reserved, true},
	{"RELEASE", Reserved, true},
	{"RENAME", Reserved, true},
	{"REPLACE", Reserved, true},
	{"RESTRICT", Reserved, true},
	{"RETURNING", Reserved, true},
	{"RIGHT", Reserved, true},
	{"ROLLBACK", Reserved, true},
	{"ROW", Reserved, true},
	{"ROWS", Reserved, true},
	{"SAVEPOINT", Reserved, true},
	{"SELECT", Reserved, true},
	{"SET", Reserved, true},
	{"TABLE", Reserved, true},
	{"TEMP", Reserved, true},
	{"TEMPORARY", Reserved, true},
	{"THEN", Reserved, true},
	{"TIES", Reserved, true},
	{"TO", Reserved, true},
	{"TRANSACTION", Reserved, true},
	{"TRIGGER", Reserved, true},
	{"UNBOUNDED", Reserved, true},
	{"UNION", Reserved, true},
	{"UNIQUE", Reserved, true},
	{"UPDATE", Reserved, true},
	{"USING", Reserved, true},
	{"VACUUM", Reserved, true},
	{"VALUES", Reserved, true},
	{"VIEW", Reserved, true},
	{"VIRTUAL", Reserved, true},
	{"WHEN", Reserved, true},
	{"WHERE", Reserved, true},
	{"WINDOW", Reserved, false}, // Turso doesn't support window functions
	{"WITH", Reserved, true},
	{"WITHOUT", Reserved, true},
}

// keywordMap provides O(1) keyword lookup by name (case-insensitive)
var keywordMap map[string]*Keyword

func init() {
	keywordMap = make(map[string]*Keyword, len(allKeywords))
	for i := range allKeywords {
		kw := &allKeywords[i]
		// Store in uppercase for case-insensitive lookup
		keywordMap[kw.Name] = kw
	}
}

// IsKeyword checks if a string is a SQLite keyword (case-insensitive).
// Returns true if the string matches any SQLite keyword.
func IsKeyword(name string) bool {
	_, exists := keywordMap[toUpper(name)]
	return exists
}

// IsReserved checks if a string is a reserved SQLite keyword (case-insensitive).
// Reserved keywords must be quoted when used as identifiers.
func IsReserved(name string) bool {
	kw, exists := keywordMap[toUpper(name)]
	return exists && kw.Type == Reserved
}

// IsTursoSupported checks if a keyword is supported by Turso LibSQL.
// Returns false if the keyword is not supported by Turso, true otherwise.
// If the string is not a keyword, returns true (assumes supported).
func IsTursoSupported(name string) bool {
	kw, exists := keywordMap[toUpper(name)]
	if !exists {
		return true // Not a keyword, so no restriction
	}
	return kw.TursoSupport
}

// GetKeyword returns keyword information for a given name (case-insensitive).
// Returns nil if the name is not a keyword.
func GetKeyword(name string) *Keyword {
	return keywordMap[toUpper(name)]
}

// AllKeywords returns a copy of all SQLite keywords.
func AllKeywords() []Keyword {
	result := make([]Keyword, len(allKeywords))
	copy(result, allKeywords)
	return result
}

// toUpper converts a string to uppercase for case-insensitive comparison.
// This is a simple ASCII uppercase conversion for SQL keywords.
func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		result[i] = c
	}
	return string(result)
}
