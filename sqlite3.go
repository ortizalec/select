package main

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
)

type QueryBuilder interface {
	From(string) *QueryBuilder
	Select(string) *QueryBuilder
	Eq(string, interface{}) *QueryBuilder
	Limit(int) *QueryBuilder
	Delete(string) *QueryBuilder
	OrderBy(string) *QueryBuilder
	Query() (*sql.Rows, error)
	QueryRow() *sql.Row
}

type KeyValPair struct {
	key   string
	value interface{}
}

func (kvp *KeyValPair) String() string {
	return fmt.Sprintf("%s = '%v'", kvp.key, kvp.value)
}

func keyValPairsToStrings(pairs []KeyValPair) []string {
	result := make([]string, len(pairs))
	for i, pair := range pairs {
		result[i] = pair.String()
	}
	return result
}

type SqliteConnection struct {
	DB        *sql.DB
	statement string
	table     string
	eq        []KeyValPair
	sel       []string
	limit     int
	orderby   string
	Err       error
}

func (s *SqliteConnection) String() string {
	return s.statement
}

func (s *SqliteConnection) Limit(l int) *SqliteConnection {
	if s.Err != nil {
		return s
	}
	s.limit = l
	s.build()
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
	return s
}

func (s *SqliteConnection) Select(columns string) *SqliteConnection {
	if s.Err != nil {
		return s
	}
	if columns == "" {
		s.build()
		return s
	}

	cols := strings.Split(columns, ",")
	for i := range cols {
		cols[i] = strings.TrimSpace(cols[i])
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

	s.sel = cols
	s.build()
	return s
}

func (s *SqliteConnection) Eq(column string, value interface{}) *SqliteConnection {
	if s.Err != nil {
		return s
	}
	s.eq = append(s.eq, KeyValPair{value: value, key: column})
	s.build()
	return s
}

func (s *SqliteConnection) OrderBy(column string) *SqliteConnection {
	if s.Err != nil {
		return s
	}
	s.orderby = column
	s.build()
	return s
}

func (s *SqliteConnection) build() *SqliteConnection {
	var stmt strings.Builder
	if len(s.sel) == 0 {
		stmt.WriteString("SELECT * ")
	} else {
		stmt.WriteString(fmt.Sprintf("SELECT %s ", strings.Join(s.sel, ", ")))
	}
	stmt.WriteString(fmt.Sprintf("FROM %s ", s.table))
	if len(s.eq) > 0 {
		stmt.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(keyValPairsToStrings(s.eq), " AND ")))
	}
	if s.orderby != "" {
		stmt.WriteString(fmt.Sprintf("ORDER BY %s ", s.orderby))
	}
	if s.limit != 0 {
		stmt.WriteString(fmt.Sprintf("LIMIT %d", s.limit))
	}
	s.statement = stmt.String()
	return s
}

func (s *SqliteConnection) Query() (*sql.Rows, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	rows, err := s.DB.Query(s.statement)
	if err != nil {
		log.Fatal(err)
	}
	return rows, err
}

func (s *SqliteConnection) QueryRow() *sql.Row {
	if s.Err != nil {
		log.Fatal(s)
		return nil
	}
	row := s.DB.QueryRow(s.statement)
	return row
}
