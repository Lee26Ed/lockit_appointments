package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func main() {

	// ! DONT HARD CODE DSN
	dsn := "postgres://lockit_appointments:one2enter@localhost/lockit_appointments?sslmode=disable"

	// connect to db 
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// context to die in 5 seconds 
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// test db connection 
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	fmt.Println("Successfully connected with context timeout")

	// testing context
	_, err = db.ExecContext(ctx, "SELECT pg_sleep(10);")
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Context cancelled: Postgres took too long")
		} else {
			log.Fatal("Query error: ", err)
		}
		return
	}
	
	mux := routes()

	fmt.Print("Server is running on port 4000\n")
	err = http.ListenAndServe(":4000", mux)
	if err != nil {
		panic(err)
	}
}