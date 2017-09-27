package grpcsql_test

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/mattn/go-sqlite3"
)

func Example() {
	server := httptest.NewUnstartedServer(grpcsql.NewServer(&sqlite3.SQLiteDriver{}))
	server.TLS = &tls.Config{NextProtos: []string{"h2"}}
	server.StartTLS()
	defer server.Close()

	targetsFunc := func() ([]string, error) {
		return []string{server.Listener.Addr().String()}, nil
	}
	driver := grpcsql.NewDriver(targetsFunc, tlsConfig, 2*time.Second)
	sql.Register("grpc", driver)

	db, err := sql.Open("grpc", ":memory:")
	if err != nil {
		log.Fatalf("failed to create grpc database: %s", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("failed to create grpc transaction: %s", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec("CREATE TABLE test (n INT)")
	if err != nil {
		log.Fatalf("failed to execute create table statement over grpc: %s", err)
	}

	result, err = tx.Exec("INSERT INTO test(n) VALUES (1)")
	if err != nil {
		log.Fatalf("failed to execute create table statement over grpc: %s", err)
	}

	rows, err := tx.Query("SELECT n FROM test")
	if err != nil {
		log.Fatalf("failed to select rows over grpc: %s", err)
	}
	numbers := []int{}
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			log.Fatalf("failed to scan row over grpc: %s", err)
		}
		numbers = append(numbers, n)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows error over grpc: %s", err)
	}
	defer rows.Close()

	// Output:
	// 1 <nil>
	// 1 <nil>
	// [1]
	fmt.Println(result.LastInsertId())
	fmt.Println(result.RowsAffected())
	fmt.Println(numbers)
}
