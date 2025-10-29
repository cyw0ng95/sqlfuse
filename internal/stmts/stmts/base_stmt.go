package stmts

// BaseStmt provides a reusable implementation of the Stmt interface.
// Statement types can embed this struct to avoid repeating the same
// SQL(), Type(), and Flavor() method implementations.
//
// This applies the Template Method pattern to eliminate boilerplate
// across statement types.
type BaseStmt struct {
	sql    string
	typ    string
	flavor FlavorConfig
}

// NewBaseStmt creates a new base statement with the given properties.
func NewBaseStmt(sql, typ string, flavor FlavorConfig) *BaseStmt {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	return &BaseStmt{
		sql:    sql,
		typ:    typ,
		flavor: flavor,
	}
}

// SQL returns the SQL string for this statement.
func (s *BaseStmt) SQL() string {
	return s.sql
}

// Type returns the statement type identifier.
func (s *BaseStmt) Type() string {
	return s.typ
}

// Flavor returns the SQL flavor/dialect configuration.
func (s *BaseStmt) Flavor() FlavorConfig {
	return s.flavor
}
