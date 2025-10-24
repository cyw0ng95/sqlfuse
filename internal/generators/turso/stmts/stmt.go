package stmts

type Stmt interface {
	SQL() string
	Type() string // e.g. "pragma", "ddl", "dml"
}

type PragmaStmt struct {
	sql string
}

func (p *PragmaStmt) SQL() string  { return p.sql }
func (p *PragmaStmt) Type() string { return "pragma" }
