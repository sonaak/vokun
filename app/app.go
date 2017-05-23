package app

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/evilwire/go-env"
	"github.com/go-zoo/bone"
	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"github.com/sonaak/vokun/models"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"
)

type SqlClient interface {
	Begin() (SqlTx, error)
	Close() error
	Exec(string, ...interface{}) (sql.Result, error)
	Ping() error
	Prepare(string) (SqlStmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
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

func (pq *PostgresClient) Exec(query string, args ...interface{}) (sql.Result, error) {
	return pq.DB.Exec(query, args...)
}

func (pq *PostgresClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return pq.DB.Query(query, args...)
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
		"host=%s user=%s password=%s dbname=vokun sslmode=disable",
		config.Db.Host,
		config.Db.User,
		config.Db.Password,
	)
	glog.Infof("Data base config: %s", dbConfig)

	sqlClient, err := sql.Open("postgres", dbConfig)
	if err != nil {
		return nil, err
	}

	// create some configurations
	external := External()
	mux := external.newMux()
	app := NewApp(config, mux, &PostgresClient{DB: sqlClient})

	// setup the app's routers
	mux.GetFunc("/healthcheck", app.GracefullyRun(app.HealthCheck))
	mux.GetFunc("/api/*", app.GracefullyRun(app.Serve))
	mux.PostFunc("/api/*", app.GracefullyRun(app.Serve))

	// get the history of a particular URL
	mux.GetFunc("/history/*", app.GracefullyRun(app.GetHistory))

	return app, nil
}

type Mux interface {
	GetFunc(string, http.HandlerFunc) *bone.Route
	PostFunc(string, http.HandlerFunc) *bone.Route
	http.Handler
}

func NewApp(config *Config, mux Mux, db SqlClient) *App {
	return &App{
		Config: config,
		Mux:    mux,
		Db:     db,
		RequestSource: &DbRequestSource{
			SqlClient: db,
		},
	}
}

type App struct {
	*Config
	Mux           Mux
	Db            SqlClient
	RequestSource models.RequestSource
}

func (app *App) GracefullyRun(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(500)
				err, ok := r.(error)
				if !ok {
					writer.Write([]byte(`Could not recover error:`))
					writer.Write(debug.Stack())
					return
				}
				writer.Write([]byte(err.Error()))
				writer.Write(debug.Stack())
			}
		}()

		handler(writer, request)
	}
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

func (app *App) RecordRequest(request *http.Request, ts time.Time) error {
	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}

	queryParams := []models.QueryParam{}
	for k, v := range request.URL.Query() {
		queryParams = append(queryParams, models.QueryParam{
			Name:  k,
			Value: v,
		})
	}

	headers := []models.Header{}
	for k, v := range request.Header {
		headers = append(headers, models.Header{
			Name:  k,
			Value: v,
		})
	}

	_, putErr := app.RequestSource.PutRequest(&models.Request{
		Headers:     headers,
		QueryParams: queryParams,
		Subpath:     request.RequestURI[5:],
		Timestamp:   ts,
		Body:        bodyBytes,
	})

	// BUG: this is a very strange behaviour. Although putErr is nil
	// returning putErr would automatically generate a non-nil *ModelError
	// with an nil Err; I am not sure about the mechanism
	if putErr == nil {
		return nil
	}

	return putErr
}

func (app *App) Serve(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	recordingErr := app.RecordRequest(request, time.Now())

	// if you cannot record the request, then fail 100% because
	// this is a server that needs to be able to answer questions
	// about its requests
	if recordingErr != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(fmt.Sprintf("Error: %v", recordingErr)))
		return
	}

	resp, err := GetResponse(request.RequestURI[5:], request.Method)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	header := writer.Header()
	for k, v := range resp.Headers {
		header.Add(k, v)
	}
	writer.Write(resp.Content)
}

func (app *App) Run() error {
	// run stuff forever in here
	glog.V(2).Infof("Starting App. Listening on port 9000")
	return http.ListenAndServe(":9000", app.Mux)
}

type HistoryResponse struct {
	Meta struct {
		Status  status `json:"status"`
		Message string `json:"message"`
		Count   uint   `json:"count"`
	} `json:"meta"`
	Data []models.Request `json:"data"`
}

func (app *App) GetHistory(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	subpath := request.RequestURI[9:]
	requests, modelErr := app.RequestSource.FindRequests(subpath)
	if modelErr != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(fmt.Sprintf("Error: %v", modelErr)))
		return
	}

	header := writer.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(200)

	requestJson, _ := json.Marshal(&HistoryResponse{
		Meta: struct {
			Status  status `json:"status"`
			Message string `json:"message"`
			Count   uint   `json:"count"`
		}{
			Status:  OK,
			Message: "",
			Count:   uint(len(requests)),
		},
		Data: requests,
	})
	writer.Write(requestJson)
}
