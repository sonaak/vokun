package app

import (
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/sonaak/vokun/models"
	"time"
)

const (
	// for finding things
	SELECT_TPL = `SELECT %s FROM models.%s WHERE %s`

	// for inserting (and erroring out if uniqueness constraints are not met)
	INSERT_TPL = `INSERT INTO models.%s (%s) VALUES (%s)`

	// for upserting (and not erroring out if uniqueness are met)
	UPSERT_TPL = INSERT_TPL + ` ON CONFLICT ON CONSTRAINT %s DO UPDATE SET (%s) = (%s)`

	// for deleting
	DELETE_TPL = `DELETE FROM models.%s WHERE %s`
)

type DbRequestSource struct {
	SqlClient
}

func (src *DbRequestSource) GetRequest(id uuid.UUID) (*models.Request, *models.ModelError) {
	return nil, nil
}

func (src *DbRequestSource) FindRequests(subpath string) ([]models.Request, *models.ModelError) {
	requests := []models.Request{}
	selectStmt := fmt.Sprintf(SELECT_TPL,
		"subpath, request_time, headers, query_params, body",
		"request",
		`subpath = $1`,
	)
	rows, queryErr := src.SqlClient.Query(selectStmt, subpath)
	if queryErr != nil {
		return requests, models.WrapErr(queryErr, models.INTERNAL)
	}
	defer rows.Close()

	for rows.Next() {
		var sp, hdrs, qps string
		var requestTime time.Time
		var body []byte
		rows.Scan(&sp, &requestTime, &hdrs, &qps, &body)

		headers := []models.Header{}
		queryParams := []models.QueryParam{}

		marshalErr := json.Unmarshal([]byte(hdrs), &headers)
		if marshalErr != nil {
			return requests, models.WrapErr(marshalErr, models.INTERNAL)
		}

		marshalErr = json.Unmarshal([]byte(qps), &queryParams)
		if marshalErr != nil {
			return requests, models.WrapErr(marshalErr, models.INTERNAL)
		}

		requests = append(requests, models.Request{
			Subpath:     subpath,
			Timestamp:   requestTime,
			Body:        body,
			Headers:     headers,
			QueryParams: queryParams,
		})
	}
	if rows.Err() != nil {
		return requests, models.WrapErr(rows.Err(), models.INTERNAL)
	}
	return requests, nil
}

func (src *DbRequestSource) PutRequest(request *models.Request) (uuid.UUID, *models.ModelError) {
	requestId := uuid.NewV4()

	insertStmt := fmt.Sprintf(INSERT_TPL,
		"request",
		`id, subpath, request_time, headers, query_params, body`,
		`$1, $2, $3, $4, $5, $6`)

	headerBytes, _ := json.Marshal(request.Headers)
	paramBytes, _ := json.Marshal(request.QueryParams)

	// get the insert string for request
	_, err := src.SqlClient.Exec(
		insertStmt,
		requestId.String(),
		request.Subpath,
		request.Timestamp,
		string(headerBytes),
		string(paramBytes),
		request.Body,
	)

	return requestId, models.WrapErr(err, models.INTERNAL)
}
