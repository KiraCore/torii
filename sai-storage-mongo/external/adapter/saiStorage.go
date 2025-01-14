package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SaiStorage struct {
	Url   string
	Token string
}

type SaiStorageResponse struct {
	Status string                   `json:"Status"`
	Result []map[string]interface{} `json:"result"`
	Count  int                      `json:"count"`
}

func (s *SaiStorage) Send(request Request) (*SaiStorageResponse, error) {
	// Define the request body
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create a new POST request with the request body
	req, err := http.NewRequest("POST", s.Url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add the Token header to the request
	req.Header.Set("Token", s.Token)

	// Send the request and get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response body into the struct
	var result = new(SaiStorageResponse)
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %v", err)
	}

	return result, nil
}
