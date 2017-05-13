package app

import (
	"net/http"
	"github.com/go-zoo/bone"
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/evilwire/go-env"
	"fmt"
)

type SqlClient interface {
	Begin() (SqlTx, error)
	Close() (error)
	Exec(string, ...interface{}) (sql.Result, error)
	Ping() error
	Prepare(string) (SqlStmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	Stats() sql.DBStats
}


type SqlTx interface {
}


type SqlStmt interface {
}


type PostgresClient struct {
	DB *sql.DB
	SqlClient
}

func (pq *PostgresClient) Ping() error {
	return pq.DB.Ping()
}

func (pq *PostgresClient) Stats() sql.DBStats {
	return pq.DB.Stats()
}


func Setup() (*App, error) {
	flag.Parse()

	envReader := goenv.NewOsEnvReader()
	config, err := NewConfig(envReader)
	if err != nil {
		return nil, err
	}

	// create a database connection
	dbConfig := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=morokei",
		config.Db.Host,
		config.Db.User,
		config.Db.Password,
	)

	sqlClient, err := sql.Open("postgres", dbConfig)
	if err != nil {
		return nil, err
	}

	// create some configurations
	external := External()
	mux := external.newMux()
	app := NewApp(config, mux, &PostgresClient { DB: sqlClient })

	// setup the app's routers
	mux.GetFunc("/healthcheck", app.HealthCheck)

	return app, nil
}

type Mux interface {
	GetFunc(string, http.HandlerFunc) *bone.Route
	http.Handler
}

func NewApp(config *Config, mux Mux, db SqlClient) *App {
	return &App {
		Config: config,
		Mux: mux,
		Db: db,
	}
}

type App struct {
	*Config
	Mux Mux
	Db SqlClient
}


func (app *App) HealthCheck(writer http.ResponseWriter, request *http.Request) {
	header := writer.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")

	healthCheck := NewHealthChecker().CheckHealth(app.Db)
	hcJson, err := json.Marshal(healthCheck)
	if err != nil {
		glog.Errorf("Error processing /healthcheck: %v", err)
		writer.WriteHeader(500)
		writer.Write([]byte(""))
		return
	}

	writer.WriteHeader(200)
	writer.Write(hcJson)
}

func (app *App) Run() error {
	// run stuff forever in here
	glog.V(2).Infof("Starting App. Listening on port 9000")
	return http.ListenAndServe(":9000", app.Mux)
}
