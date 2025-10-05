package mistral

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	assert.Equal(t, defaultBaseURL, client.baseURL)
	assert.Equal(t, apiKey, client.apiKey)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, defaultTimeout, client.httpClient.Timeout)
}

func TestNewClientWithOptions(t *testing.T) {
	apiKey := "test-api-key"
	customURL := "https://custom.api.example.com"
	customTimeout := 30 * time.Second

	client := NewClient(apiKey,
		WithBaseURL(customURL),
		WithTimeout(customTimeout),
	)

	assert.Equal(t, customURL, client.baseURL)
	assert.Equal(t, apiKey, client.apiKey)
	assert.Equal(t, customTimeout, client.httpClient.Timeout)
}

func TestCreateChatCompletion(t *testing.T) {
	expectedResponse := ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mistral-large-latest",
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    RoleAssistant,
					Content: "The capital of France is Paris.",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "mistral-large-latest", req.Model)
		assert.Len(t, req.Messages, 1)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.CreateChatCompletion(context.Background(), &ChatCompletionRequest{
		Model: "mistral-large-latest",
		Messages: []ChatMessage{
			{
				Role:    RoleUser,
				Content: "What is the capital of France?",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedResponse.ID, resp.ID)
	assert.Equal(t, expectedResponse.Model, resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "The capital of France is Paris.", resp.Choices[0].Message.Content)
}

func TestCreateChatCompletionStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))

		// Verify request has stream: true
		var req ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.Stream)

		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		// Send streaming chunks
		chunks := []ChatCompletionStreamResponse{
			{
				ID:      "test-id",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "mistral-large-latest",
				Choices: []ChatCompletionChoice{
					{
						Index: 0,
						Delta: &ChatMessage{
							Role:    RoleAssistant,
							Content: "Hello",
						},
					},
				},
			},
			{
				ID:      "test-id",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "mistral-large-latest",
				Choices: []ChatCompletionChoice{
					{
						Index: 0,
						Delta: &ChatMessage{
							Content: " world",
						},
					},
				},
			},
		}

		for _, chunk := range chunks {
			data, _ := json.Marshal(chunk)
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}

		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	respChan, errChan := client.CreateChatCompletionStream(context.Background(), &ChatCompletionRequest{
		Model: "mistral-large-latest",
		Messages: []ChatMessage{
			{
				Role:    RoleUser,
				Content: "Say hello",
			},
		},
	})

	var chunks []ChatCompletionStreamResponse
	for {
		select {
		case chunk, ok := <-respChan:
			if !ok {
				// Channel closed, we're done
				assert.Len(t, chunks, 2)
				assert.Equal(t, "Hello", chunks[0].Choices[0].Delta.Content)
				assert.Equal(t, " world", chunks[1].Choices[0].Delta.Content)
				return
			}
			chunks = append(chunks, chunk)
		case err := <-errChan:
			require.NoError(t, err)
			return
		}
	}
}

