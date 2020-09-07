package mini_orm

import (
	sq "github.com/Masterminds/squirrel"
)

// StatementType statement type
type StatementType int

const (
	UnknownStatement StatementType = 0
	InsertStatement  StatementType = 1
	UpdateStatement  StatementType = 2
	DeleteStatement  StatementType = 3
	SelectStatement  StatementType = 4
)

// Statement statement
type Statement struct {
	stType     StatementType
	table      string
	columns    []string
	limit      uint64
	offset     uint64
	orderBys   []string
	conditions []Condition
	values     [][]interface{}
}

// Reset Statement Reset
func (st *Statement) Reset() {
	st.stType = UnknownStatement
	st.table = ""
	st.columns = make([]string, 0)
	st.offset = 0
	st.limit = 0
	st.conditions = make([]Condition, 0)
	st.orderBys = make([]string, 0)
	st.values = make([][]interface{}, 0)
}

// Select set select statment
func (st *Statement) Select(columns ...string) *Statement {
	st.Reset()
	st.stType = SelectStatement
	if len(columns) > 0 {
		st.Columns(columns...)
	}
	return st
}

// From set sql table
func (st *Statement) From(table string) *Statement {
	st.table = table
	return st
}

// Columns set sql columns atttentio Columns will reset st.columns
func (st *Statement) Columns(columns ...string) *Statement {
	cs := make([]string, 0)
	cs = append(cs, columns...)
	st.columns = cs
	return st
}

// Insert set insert statement
func (st *Statement) Insert() *Statement {
	st.Reset()
	st.stType = InsertStatement
	return st
}

// Update set update statement
func (st *Statement) Update() *Statement {
	st.Reset()
	st.stType = UpdateStatement
	return st
}

// Delete set delete statement
func (st *Statement) Delete() *Statement {
	st.Reset()
	st.stType = DeleteStatement
	return st
}

// Offset set offset
func (st *Statement) Offset(offset uint64) *Statement {
	st.offset = offset
	return st
}

// Limit set limit
func (st *Statement) Limit(limit uint64) *Statement {
	st.limit = limit
	return st
}

// Where set condition
func (st *Statement) Where(expr ...interface{}) *Statement {
	for _, e := range expr {
		st.conditions = append(st.conditions, Condition{e})
	}
	return st
}

// OrderBy set orderby
func (st *Statement) OrderBy(orderby ...string) *Statement {
	for _, o := range orderby {
		st.orderBys = append(st.orderBys, o)
	}
	return st
}

// Values set values
func (st *Statement) Values(val []interface{}) *Statement {
	st.values = append(st.values, val)
	return st
}

// ToSQL gen SQl
func (st *Statement) ToSQL() (string, []interface{}, error) {
	if st.table == "" {
		return "", nil, StatementTableNotSet
	}
	if st.stType == UnknownStatement {
		return "", nil, StatementTypeNotSet
	}
	switch st.stType {
	case SelectStatement:
		var builder sq.SelectBuilder
		if len(st.columns) > 0 {
			builder = sq.Select(st.columns...)
		} else {
			builder = sq.Select("*")
		}
		builder = builder.From(st.table)
		for _, c := range st.conditions {
			builder = builder.Where(st.ConvertCondition(c.Expr))
		}
		if st.offset > 0 {
			builder = builder.Offset(st.offset)
		}
		if st.limit > 0 {
			builder = builder.Limit(st.limit)
		}
		if len(st.orderBys) > 0 {
			builder = builder.OrderBy(st.orderBys...)
		}
		return builder.ToSql()
	case DeleteStatement:
		builder := sq.Delete(st.table)
		for _, c := range st.conditions {
			builder = builder.Where(st.ConvertCondition(c.Expr))
		}
		if st.offset > 0 {
			builder = builder.Offset(st.offset)
		}
		if st.limit > 0 {
			builder = builder.Limit(st.limit)
		}
		if len(st.orderBys) > 0 {
			builder = builder.OrderBy(st.orderBys...)
		}
		return builder.ToSql()
	case InsertStatement:
		builder := sq.Insert(st.table)
		builder = builder.Columns(st.columns...)
		for _, v := range st.values {
			builder = builder.Values(v...)
		}
		return builder.ToSql()
	case UpdateStatement:
		builder := sq.Update(st.table)
		for _, v := range st.values {
			uval := make(map[string]interface{})
			for i, f := range st.columns {
				uval[f] = v[i]
			}
			builder = builder.SetMap(uval)
		}
		for _, c := range st.conditions {
			builder = builder.Where(st.ConvertCondition(c.Expr))
		}
		if st.offset > 0 {
			builder = builder.Offset(st.offset)
		}
		if st.limit > 0 {
			builder = builder.Limit(st.limit)
		}
		if len(st.orderBys) > 0 {
			builder = builder.OrderBy(st.orderBys...)
		}
		return builder.ToSql()
	}
	return "", nil, StatementTypeNotSet
}

// ConvertCondition convert condition to sq condition it will panic if convert not found
func (st *Statement) ConvertCondition(c interface{}) interface{} {
	switch expr := c.(type) {
	case Eq, Ne, Like, NotLike, GT, GTE, LT, LTE:
		sqlize := expr.(Sqlizer)
		return sqlize.ToSqlizer()
	case AND:
		e := sq.And{}
		for _, v := range expr {
			s := v.(Sqlizer)
			e = append(e, s.ToSqlizer())
		}
		return e
	case OR:
		e := sq.Or{}
		for _, v := range expr {
			s := v.(Sqlizer)
			e = append(e, s.ToSqlizer())
		}
		return e
	default:
		panic("ConvertCondition not support")
	}
}
