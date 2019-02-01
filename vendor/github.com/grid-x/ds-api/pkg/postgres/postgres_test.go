package postgres_test

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/pkg/postgres"
)

// DB represents the database connection.
var DB *sqlx.DB
var LOGGER *log.Logger

func TestMain(m *testing.M) {
	var (
		postgresURL  = flag.String("postgres.url", "postgres://postgres:pw@192.168.99.100/postgres?sslmode=disable", "postgresDB for testing")
		skipTruncate = flag.Bool("skip-truncate", false, "skip truncating database")
		logVerbose   = flag.Bool("log-debug", false, "log verbose output")
	)
	flag.Parse()

	if testing.Short() {
		fmt.Println("skipped all tests")
		os.Exit(0)
	}

	if *postgresURL == "" {
		fmt.Println("flag: no postgres_url set")
		os.Exit(1)
	}
	var err error
	DB, err = sqlx.Open("postgres", *postgresURL)
	if err != nil {
		fmt.Printf("postgres: %+v\n", err)
		os.Exit(1)
	}

	n, err := postgres.Migrate(DB)
	if err != nil {
		fmt.Printf("postgres: %+v\n", err)
		DB.Close()
		os.Exit(1)
	}
	fmt.Printf("postgres: %d migrations\n", n)

	// Empty database.
	if !*skipTruncate {
		truncate()
	}

	// Init Logger
	logLevel := log.ErrorLevel

	if *logVerbose {
		logLevel = log.DebugLevel
	}

	LOGGER = &log.Logger{
		Out:       os.Stderr,
		Formatter: new(log.TextFormatter),
		Hooks:     make(log.LevelHooks),
		Level:     logLevel,
	}

	code := m.Run()
	LOGGER.Infof("Closed DB with %d open connections.", DB.Stats().OpenConnections)
	DB.Close()

	os.Exit(code)
}

func truncate() {
	DB.MustExec("TRUNCATE TABLE users CASCADE;")
	DB.MustExec("TRUNCATE TABLE accounts CASCADE;")
}

// InitTest clears the database and seeds the random
func InitTest() {
	truncate()
	rand.Seed(int64(time.Now().UnixNano()))
}
