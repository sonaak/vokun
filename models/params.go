package models

type Param struct {
	Name  string   `json:"name"`
	Value []string `json:"value"`
}

type Header Param

type QueryParam Param
