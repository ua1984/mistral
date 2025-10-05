// Package mistral provides a Go client library for the Mistral AI API.
//
// This library allows you to easily integrate Mistral AI's powerful language models
// into your Go applications. It supports all major features of the Mistral AI API:
//
//   - Chat completions (both standard and streaming)
//   - Text embeddings for semantic search and similarity
//   - File management for fine-tuning and batch processing
//   - Model information and management
//   - Function/tool calling for building AI agents
//
// # Getting Started
//
// First, create a client with your API key:
//
//	client := mistral.NewClient("your-api-key-here")
//
// Then use the client methods to interact with the API:
//
//	resp, err := client.CreateChatCompletion(ctx, &mistral.ChatCompletionRequest{
//	    Model: "mistral-large-latest",
//	    Messages: []mistral.ChatMessage{
//	        {Role: mistral.RoleUser, Content: "Hello!"},
//	    },
//	})
//
// # Authentication
//
// You need a Mistral AI API key to use this library. Get one at https://console.mistral.ai/
// Pass your API key when creating the client:
//
//	client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))
//
// # Configuration
//
// Customize the client with functional options:
//
//	client := mistral.NewClient(
//	    apiKey,
//	    mistral.WithTimeout(30*time.Second),
//	    mistral.WithBaseURL("https://api.custom-domain.com"),
//	)
//
// # Error Handling
//
// API errors are returned as *APIError, which includes the HTTP status code,
// error message, type, and code:
//
//	resp, err := client.CreateChatCompletion(ctx, req)
//	if err != nil {
//	    if apiErr, ok := err.(*mistral.APIError); ok {
//	        fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
//	    }
//	    return err
//	}
//
// # Examples
//
// See the example_test.go file for comprehensive examples of all features.
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

// Client represents a Mistral AI API client.
// This is the main entry point for interacting with the Mistral AI API.
// Create a new client using NewClient and then call its methods to access
// chat completions, embeddings, file management, and model operations.
//
// The client is safe for concurrent use by multiple goroutines.
type Client struct {
	// baseURL is the base URL for API requests (default: "https://api.mistral.ai").
	baseURL string

	// apiKey is your Mistral AI API key used for authentication.
	apiKey string

	// httpClient is the underlying HTTP client used for making requests.
	httpClient *http.Client
}

// NewClient creates a new Mistral AI API client with the provided API key.
// You can customize the client behavior by passing functional options.
//
// Parameters:
//   - apiKey: Your Mistral AI API key (required). Get one from https://console.mistral.ai/
//   - opts: Optional configuration functions to customize the client (see WithBaseURL,
//     WithHTTPClient, WithTimeout)
//
// Returns:
//   - A configured Client ready to make API requests
//
// Example:
//
//	client := mistral.NewClient("your-api-key-here")
//	// Or with options:
//	client := mistral.NewClient(
//	    "your-api-key-here",
//	    mistral.WithTimeout(30*time.Second),
//	)
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

// doRequest performs an HTTP request to the Mistral API and handles the response.
// This is an internal helper method used by all public API methods to make HTTP calls
// with consistent error handling, authentication, and JSON encoding/decoding.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - method: HTTP method (GET, POST, DELETE, etc.)
//   - path: API endpoint path (e.g., "/v1/chat/completions")
//   - body: Request body to be JSON-encoded, or nil for no body
//   - result: Pointer to a struct where the JSON response will be decoded, or nil to discard response
//
// Returns:
//   - An error if the request fails, the response status is not 2xx, or JSON decoding fails.
//     API errors are returned as *APIError with detailed error information
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

// handleErrorResponse processes error responses from the Mistral API.
// It attempts to parse the error response body as an APIError struct for detailed
// error information. If parsing fails, it returns a generic APIError with the raw response.
//
// Parameters:
//   - resp: The HTTP response with a non-2xx status code
//
// Returns:
//   - An *APIError containing the HTTP status code, error message, type, and code
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

