package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Person struct {
	ID   int    `db:"id"`
	Age  string `db:"age"`
	Name string `db:"name"`
}

func main() {
	db, err := sql.Open("sqlite3", "example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var mydb SqliteConnection
	mydb.DB = db
	mydb.Err = nil
	rows, err := mydb.From("users").Select("age, name").Eq("age", 25).Eq("name", "alec").OrderBy("name").Query()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to execute query: %w", err))
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var age string
		err := rows.Scan(&age, &name)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to scan rows %w", err))
		}
		log.Println(name, age)
	}

}
