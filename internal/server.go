package internal

import (
	"database/sql"
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"test_project/internal/db"
	"time"
)

var (
	ErrInvalidRequestBody = &Err{"request body is invalid", 1}
	ErrInternalServer     = &Err{"internal server error", 2}
	ErrStartStoppedTask   = &Err{"unable to start stopped or finished task", 3}
	ErrStartRunningTask   = &Err{"unable to start task: task with this id is already running", 4}
	ErrTooLateToStop      = &Err{"too late to stop", 5}
)

type Request struct {
	RequestId   string `json:"requestId"`
	RequestType string `json:"type"`
}

type Err struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type Response struct {
	Data  interface{} `json:"data"`
	Error *Err        `json:"error"`
}

type server struct {
	db *pg.DB
}

type Env struct {
	Port       string `envconfig:"PORT"`
	PgUsername string `envconfig:"POSTGRES_USER"`
	PgDatabase string `envconfig:"POSTGRES_DB"`
	PgPassword string `envconfig:"POSTGRES_PASSWORD"`
	PgAddress  string `envconfig:"POSTGRES_ADDRESS"`
}

func runMigrations(vars Env) {
	var migrationsDbConn *sql.DB
	var err error
	postgresDsn := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", vars.PgUsername, vars.PgPassword,
		vars.PgAddress, vars.PgDatabase)
	for {
		migrationsDbConn, err = sql.Open("postgres", postgresDsn)
		if err != nil {
			log.Printf("error while opening pgDb: %v", err)
		}
		err = migrationsDbConn.Ping()
		if err != nil {
			log.Printf("error while opening pgDb: %v", err)
		}
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	defer migrationsDbConn.Close()
	if err := db.Migrate(migrationsDbConn); err != nil {
		log.Fatalf("error while migrating: %v", err)
	}

}

func Run() {
	var vars Env
	err := envconfig.Process("", &vars)
	if err != nil {
		log.Fatal(err)
	}
	runMigrations(vars)
	pgDb := pg.Connect(&pg.Options{
		User:     vars.PgUsername,
		Password: vars.PgPassword,
		Database: vars.PgDatabase,
		Addr:     vars.PgAddress,
	})
	server := &server{db: pgDb}
	r := mux.NewRouter()
	r.HandleFunc("/", server.Handler).Methods(http.MethodPost)
	http.Handle("/", r)
	log.Print("listening on ", vars.Port)
	log.Fatal(http.ListenAndServe(":"+vars.Port, nil))

}
