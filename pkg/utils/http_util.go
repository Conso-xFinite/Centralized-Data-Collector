package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	HttpMethodPost = "POST"
	HttpMethodGet  = "GET"
)

func HttpRequest(ctx context.Context, method string, url string, body []byte, headers map[string]string) (httpCode int, respBody []byte, err error) {
	var reader io.Reader
	switch method {
	default:
		return 0, nil, fmt.Errorf("unsupported HTTP method: %s", method)
	case HttpMethodGet:
		reader = nil
	case HttpMethodPost:
		if len(body) == 0 {
			return 0, nil, fmt.Errorf("request body cannot be empty")
		} else {
			reader = bytes.NewBuffer(body)
		}
	}

	// logger.Info("%s URL: %s, body: %v, bodyBuffer: %v", method, url, body, reader)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %v", err)
	}

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

// HttpPostWithPB 发起HTTP POST请求
func HttpPostWithPB(ctx context.Context, url string, reqMsg proto.Message, respMsg proto.Message, headers map[string]string) (int, error) {
	if reqMsg == nil || respMsg == nil {
		return 0, fmt.Errorf("request and response messages cannot be nil")
	}
	body, err := proto.Marshal(reqMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request message: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, HttpMethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	} else if len(respBody) == 0 {
		return resp.StatusCode, fmt.Errorf("empty response body")
	} else if err := proto.Unmarshal(respBody, respMsg); err != nil {
		return resp.StatusCode, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return resp.StatusCode, nil
}

func HttpPostWithBody(ctx context.Context, url string, reqBody []byte, headers map[string]string) (int, []byte, error) {
	if len(reqBody) == 0 {
		return 0, nil, fmt.Errorf("request body cannot be empty")
	}
	req, err := http.NewRequestWithContext(ctx, HttpMethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %v", err)
	}

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
	} else if len(respBody) == 0 {
		return resp.StatusCode, nil, fmt.Errorf("empty response body")
	}

	return resp.StatusCode, respBody, nil
}

func HttpPostWithJsonMap(ctx context.Context, url string, reqJsonMap map[string]interface{}, headers map[string]string) (int, map[string]interface{}, error) {
	if len(reqJsonMap) == 0 {
		return 0, nil, fmt.Errorf("request JSON map cannot be empty")
	}
	body, err := json.Marshal(reqJsonMap)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal request JSON map: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, HttpMethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %v", err)
	}

	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
	} else if len(respBody) == 0 {
		return resp.StatusCode, nil, fmt.Errorf("empty response body")
	}

	var respJsonMap map[string]interface{}
	if err := json.Unmarshal(respBody, &respJsonMap); err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return resp.StatusCode, respJsonMap, nil
}

func HttpPostWithJsonStruct(ctx context.Context, url string, reqStruct interface{}, respStruct interface{}) (int, error) {
	if reqStruct == nil || respStruct == nil {
		return 0, fmt.Errorf("request and response structs cannot be nil")
	}
	body, err := json.Marshal(reqStruct)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request struct: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, HttpMethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	} else if len(respBody) == 0 {
		return resp.StatusCode, fmt.Errorf("empty response body")
	} else if err := json.Unmarshal(respBody, respStruct); err != nil {
		return resp.StatusCode, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return resp.StatusCode, nil
}

type Option struct {
	Hostname string
	Path     string
	Method   string
	Headers  map[string]string
}

// func HttpRequestWithJsonMap(ctx context.Context, option Option, reqJsonMap map[string]interface{}) (int, map[string]interface{}, error) {
// 	if len(reqJsonMap) == 0 {
// 		return 0, nil, fmt.Errorf("request JSON map cannot be empty")
// 	}
// 	body, err := json.Marshal(reqJsonMap)
// 	if err != nil {
// 		return 0, nil, fmt.Errorf("failed to marshal request JSON map: %v", err)
// 	}

// 	req, err := http.NewRequestWithContext(ctx, option.Method, fmt.Sprintf("https://%s%s", option.Hostname, option.Path), bytes.NewBuffer(body))
// 	if err != nil {
// 		return 0, nil, fmt.Errorf("failed to create request: %v", err)
// 	}
// 	if len(option.Headers) != 0 {
// 		for k, v := range option.Headers {
// 			req.Header.Set(k, v)
// 		}
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{Timeout: 30 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return 0, nil, fmt.Errorf("request failed: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
// 	} else if len(respBody) == 0 {
// 		return resp.StatusCode, nil, fmt.Errorf("empty response body")
// 	}

// 	var respJsonMap map[string]interface{}
// 	if err := json.Unmarshal(respBody, &respJsonMap); err != nil {
// 		return resp.StatusCode, nil, fmt.Errorf("failed to unmarshal response body: %w", err)
// 	}

// 	return resp.StatusCode, respJsonMap, nil
// }
