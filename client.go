package mistral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.mistral.ai"
	defaultTimeout = 60 * time.Second
)

// Client represents a Mistral AI API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Mistral AI API client
func NewClient(apiKey string, opts ...Option) *Client {
	client := &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// doRequest performs an HTTP request and handles the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleErrorResponse(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// handleErrorResponse processes error responses from the API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiError APIError
	if err := json.Unmarshal(body, &apiError); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	apiError.StatusCode = resp.StatusCode
	return &apiError
}

// CreateChatCompletion creates a chat completion
func (c *Client) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	var resp ChatCompletionResponse
	if err := c.doRequest(ctx, http.MethodPost, "/v1/chat/completions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *Client) CreateChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan ChatCompletionStreamResponse, <-chan error) {
	req.Stream = true

	respChan := make(chan ChatCompletionStreamResponse)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		if err := c.streamChatCompletion(ctx, req, respChan); err != nil {
			errChan <- err
		}
	}()

	return respChan, errChan
}

func (c *Client) streamChatCompletion(ctx context.Context, req *ChatCompletionRequest, respChan chan<- ChatCompletionStreamResponse) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(httpResp)
	}

	scanner := bufio.NewScanner(httpResp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("failed to unmarshal stream chunk: %w", err)
		}

		select {
		case respChan <- chunk:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// CreateEmbedding creates embeddings for the given input
func (c *Client) CreateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	var resp EmbeddingResponse
	if err := c.doRequest(ctx, http.MethodPost, "/v1/embeddings", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UploadFile uploads a file to the Mistral API
func (c *Client) UploadFile(ctx context.Context, req *UploadFileRequest) (*File, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add purpose field
	if req.Purpose != "" {
		if err := writer.WriteField("purpose", string(req.Purpose)); err != nil {
			return nil, fmt.Errorf("failed to write purpose field: %w", err)
		}
	}

	// Add file field
	part, err := writer.CreateFormFile("file", req.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, req.File); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/files", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.handleErrorResponse(resp)
	}

	var file File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &file, nil
}

// ListFiles retrieves a list of files
func (c *Client) ListFiles(ctx context.Context, params *ListFilesParams) (*FileList, error) {
	path := "/v1/files"
	if params != nil {
		path += "?"
		if params.Page > 0 {
			path += fmt.Sprintf("page=%d&", params.Page)
		}
		if params.PageSize > 0 {
			path += fmt.Sprintf("page_size=%d&", params.PageSize)
		}
		if params.Purpose != "" {
			path += fmt.Sprintf("purpose=%s&", params.Purpose)
		}
		if params.Search != "" {
			path += fmt.Sprintf("search=%s&", params.Search)
		}
	}

	var resp FileList
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFile retrieves information about a specific file
func (c *Client) GetFile(ctx context.Context, fileID string) (*File, error) {
	var resp File
	path := fmt.Sprintf("/v1/files/%s", fileID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteFile deletes a file
func (c *Client) DeleteFile(ctx context.Context, fileID string) (*DeleteFileResponse, error) {
	var resp DeleteFileResponse
	path := fmt.Sprintf("/v1/files/%s", fileID)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadFile downloads the content of a file
func (c *Client) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	path := fmt.Sprintf("/v1/files/%s/content", fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, c.handleErrorResponse(resp)
	}

	return resp.Body, nil
}

// ListModels retrieves a list of available models
func (c *Client) ListModels(ctx context.Context) (*ModelList, error) {
	var resp ModelList
	if err := c.doRequest(ctx, http.MethodGet, "/v1/models", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetModel retrieves information about a specific model
func (c *Client) GetModel(ctx context.Context, modelID string) (*Model, error) {
	var resp Model
	path := fmt.Sprintf("/v1/models/%s", modelID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteModel deletes a fine-tuned model
func (c *Client) DeleteModel(ctx context.Context, modelID string) (*DeleteModelResponse, error) {
	var resp DeleteModelResponse
	path := fmt.Sprintf("/v1/models/%s", modelID)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
