// Package httputil provides utility functions for HTTP operations.
package httputil

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient represents an HTTP client with configuration
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// ClientOption defines a function type for configuring HTTPClient
type ClientOption func(*HTTPClient)

// NewClient creates a new HTTP client with options
func NewClient(options ...ClientOption) *HTTPClient {
	client := &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// WithTimeout sets the timeout for HTTP requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *HTTPClient) {
		c.client.Timeout = timeout
	}
}

// WithBaseURL sets the base URL for HTTP requests
func WithBaseURL(baseURL string) ClientOption {
	return func(c *HTTPClient) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithHeader adds a default header to all requests
func WithHeader(key, value string) ClientOption {
	return func(c *HTTPClient) {
		c.headers[key] = value
	}
}

// WithUserAgent sets the User-Agent header
func WithUserAgent(userAgent string) ClientOption {
	return func(c *HTTPClient) {
		c.headers["User-Agent"] = userAgent
	}
}

// WithBearerToken sets the Authorization header with Bearer token
func WithBearerToken(token string) ClientOption {
	return func(c *HTTPClient) {
		c.headers["Authorization"] = "Bearer " + token
	}
}

// WithBasicAuth sets the Authorization header with Basic auth
func WithBasicAuth(username, password string) ClientOption {
	return func(c *HTTPClient) {
		auth := username + ":" + password
		encoded := "Basic " + encodeBase64([]byte(auth))
		c.headers["Authorization"] = encoded
	}
}

// Request represents an HTTP request configuration
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        interface{}
	ContentType string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	Request    *http.Request
}

// NewRequest creates a new request
func NewRequest(method, url string) *Request {
	return &Request{
		Method:      method,
		URL:         url,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		ContentType: "application/json",
	}
}

// WithHeader adds a header to the request
func (r *Request) WithHeader(key, value string) *Request {
	r.Headers[key] = value
	return r
}

// WithQuery adds a query parameter to the request
func (r *Request) WithQuery(key, value string) *Request {
	r.QueryParams[key] = value
	return r
}

// WithBody sets the request body
func (r *Request) WithBody(body interface{}) *Request {
	r.Body = body
	return r
}

// WithJSON sets the request body as JSON and content type
func (r *Request) WithJSON(body interface{}) *Request {
	r.Body = body
	r.ContentType = "application/json"
	return r
}

// WithForm sets the request body as form data
func (r *Request) WithForm(data map[string]string) *Request {
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}
	r.Body = values.Encode()
	r.ContentType = "application/x-www-form-urlencoded"
	return r
}

