package internal

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"test_project/internal/db"
	"time"
)

type Request struct {
	requestId   string `json:"requestId"`
	requestType string `json:"type"`
}

type Err struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type Response struct {
	Data  interface{} `json:"data"`
	Error *Err        `json:"error"`
}

var (
	ErrInvalidRequestBody = &Err{"request body is invalid", 1}
)

type server struct {
	db *sql.DB
}

type Env struct {
	Port       string `envconfig:"PORT"`
	PgUsername string `envconfig:"POSTGRES_USER"`
	PgDatabase string `envconfig:"POSTGRES_DB"`
	PgPassword string `envconfig:"POSTGRES_PASSWORD"`
	PgAddress  string `envconfig:"POSTGRES_ADDRESS"`
}

func Run() {
	var vars Env
	err := envconfig.Process("", &vars)
	if err != nil {
		log.Fatal(err)
	}
	var pgDb *sql.DB
	postgresDsn := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", vars.PgUsername, vars.PgPassword,
		vars.PgAddress, vars.PgDatabase)
	for {
		pgDb, err = sql.Open("postgres", postgresDsn)
		if err != nil {
			log.Printf("error while opening pgDb: %v", err)
		}
		err = pgDb.Ping()
		if err != nil {
			log.Printf("error while opening pgDb: %v", err)
		}
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	defer pgDb.Close()
	if err := db.Migrate(pgDb); err != nil {
		log.Fatalf("error while migrating: %v", err)
	}
	server := &server{db: pgDb}
	r := mux.NewRouter()
	r.HandleFunc("/", server.Handler).Methods(http.MethodPost)
	http.Handle("/", r)
	log.Print("listening on ", vars.Port)
	log.Fatal(http.ListenAndServe(":"+vars.Port, nil))

}
