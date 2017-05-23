package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type Request struct {
	Timestamp   time.Time    `json:"timestamp"`
	Body        []byte       `json:"body"`
	Subpath     string       `json:"subpath"`
	QueryParams []QueryParam `json:"query-params"`
	Headers     []Header     `json:"headers"`
}

type RequestSource interface {
	GetRequest(id uuid.UUID) (*Request, *ModelError)
	FindRequests(subpath string) ([]Request, *ModelError)
	PutRequest(request *Request) (uuid.UUID, *ModelError)
}
