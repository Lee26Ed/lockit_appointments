package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type serverConfig struct {
    port int 
    environment string
	dsn string
}

type applicationDependencies struct {
    config serverConfig
    logger *slog.Logger
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")
    flag.StringVar(&settings.environment, "env", "development",
                  "Environment(development|staging|production)")
	flag.StringVar(&settings.dsn, "dsn", "postgres://username:password@localhost/db_name?sslmode=disable", "PostgreSQL DSN")
    flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// connect to db 
	db, err := sql.Open("postgres", settings.dsn)
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

	app := &applicationDependencies {
        config: settings,
        logger: logger,
    }
	
	err = app.Serve()
	if err != nil {
		log.Fatal(err)
	}

}