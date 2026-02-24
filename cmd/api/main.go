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

	"github.com/Lee26Ed/lockit_appointments/cmd/api/types"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	_ "github.com/lib/pq"
)

const appVersion = "1.0.0"


type applicationDependencies struct {
    config types.ServerConfig
    logger *slog.Logger
	models *data.Models
}

func main() {
	var settings types.ServerConfig

	flag.IntVar(&settings.Port, "port", 4000, "Server port")
    flag.StringVar(&settings.Environment, "env", "development",
                  "Environment(development|staging|production)")
	flag.StringVar(&settings.DSN, "dsn", "postgres://username:password@localhost/db_name?sslmode=disable", "PostgreSQL DSN")
    flag.Parse()
	settings.AppVersion = appVersion

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// connect to db 
	db, err := sql.Open("postgres", settings.DSN)
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
        models: data.CreateModels(db),
    }
	
	err = app.Serve()
	if err != nil {
		log.Fatal(err)
	}

}