func TestCreateEmbedding(t *testing.T) {
	expectedResponse := EmbeddingResponse{
		ID:     "test-id",
		Object: "list",
		Model:  "mistral-embed",
		Data: []EmbeddingObject{
			{
				Object:    "embedding",
				Embedding: []float64{0.1, 0.2, 0.3},
				Index:     0,
			},
			{
				Object:    "embedding",
				Embedding: []float64{0.4, 0.5, 0.6},
				Index:     1,
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 0,
			TotalTokens:      10,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/embeddings", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		var req EmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "mistral-embed", req.Model)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.CreateEmbedding(context.Background(), &EmbeddingRequest{
		Model: "mistral-embed",
		Input: []string{"Hello", "World"},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedResponse.ID, resp.ID)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, resp.Data[0].Embedding)
}

func TestUploadFile(t *testing.T) {
	expectedFile := File{
		ID:        "file-123",
		Object:    "file",
		Bytes:     1024,
		CreatedAt: time.Now(),
		Filename:  "test.jsonl",
		Purpose:   "fine-tune",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/files", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		assert.Equal(t, "fine-tune", r.FormValue("purpose"))

		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()
		assert.Equal(t, "test.jsonl", header.Filename)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedFile)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	fileContent := strings.NewReader(`{"text": "test"}`)
	resp, err := client.UploadFile(context.Background(), &UploadFileRequest{
		File:     fileContent,
		Filename: "test.jsonl",
		Purpose:  FilePurposeFineTune,
	})

	require.NoError(t, err)
	assert.Equal(t, expectedFile.ID, resp.ID)
	assert.Equal(t, expectedFile.Filename, resp.Filename)
}

func TestListFiles(t *testing.T) {
	expectedResponse := FileList{
		Object: "list",
		Data: []File{
			{
				ID:        "file-1",
				Object:    "file",
				Bytes:     1024,
				CreatedAt: time.Now(),
				Filename:  "test1.jsonl",
				Purpose:   "fine-tune",
			},
			{
				ID:        "file-2",
				Object:    "file",
				Bytes:     2048,
				CreatedAt: time.Now(),
				Filename:  "test2.jsonl",
				Purpose:   "fine-tune",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/files", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.ListFiles(context.Background(), nil)

	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "file-1", resp.Data[0].ID)
}

func TestListFilesWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.RawQuery, "page=2")
		assert.Contains(t, r.URL.RawQuery, "page_size=10")
		assert.Contains(t, r.URL.RawQuery, "purpose=fine-tune")

		resp := FileList{Object: "list", Data: []File{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	_, err := client.ListFiles(context.Background(), &ListFilesParams{
		Page:     2,
		PageSize: 10,
		Purpose:  "fine-tune",
	})

	require.NoError(t, err)
}

func TestGetFile(t *testing.T) {
	expectedFile := File{
		ID:        "file-123",
		Object:    "file",
		Bytes:     1024,
		CreatedAt: time.Now(),
		Filename:  "test.jsonl",
		Purpose:   "fine-tune",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/files/file-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedFile)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.GetFile(context.Background(), "file-123")

	require.NoError(t, err)
	assert.Equal(t, expectedFile.ID, resp.ID)
	assert.Equal(t, expectedFile.Filename, resp.Filename)
}

func TestDeleteFile(t *testing.T) {
	expectedResponse := DeleteFileResponse{
		ID:      "file-123",
		Object:  "file",
		Deleted: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/files/file-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.DeleteFile(context.Background(), "file-123")

	require.NoError(t, err)
	assert.Equal(t, "file-123", resp.ID)
	assert.True(t, resp.Deleted)
}

func TestDownloadFile(t *testing.T) {
	expectedContent := "file content here"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/files/file-123/content", r.URL.Path)

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.DownloadFile(context.Background(), "file-123")
	require.NoError(t, err)
	defer resp.Close()

	content, err := io.ReadAll(resp)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))
}

func TestListModels(t *testing.T) {
	expectedResponse := ModelList{
		Object: "list",
		Data: []Model{
			{
				ID:      "mistral-large-latest",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "mistralai",
			},
			{
				ID:      "mistral-small-latest",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "mistralai",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/models", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.ListModels(context.Background())

	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "mistral-large-latest", resp.Data[0].ID)
}

func TestGetModel(t *testing.T) {
	expectedModel := Model{
		ID:      "mistral-large-latest",
		Object:  "model",
		Created: time.Now().Unix(),
		OwnedBy: "mistralai",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/models/mistral-large-latest", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedModel)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.GetModel(context.Background(), "mistral-large-latest")

	require.NoError(t, err)
	assert.Equal(t, expectedModel.ID, resp.ID)
	assert.Equal(t, expectedModel.OwnedBy, resp.OwnedBy)
}

func TestDeleteModel(t *testing.T) {
	expectedResponse := DeleteModelResponse{
		ID:      "ft:mistral-small:custom",
		Object:  "model",
		Deleted: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/models/ft:mistral-small:custom", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	resp, err := client.DeleteModel(context.Background(), "ft:mistral-small:custom")

	require.NoError(t, err)
	assert.Equal(t, "ft:mistral-small:custom", resp.ID)
	assert.True(t, resp.Deleted)
}

func TestHandleErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   interface{}
		expectedError  string
		expectedStatus int
	}{
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			responseBody: map[string]interface{}{
				"message": "Invalid API key",
				"type":    "authentication_error",
			},
			expectedError:  "authentication_error: Invalid API key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "429 Rate Limit",
			statusCode: http.StatusTooManyRequests,
			responseBody: map[string]interface{}{
				"message": "Rate limit exceeded",
				"type":    "rate_limit_error",
			},
			expectedError:  "rate_limit_error: Rate limit exceeded",
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "500 Server Error with non-JSON",
			statusCode:     http.StatusInternalServerError,
			responseBody:   "Internal Server Error",
			expectedError:  "Internal Server Error",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if str, ok := tt.responseBody.(string); ok {
					w.Write([]byte(str))
				} else {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient("test-api-key", WithBaseURL(server.URL))

			_, err := client.CreateChatCompletion(context.Background(), &ChatCompletionRequest{
				Model: "mistral-large-latest",
				Messages: []ChatMessage{
					{
						Role:    RoleUser,
						Content: "test",
					},
				},
			})

			require.Error(t, err)
			apiErr, ok := err.(*APIError)
			require.True(t, ok, "error should be of type *APIError")
			assert.Equal(t, tt.expectedStatus, apiErr.StatusCode)
			assert.Equal(t, tt.expectedError, apiErr.Error())
		})
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.CreateChatCompletion(ctx, &ChatCompletionRequest{
		Model: "mistral-large-latest",
		Messages: []ChatMessage{
			{
				Role:    RoleUser,
				Content: "test",
			},
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestStreamContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		// Send one chunk then sleep
		chunk := ChatCompletionStreamResponse{
			ID:      "test-id",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   "mistral-large-latest",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Delta: &ChatMessage{
						Content: "test",
					},
				},
			},
		}
		data, _ := json.Marshal(chunk)
		w.Write([]byte("data: "))
		w.Write(data)
		w.Write([]byte("\n\n"))
		flusher.Flush()

		// Sleep to allow context cancellation
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())

	respChan, errChan := client.CreateChatCompletionStream(ctx, &ChatCompletionRequest{
		Model: "mistral-large-latest",
		Messages: []ChatMessage{
			{
				Role:    RoleUser,
				Content: "test",
			},
		},
	})

	// Read first chunk
	select {
	case chunk := <-respChan:
		assert.Equal(t, "test", chunk.Choices[0].Delta.Content)
	case err := <-errChan:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for chunk")
	}

	// Cancel context
	cancel()

	// Should receive context cancellation error
	select {
	case err := <-errChan:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestDoRequestWithNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Empty(t, body)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result map[string]string
	err := client.doRequest(context.Background(), http.MethodGet, "/test", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestDoRequestWithNilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	err := client.doRequest(context.Background(), http.MethodDelete, "/test", nil, nil)

	require.NoError(t, err)
}

func TestAPIErrorError(t *testing.T) {
	tests := []struct {
		name     string
		apiError APIError
		expected string
	}{
		{
			name: "with type",
			apiError: APIError{
				StatusCode: 401,
				Type:       "authentication_error",
				Message:    "Invalid API key",
			},
			expected: "authentication_error: Invalid API key",
		},
		{
			name: "without type",
			apiError: APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expected: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.apiError.Error())
		})
	}
}

func TestUploadFileWithEmptyPurpose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		// Purpose should not be in form when empty
		assert.Empty(t, r.FormValue("purpose"))

		resp := File{ID: "file-123"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	_, err := client.UploadFile(context.Background(), &UploadFileRequest{
		File:     bytes.NewReader([]byte("test")),
		Filename: "test.txt",
		Purpose:  "", // Empty purpose
	})

	require.NoError(t, err)
}
