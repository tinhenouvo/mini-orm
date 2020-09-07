package mini_orm

import (
	sq "github.com/Masterminds/squirrel"
)

// ConditionExpr condition expr
type ConditionExpr map[string]interface{}

// Sqlizer convert to sq.Sqlizer
type Sqlizer interface {
	ToSqlizer() sq.Sqlizer
}

// Condition just condition
type Condition struct {
	Expr interface{}
}

// Eq e.g. Eq{"name": "qingning", "id": [1, 2, 3]} => name="qingning" id IN [1, 2, 3]
type Eq ConditionExpr

// ToSqlizer to sq.Eq
func (c Eq) ToSqlizer() sq.Sqlizer {
	e := sq.Eq{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// Ne like Eq
type Ne ConditionExpr

// ToSqlizer to sq.NotEq
func (c Ne) ToSqlizer() sq.Sqlizer {
	e := sq.NotEq{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// Like e.g. Like{"name": "%laojun%"} => name LIKE "%laojun%"
type Like ConditionExpr

// ToSqlizer to sq.Like
func (c Like) ToSqlizer() sq.Sqlizer {
	e := sq.Like{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// NotLike e.g. NotLike{"name": "%laojun%"} => name NOT LIKE "%laojun%"
type NotLike ConditionExpr

// ToSqlizer to sq.NotLike
func (c NotLike) ToSqlizer() sq.Sqlizer {
	e := sq.NotLike{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// LT e.g. LT{"id": 12} => id < 12
type LT ConditionExpr

// ToSqlizer to sq.Lt
func (c LT) ToSqlizer() sq.Sqlizer {
	e := sq.Lt{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// LTE e.g. LT{"id": 12} => id <= 12
type LTE ConditionExpr

// ToSqlizer to sq.LtOrEq
func (c LTE) ToSqlizer() sq.Sqlizer {
	e := sq.LtOrEq{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// GT e.g. LT{"id": 12} => id > 12
type GT ConditionExpr

// ToSqlizer to sq.Gt
func (c GT) ToSqlizer() sq.Sqlizer {
	e := sq.Gt{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// GTE e.g. LT{"id": 12} => id >= 12
type GTE ConditionExpr

// ToSqlizer to sq.GtOrEq
func (c GTE) ToSqlizer() sq.Sqlizer {
	e := sq.GtOrEq{}
	for k, v := range c {
		e[k] = v
	}
	return e
}

// AND and expr
type AND []Sqlizer

// OR or expr
type OR []Sqlizer
