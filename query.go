package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type TodoQuery struct {
	priorityMin sql.NullInt16
	priorityMax sql.NullInt16
	completed   sql.NullBool
	fulltext    sql.NullString
}

func NewTodoQuery(options []TodoQueryOption) TodoQuery {
	var res TodoQuery

	for i := range options {
		options[i](&res)
	}
	return res
}

type TodoQueryOption func(*TodoQuery)

func WithPriorityMin(min int) TodoQueryOption {
	return func(tq *TodoQuery) {
		tq.priorityMin = sql.NullInt16{Int16: int16(min), Valid: true}
	}
}
func WithPriorityMax(max int) TodoQueryOption {
	return func(tq *TodoQuery) {
		tq.priorityMax = sql.NullInt16{Int16: int16(max), Valid: true}
	}
}

func WithCompleted(val bool) TodoQueryOption {
	return func(tq *TodoQuery) {
		tq.completed = sql.NullBool{Bool: val, Valid: true}
	}
}

func WithFulltext(val string) TodoQueryOption {
	return func(tq *TodoQuery) {
		tq.fulltext = sql.NullString{String: val, Valid: true}
	}
}

// Returns Postgres SQL query with $1... placeholders
// and a slice of corresponding arguments.
//
// Placeholder used only for full-text search string
// because other values are safe.
func (tq *TodoQuery) SQL() (query string, args []interface{}) {
	cond := make([]string, 0)
	i := 0 // args counter; pg query placeholder index
	if tq.completed.Valid {
		cond = append(cond, fmt.Sprintf("completed=%v", tq.completed.Bool))
	}
	if tq.priorityMin.Valid && tq.priorityMax.Valid {
		cond = append(cond, fmt.Sprintf("priority BETWEEN %d AND %d", tq.priorityMin.Int16, tq.priorityMax.Int16))
	} else if tq.priorityMin.Valid {
		cond = append(cond, fmt.Sprintf("priority >= %d", tq.priorityMin.Int16))
	} else if tq.priorityMax.Valid {
		cond = append(cond, fmt.Sprintf("priority <= %d", tq.priorityMax.Int16))
	}
	if tq.fulltext.Valid {
		i++
		cond = append(cond, fmt.Sprintf("ts_index @@ to_tsquery($%d)", i))
		args = append(args, tq.fulltext.String)
	}

	return strings.Join(cond, " AND "), args
}
