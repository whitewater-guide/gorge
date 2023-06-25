package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/whitewater-guide/gorge/core"
)

type HTTPClient struct {
	*http.Client
}

var Client = HTTPClient{Client: &http.Client{}}

func (client *HTTPClient) GetTo(path string, dest interface{}) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", endpointURL, path), nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request to `%s`: %w", path, err)
	}
	return client.DoJSON(req, dest)
}

func (client *HTTPClient) Delete(path string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", endpointURL, path), nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request to `%s`: %w", path, err)
	}
	return client.DoJSON(req, nil)
}

func (client *HTTPClient) PostTo(path string, payload interface{}, dest interface{}) error {
	bs, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload to `%s`: %v", path, err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", endpointURL, path), bytes.NewReader(bs))
	if err != nil {
		return fmt.Errorf("failed to create POST request to `%s`: %v", path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	return client.DoJSON(req, dest)
}

func (client *HTTPClient) DoJSON(req *http.Request, dest interface{}) error {
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request to `%s`: %w", req.URL, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from `%s`: %w", req.URL, err)
	}

	if res.StatusCode != http.StatusOK {
		var errResp core.ErrorResponse
		err := json.Unmarshal(body, &errResp)
		if err != nil {
			return fmt.Errorf("failed to parse error body `%s` from `%s`: %w", body, req.URL, err)
		}
		return fmt.Errorf("%s\n\t%s", errResp.StatusText, errResp.Msg)
	}

	err = json.Unmarshal(body, dest)
	if err != nil {
		return fmt.Errorf("failed to parse response body from `%s`: %w", req.URL, err)
	}
	return nil
}
