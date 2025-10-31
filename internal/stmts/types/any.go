package types

import "sqlfuse/internal/common"

// ValueForType returns an SQL literal string appropriate for the declared column type.
func ValueForType(typ string, lcg *common.LCG, hint string) string {
	if typ == "" {
		// fallback to string
		return StringLiteral(lcg, hint)
	}
	t := typ
	// simple case-insensitive checks
	if contains := func(substr string) bool { return len(t) >= len(substr) && (t == substr || t[:len(substr)] == substr) }; contains("INT") || contains("BIGINT") || contains("SMALLINT") {
		return IntLiteral(lcg, hint)
	}
	if contains := func(substr string) bool { return len(t) >= len(substr) && (t == substr || t[:len(substr)] == substr) }; contains("REAL") || contains("DEC") || contains("FLOA") || contains("DOUB") {
		return RealLiteral(lcg)
	}
	if contains := func(substr string) bool { return len(t) >= len(substr) && (t == substr || t[:len(substr)] == substr) }; contains("BLOB") {
		return BlobLiteral(lcg)
	}
	if contains := func(substr string) bool { return len(t) >= len(substr) && (t == substr || t[:len(substr)] == substr) }; contains("JSON") {
		return JSONLiteral(lcg)
	}
	// default
	return StringLiteral(lcg, hint)
}
