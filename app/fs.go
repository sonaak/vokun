package app

import (
	"io/ioutil"
	"fmt"
	"strings"
	"github.com/pkg/errors"
	"encoding/json"
)

type APIResponse struct {
	Headers map[string]string `json:"headers"`
	Code int `json:"status_code"`
	Content []byte
}

type APIResponseError struct {
	error
	Type string
}


func GetResponse(path, verb string) (*APIResponse, error) {
	filePath := fmt.Sprintf("/opt/sonaak/vokun-api/%s.%s",
		path, strings.ToLower(verb),
	)

	fBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Errorf("Cannot generate response: cannot read file %s",
			filePath,
		)
	}

	// seek "---" in a line by itself
	fStr := string(fBytes)
	index := strings.Index(fStr, "\n---\n")
	if index < 0 {
		return nil, errors.Errorf("Cannot generate response from file\n\n%s\n\n(path: %s)",
			fStr, filePath,
		)
	}

	// beginning
	apiResponse := &APIResponse{}
	unmarshalErr := json.Unmarshal([]byte(fStr[:index]), apiResponse)
	if unmarshalErr != nil {
		return nil, errors.Errorf("Cannot unmarshal API meta info from %s",
			fStr[:index],
		)
	}

	apiResponse.Content = []byte(fStr[index + 5:])
	return apiResponse, nil
}