// CreateChatCompletion creates a chat completion using the Mistral AI chat API.
// This method sends a conversation history to the model and receives a complete response.
// For streaming responses, use CreateChatCompletionStream instead.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - req: The chat completion request containing model, messages, and generation parameters
//
// Returns:
//   - A ChatCompletionResponse with the generated completion and metadata, or an error if the request fails
//
// Example:
//
//	resp, err := client.CreateChatCompletion(ctx, &mistral.ChatCompletionRequest{
//	    Model: "mistral-large-latest",
//	    Messages: []mistral.ChatMessage{
//	        {Role: mistral.RoleUser, Content: "Hello!"},
//	    },
//	})
func (c *Client) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	var resp ChatCompletionResponse
	if err := c.doRequest(ctx, http.MethodPost, "/v1/chat/completions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateChatCompletionStream creates a streaming chat completion.
// This method returns two channels: one for receiving response chunks as they're generated,
// and one for errors. The chunks arrive incrementally, allowing you to display partial
// responses to users as they're generated. This automatically sets req.Stream to true.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - req: The chat completion request. The Stream field will be set to true automatically
//
// Returns:
//   - A channel that receives ChatCompletionStreamResponse chunks as they arrive
//   - A channel that receives at most one error (or nil if the stream completes successfully)
//
// Both channels are closed when the stream ends or an error occurs.
//
// Example:
//
//	respChan, errChan := client.CreateChatCompletionStream(ctx, &mistral.ChatCompletionRequest{
//	    Model: "mistral-large-latest",
//	    Messages: []mistral.ChatMessage{
//	        {Role: mistral.RoleUser, Content: "Tell me a story"},
//	    },
//	})
//	for chunk := range respChan {
//	    // Process each chunk as it arrives
//	    fmt.Print(chunk.Choices[0].Delta.Content)
//	}
//	if err := <-errChan; err != nil {
//	    // Handle error
//	}
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

// streamChatCompletion is an internal helper that handles the streaming chat completion logic.
// It creates the HTTP request, processes server-sent events, and sends chunks to the response channel.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - req: The chat completion request (with Stream set to true)
//   - respChan: Channel to send response chunks to
//
// Returns:
//   - An error if the stream fails, or nil if it completes successfully
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

// CreateEmbedding creates embeddings for the given input texts.
// Embeddings are dense vector representations of text that capture semantic meaning,
// useful for similarity search, clustering, classification, and other ML tasks.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - req: The embedding request containing the model and input texts
//
// Returns:
//   - An EmbeddingResponse containing embedding vectors for each input text, or an error
//
// Example:
//
//	resp, err := client.CreateEmbedding(ctx, &mistral.EmbeddingRequest{
//	    Model: "mistral-embed",
//	    Input: []string{"Hello world", "Goodbye world"},
//	})
func (c *Client) CreateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	var resp EmbeddingResponse
	if err := c.doRequest(ctx, http.MethodPost, "/v1/embeddings", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UploadFile uploads a file to the Mistral API for use in fine-tuning or batch processing.
// The file content is sent as multipart form data along with metadata.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - req: The upload request containing the file content, filename, and purpose
//
// Returns:
//   - A File object with metadata about the uploaded file (including its ID for future reference),
//     or an error if the upload fails
//
// Example:
//
//	file, err := os.Open("training_data.jsonl")
//	if err != nil {
//	    return err
//	}
//	defer file.Close()
//
//	uploadedFile, err := client.UploadFile(ctx, &mistral.UploadFileRequest{
//	    File:     file,
//	    Filename: "training_data.jsonl",
//	    Purpose:  mistral.FilePurposeFineTune,
//	})
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

// ListFiles retrieves a paginated list of files that have been uploaded to your account.
// You can filter by purpose, search by filename, and control pagination.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - params: Optional filtering and pagination parameters. Pass nil to use defaults
//
// Returns:
//   - A FileList containing an array of File objects and pagination metadata,
//     or an error if the request fails
//
// Example:
//
//	// List all files
//	files, err := client.ListFiles(ctx, nil)
//
//	// List only fine-tuning files with pagination
//	files, err := client.ListFiles(ctx, &mistral.ListFilesParams{
//	    Purpose:  mistral.FilePurposeFineTune,
//	    Page:     1,
//	    PageSize: 20,
//	})
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

// GetFile retrieves metadata about a specific file by its ID.
// This returns information about the file but not its content. Use DownloadFile
// to retrieve the actual file content.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - fileID: The unique identifier of the file to retrieve
//
// Returns:
//   - A File object containing metadata about the file, or an error if the file
//     doesn't exist or the request fails
//
// Example:
//
//	file, err := client.GetFile(ctx, "file-abc123")
func (c *Client) GetFile(ctx context.Context, fileID string) (*File, error) {
	var resp File
	path := fmt.Sprintf("/v1/files/%s", fileID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteFile deletes a file from your account.
// This permanently removes the file and it cannot be recovered. The file ID will
// no longer be valid for any operations.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - fileID: The unique identifier of the file to delete
//
// Returns:
//   - A DeleteFileResponse confirming the deletion, or an error if the file
//     doesn't exist or the request fails
//
// Example:
//
//	resp, err := client.DeleteFile(ctx, "file-abc123")
//	if err != nil {
//	    return err
//	}
//	if resp.Deleted {
//	    fmt.Println("File successfully deleted")
//	}
func (c *Client) DeleteFile(ctx context.Context, fileID string) (*DeleteFileResponse, error) {
	var resp DeleteFileResponse
	path := fmt.Sprintf("/v1/files/%s", fileID)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadFile downloads the actual content of a file.
// This returns an io.ReadCloser that streams the file content. You are responsible
// for closing the reader when done to free resources.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - fileID: The unique identifier of the file to download
//
// Returns:
//   - An io.ReadCloser that streams the file content. Must be closed by the caller.
//     Returns an error if the file doesn't exist or the request fails
//
// Example:
//
//	reader, err := client.DownloadFile(ctx, "file-abc123")
//	if err != nil {
//	    return err
//	}
//	defer reader.Close()
//
//	content, err := io.ReadAll(reader)
//	if err != nil {
//	    return err
//	}
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

// ListModels retrieves a list of all available models.
// This includes both base models provided by Mistral AI and any fine-tuned models
// in your account. The response includes metadata about each model's capabilities
// and limitations.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//
// Returns:
//   - A ModelList containing an array of Model objects, or an error if the request fails
//
// Example:
//
//	models, err := client.ListModels(ctx)
//	if err != nil {
//	    return err
//	}
//	for _, model := range models.Data {
//	    fmt.Printf("Model: %s, Max Tokens: %d\n", model.ID, model.MaxTokens)
//	}
func (c *Client) ListModels(ctx context.Context) (*ModelList, error) {
	var resp ModelList
	if err := c.doRequest(ctx, http.MethodGet, "/v1/models", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetModel retrieves detailed information about a specific model by its ID.
// This is useful for checking a model's capabilities, token limits, and other
// metadata before using it.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - modelID: The unique identifier of the model (e.g., "mistral-large-latest")
//
// Returns:
//   - A Model object containing detailed metadata, or an error if the model
//     doesn't exist or the request fails
//
// Example:
//
//	model, err := client.GetModel(ctx, "mistral-large-latest")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Max tokens: %d\n", model.MaxTokens)
func (c *Client) GetModel(ctx context.Context, modelID string) (*Model, error) {
	var resp Model
	path := fmt.Sprintf("/v1/models/%s", modelID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteModel deletes a fine-tuned model from your account.
// This only works for custom fine-tuned models; you cannot delete base models
// provided by Mistral AI. Once deleted, the model cannot be recovered and will
// no longer be usable for completions.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - modelID: The unique identifier of the fine-tuned model to delete
//
// Returns:
//   - A DeleteModelResponse confirming the deletion, or an error if the model
//     doesn't exist, is not a fine-tuned model, or the request fails
//
// Example:
//
//	resp, err := client.DeleteModel(ctx, "my-fine-tuned-model")
//	if err != nil {
//	    return err
//	}
//	if resp.Deleted {
//	    fmt.Println("Model successfully deleted")
//	}
func (c *Client) DeleteModel(ctx context.Context, modelID string) (*DeleteModelResponse, error) {
	var resp DeleteModelResponse
	path := fmt.Sprintf("/v1/models/%s", modelID)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
