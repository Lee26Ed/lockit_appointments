package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/types"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/mailer"
	_ "github.com/lib/pq"
)

const appVersion = "1.1.0"

var smtpHost = os.Getenv("SMTP_HOST")
var smtpPort = 587 // default SMTP port
var smtpUsername = os.Getenv("SMTP_USERNAME")
var smtpPassword = os.Getenv("SMTP_PASSWORD")
var smtpSender = os.Getenv("SMTP_SENDER")

type applicationDependencies struct {
    config types.ServerConfig
    logger *slog.Logger
	models *data.Models
	mailer mailer.Mailer
	wg sync.WaitGroup
}

func loadConfig() types.ServerConfig {
	var settings types.ServerConfig
	flag.IntVar(&settings.Port, "port", 4000, "Server port")
	flag.StringVar(&settings.Environment, "env", "development",
				  "Environment(development|staging|production)")
	flag.StringVar(&settings.DSN, "dsn", "postgres://username:password@localhost/db_name?sslmode=disable", "PostgreSQL DSN")

	// rate limiting settings
	flag.Float64Var(&settings.Limiter.RPS, "limiter-rps", 2,
		"Rate Limiter maximum requests per second")

	flag.IntVar(&settings.Limiter.Burst, "limiter-burst", 5,
		"Rate Limiter maximum burst")

	flag.BoolVar(&settings.Limiter.Enabled, "limiter-enabled", true,
		"Enable rate limiter")

	// CORS settings
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)",
		func(val string) error {
			settings.CORS.TrustedOrigins = strings.Fields(val)
			return nil
		})

	// SMTP settings
	flag.StringVar(&settings.SMTP.Host, "smtp-host", smtpHost, "SMTP host")
	flag.IntVar(&settings.SMTP.Port, "smtp-port", smtpPort, "SMTP port")
	flag.StringVar(&settings.SMTP.Username, "smtp-username", smtpUsername, "SMTP username")
	flag.StringVar(&settings.SMTP.Password, "smtp-password", smtpPassword, "SMTP password")
	flag.StringVar(&settings.SMTP.Sender, "smtp-sender", smtpSender, "SMTP sender email")

	flag.Parse()
	settings.AppVersion = appVersion

	return settings 

}

func main() {

	settings := loadConfig()
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
		wg: sync.WaitGroup{},
		mailer: mailer.New(settings.SMTP.Host, settings.SMTP.Port, settings.SMTP.Username, settings.SMTP.Password, settings.SMTP.Sender),
    }

	// Publish basic expvar metrics
	expvar.NewString("version").Set(app.config.AppVersion)
	expvar.NewString("env").Set(app.config.Environment)
	expvar.Publish("goroutines", expvar.Func(func() any { return runtime.NumGoroutine() }))
	expvar.Publish("database", expvar.Func(func() any { return db.Stats() }))

	
	err = app.Serve()
	if err != nil {
		log.Fatal(err)
	}

}