// Command xo generates code from database schemas and custom queries. Works
// with PostgreSQL, MySQL, Microsoft SQL Server, Oracle Database, and SQLite3.
package main

//go:generate ./gen.sh models
//go:generate go generate ./internal

import (
	"context"
	"database/sql"
	"fmt"
	// drivers
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/sijms/go-ora/v2"

	"github.com/calebhiebert/xo/cmd"
)

func main() {
	db, err := sql.Open("postgres", "postgres://f4:f4@localhost:5432/f4?sslmode=disable")
	if err != nil {
		panic(err)
	}

	intro, err := cmd.IntrospectSchema(context.Background(), "public", "postgres", db)
	if err != nil {
		panic(err)
	}

	fmt.Println(intro)
}
