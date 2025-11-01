package stmts

import (
	"database/sql"
	"sqlfuse/internal/common"
)

// GenSelectJoin dispatches to one of the join variants supported by SQLite/Turso.
func GenSelectJoin(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	switch lcg.Intn(6) {
	case 0:
		return GenSelectCrossJoin(db, lcg)
	case 1:
		return GenSelectInnerJoin(db, lcg)
	case 2:
		return GenSelectOuterJoin(db, lcg)
	case 3:
		return GenSelectJoinUsing(db, lcg)
	case 4:
		return GenSelectNaturalJoin(db, lcg)
	default:
		return GenSelectInnerJoin(db, lcg)
	}
}
