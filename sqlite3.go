package main

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
)

type QueryBuilder interface {
	From(string) *QueryBuilder
	Select(string) *QueryBuilder
	Eq(string) *QueryBuilder
	Single() *QueryBuilder
	Delete(string) *QueryBuilder
}

type SqliteConnection struct {
	DB    *sql.DB
	Query string
	table string
	Err   error
}

func (s *SqliteConnection) String() string {
	return s.Query
}

func (s *SqliteConnection) Single() *SqliteConnection {
	if s.Err != nil {
		return s
	}
	s.Query += "LIMIT 1"
	s.Err = nil
	return s
}

func (s *SqliteConnection) From(table string) *SqliteConnection {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"

	err := s.DB.QueryRow(query, table).Scan()
	if err != nil {
		if err == sql.ErrNoRows {
			s.Err = fmt.Errorf("table '%s' does not exist: %w", table, err)
			return s
		}
	}
	s.table = table

	s.Query += fmt.Sprintf("FROM %s ", table)
	s.Err = nil
	return s
}

func (s *SqliteConnection) Select(columns string) *SqliteConnection {
	if s.Err != nil {
		return s
	}
	cols := strings.Split(columns, ",")
	for i := range cols {
		cols[i] = strings.TrimSpace(cols[i])
	}
	// if you include * return early here
	if slices.Contains(cols, "*") {
		var stmt = "SELECT * "
		stmt += s.Query
		s.Query = stmt
		s.Err = nil
		return s
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", s.table)
	rows, err := s.DB.Query(query)
	if err != nil {
		s.Err = err
		return s
	}
	defer rows.Close()

	var columnNames []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		err = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			s.Err = fmt.Errorf("failed trying to validate column names: %w", err)
			return s
		}
		columnNames = append(columnNames, name)
	}

	for _, col := range cols {
		if !slices.Contains(columnNames, col) {
			s.Err = fmt.Errorf("column name '%s' does not exist on table '%s' available columns %s", col, s.table, strings.Join(columnNames, ", "))
			return s
		}
	}

	var stmt strings.Builder
	stmt.WriteString(fmt.Sprintf("SELECT %s ", strings.Join(cols, ", ")))
	stmt.WriteString(s.Query)
	s.Query = stmt.String()
	s.Err = nil
	return s
}
