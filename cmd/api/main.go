package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/shyngys9219/greenlight/internal/data"
	"github.com/shyngys9219/greenlight/internal/jsonlog"
	"github.com/shyngys9219/greenlight/internal/mailer"
	// undescore (alias) is used to avoid go compiler complaining or erasing this
	// library.
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// Add a db struct field to hold the configuration settings for our database connection
// pool. For now this only holds the DSN, which we will read in from a command-line flag.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string // a conenction string to a sql server
		maxOpenConns int    // limit on the number of ‘open’ connections
		maxIdleConns int    // limit on the number of idle connections in the pool
		maxIdleTime  string // the maximum length of time that a connection can be idle
		// maxLifetime  string //optional here; maximum length of time that a connection can be reused for
	}

	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	// smtp sever credentials & sender (email) info
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger // new customized logger
	models data.Models     // hold new models in app
	mailer mailer.Mailer   // use ower mailer from mailer.go
	// used to wait for a collection of goroutines to finish their work
	wg sync.WaitGroup
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.
	// in powershell use next command: $env:DSN="postgres://postgres:1210@localhost:5433/greenlight?sslmode=disable"
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DSN"), "PostgreSQL DSN")

	// Setting restrictions on db connections
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")
	// flag.StringVar(&cfg.db.maxLifetime, "db-max-lifetime", "1h", "PostgreSQL max idle time")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	// use your own credentials here as username and password
	// $env:SMTPUSERNAME="smtp_server_username_here"
	// $env:SMTPPASSWORD="smtp_server_username_here"
	flag.StringVar(&cfg.smtp.username, "smtp-username", "f829dbe6a516d7", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "6b891d006e84e6", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Test <from@example.com>", "SMTP sender")

	flag.Parse()
	// Using new json oriented logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil) // calling PrintFatal function if there is an error with db server connection
	}
	// db will be closed before main function is completed.
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil) // printing custom info if db server connection is established

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // data.NewModels() function to initialize a Models struct
		// Initialize a new Mailer instance using the settings from the command line
		// flags, and add it to the application struct.
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	// new way of declaration of server part

	// reuse defined variable err
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	// optional lifetime limit, to use this, uncomment db substruct field and corresponding flag stringvar
	// lifetime, err := time.ParseDuration(cfg.db.maxIdleTime)
	// if err != nil {
	// 	return nil, err
	// }
	// db.SetConnMaxLifetime(lifetime)

	//context with a 5 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx) //create a connection and verify that everything is set up correctly.

	if err != nil {
		return nil, err
	}

	return db, nil
}