// Do executes the HTTP request
func (c *HTTPClient) Do(ctx context.Context, req *Request) (*Response, error) {
	// Build URL
	requestURL := req.URL
	if c.baseURL != "" && !strings.HasPrefix(req.URL, "http") {
		requestURL = c.baseURL + "/" + strings.TrimPrefix(req.URL, "/")
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		u, err := url.Parse(requestURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		query := u.Query()
		for key, value := range req.QueryParams {
			query.Set(key, value)
		}
		u.RawQuery = query.Encode()
		requestURL = u.String()
	}

	// Prepare body
	var body io.Reader
	if req.Body != nil {
		switch v := req.Body.(type) {
		case string:
			body = strings.NewReader(v)
		case []byte:
			body = bytes.NewReader(v)
		case io.Reader:
			body = v
		default:
			// Try to marshal as JSON
			jsonData, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			body = bytes.NewReader(jsonData)
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	// Set request headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type
	if req.Body != nil && req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	// Execute request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
		Request:    httpReq,
	}, nil
}

// Get performs a GET request
func (c *HTTPClient) Get(ctx context.Context, url string) (*Response, error) {
	req := NewRequest("GET", url)
	return c.Do(ctx, req)
}

// Post performs a POST request with JSON body
func (c *HTTPClient) Post(ctx context.Context, url string, body interface{}) (*Response, error) {
	req := NewRequest("POST", url).WithJSON(body)
	return c.Do(ctx, req)
}

// Put performs a PUT request with JSON body
func (c *HTTPClient) Put(ctx context.Context, url string, body interface{}) (*Response, error) {
	req := NewRequest("PUT", url).WithJSON(body)
	return c.Do(ctx, req)
}

// Patch performs a PATCH request with JSON body
func (c *HTTPClient) Patch(ctx context.Context, url string, body interface{}) (*Response, error) {
	req := NewRequest("PATCH", url).WithJSON(body)
	return c.Do(ctx, req)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(ctx context.Context, url string) (*Response, error) {
	req := NewRequest("DELETE", url)
	return c.Do(ctx, req)
}

// IsSuccess checks if the response status code indicates success (2xx)
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError checks if the response status code indicates client error (4xx)
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError checks if the response status code indicates server error (5xx)
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// String returns the response body as string
func (r *Response) String() string {
	return string(r.Body)
}

// JSON unmarshals the response body as JSON
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// GetHeader gets a header value from the response
func (r *Response) GetHeader(key string) string {
	values := r.Headers[key]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// Simple HTTP functions for quick operations

// Get performs a simple GET request
func Get(ctx context.Context, url string) (*Response, error) {
	client := NewClient()
	return client.Get(ctx, url)
}

// Post performs a simple POST request with JSON body
func Post(ctx context.Context, url string, body interface{}) (*Response, error) {
	client := NewClient()
	return client.Post(ctx, url, body)
}

// Put performs a simple PUT request with JSON body
func Put(ctx context.Context, url string, body interface{}) (*Response, error) {
	client := NewClient()
	return client.Put(ctx, url, body)
}

// Patch performs a simple PATCH request with JSON body
func Patch(ctx context.Context, url string, body interface{}) (*Response, error) {
	client := NewClient()
	return client.Patch(ctx, url, body)
}

// Delete performs a simple DELETE request
func Delete(ctx context.Context, url string) (*Response, error) {
	client := NewClient()
	return client.Delete(ctx, url)
}

// DownloadFile downloads a file from URL and returns the content
func DownloadFile(ctx context.Context, url string) ([]byte, error) {
	resp, err := Get(ctx, url)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// UploadFile uploads a file to the specified URL
func UploadFile(ctx context.Context, url string, fieldName string, fileContent []byte, fileName string) (*Response, error) {
	body := &bytes.Buffer{}

	// Create multipart form
	boundary := "----MyLibFormBoundary"

	// Write file field
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n", fieldName, fileName))
	body.WriteString("Content-Type: application/octet-stream\r\n\r\n")
	body.Write(fileContent)
	body.WriteString("\r\n--" + boundary + "--\r\n")

	req := NewRequest("POST", url).
		WithHeader("Content-Type", "multipart/form-data; boundary="+boundary).
		WithBody(body.Bytes())

	client := NewClient()
	return client.Do(ctx, req)
}

// BuildURL builds URL with query parameters
func BuildURL(baseURL string, params map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()

	return u.String(), nil
}

// ParseURL parses URL and returns components
func ParseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// IsValidURL checks if the string is a valid URL
func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// GetQueryParam gets a query parameter value from URL
func GetQueryParam(rawURL, param string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return u.Query().Get(param), nil
}

// SetQueryParam sets a query parameter in URL
func SetQueryParam(rawURL, param, value string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set(param, value)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

// URLEncode encodes string for URL
func URLEncode(s string) string {
	return url.QueryEscape(s)
}

// URLDecode decodes URL encoded string
func URLDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// Helper function for base64 encoding (used in BasicAuth)
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// ResponseError represents an HTTP error response
type ResponseError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *ResponseError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// CheckResponse checks if response is successful, returns error if not
func CheckResponse(resp *Response) error {
	if resp.IsSuccess() {
		return nil
	}

	return &ResponseError{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Message:    string(resp.Body),
	}
}

// RetryConfig defines retry configuration
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
	BackoffFn  func(int) time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		Delay:      time.Second,
		BackoffFn: func(attempt int) time.Duration {
			return time.Duration(attempt) * time.Second
		},
	}
}

// DoWithRetry executes HTTP request with retry logic
func (c *HTTPClient) DoWithRetry(ctx context.Context, req *Request, config *RetryConfig) (*Response, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		resp, err := c.Do(ctx, req)

		// Success case
		if err == nil && resp.IsSuccess() {
			return resp, nil
		}

		// Store the error for potential return
		lastErr = err
		if err == nil {
			lastErr = CheckResponse(resp)
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Don't retry on client errors (4xx)
		if resp != nil && resp.IsClientError() {
			break
		}

		// Calculate delay
		delay := config.Delay
		if config.BackoffFn != nil {
			delay = config.BackoffFn(attempt + 1)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return nil, lastErr
}
