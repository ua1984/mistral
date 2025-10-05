# Mistral Go SDK


[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/ua1984/mistral/test.yaml?branch=master&amp;style=flat-square)](https://github.com/ua1984/mistral/actions?query=workflow%3Atest)
[![GoDoc](https://pkg.go.dev/badge/mod/github.com/ua1984/mistral)](https://pkg.go.dev/mod/github.com/ua1984/mistral)
[![Go Report Card](https://goreportcard.com/badge/github.com/ua1984/mistral)](https://goreportcard.com/report/github.com/ua1984/mistral)
[![Release](https://img.shields.io/github/release/ua1984/mistral.svg?style=flat-square)](https://github.com/ua1984/mistral/releases/latest)


A Go client library for the [Mistral AI API](https://docs.mistral.ai/).

## Installation

```bash
go get github.com/ua1984/mistral
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/ua1984/mistral"
)

func main() {
    client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))

    resp, err := client.CreateChatCompletion(context.Background(), &mistral.ChatCompletionRequest{
        Model: "mistral-large-latest",
        Messages: []mistral.ChatMessage{
            {
                Role:    mistral.RoleUser,
                Content: "What is the capital of France?",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Features

- **Chat Completions**: Create chat completions with support for streaming
- **Embeddings**: Generate embeddings for text inputs
- **File Management**: Upload, download, list, and delete files
- **Model Management**: List and retrieve model information
- **Streaming Support**: Real-time streaming responses for chat completions
- **Context Support**: Full `context.Context` support for all API calls
- **Custom Configuration**: Configurable base URL, HTTP client, and timeouts

## Usage Examples

### Chat Completion

```go
client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))

resp, err := client.CreateChatCompletion(context.Background(), &mistral.ChatCompletionRequest{
    Model: "mistral-large-latest",
    Messages: []mistral.ChatMessage{
        {
            Role:    mistral.RoleUser,
            Content: "What is the capital of France?",
        },
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

### Streaming Chat Completion

```go
client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))
ctx := context.Background()

respChan, errChan := client.CreateChatCompletionStream(ctx, &mistral.ChatCompletionRequest{
    Model: "mistral-large-latest",
    Messages: []mistral.ChatMessage{
        {
            Role:    mistral.RoleUser,
            Content: "Tell me a short story",
        },
    },
})

for {
    select {
    case chunk, ok := <-respChan:
        if !ok {
            return
        }
        if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
            fmt.Print(chunk.Choices[0].Delta.Content)
        }
    case err := <-errChan:
        if err != nil {
            log.Fatal(err)
        }
        return
    case <-ctx.Done():
        return
    }
}
```

### Embeddings

```go
client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))

resp, err := client.CreateEmbedding(context.Background(), &mistral.EmbeddingRequest{
    Model: "mistral-embed",
    Input: []string{
        "Hello, world!",
        "How are you?",
    },
})
if err != nil {
    log.Fatal(err)
}

for i, emb := range resp.Data {
    fmt.Printf("Embedding %d has %d dimensions\n", i, len(emb.Embedding))
}
```

### File Upload

```go
client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))

file, err := os.Open("training_data.jsonl")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

uploadedFile, err := client.UploadFile(context.Background(), &mistral.UploadFileRequest{
    File:     file,
    Filename: "training_data.jsonl",
    Purpose:  mistral.FilePurposeFineTune,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Uploaded file ID: %s\n", uploadedFile.ID)
```

### List Models

```go
client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))

models, err := client.ListModels(context.Background())
if err != nil {
    log.Fatal(err)
}

for _, model := range models.Data {
    fmt.Printf("Model: %s\n", model.ID)
}
```

## Configuration Options

The client supports various configuration options:

```go
client := mistral.NewClient(
    "your-api-key",
    mistral.WithBaseURL("https://custom.api.url"),
    mistral.WithTimeout(120 * time.Second),
    mistral.WithHTTPClient(customHTTPClient),
)
```

### Available Options

- `WithBaseURL(url string)`: Set a custom base URL for the API
- `WithTimeout(timeout time.Duration)`: Set the HTTP client timeout
- `WithHTTPClient(client *http.Client)`: Use a custom HTTP client

## API Reference

### Chat Completions

- `CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)`
- `CreateChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan ChatCompletionStreamResponse, <-chan error)`

### Embeddings

- `CreateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)`

### Files

- `UploadFile(ctx context.Context, req *UploadFileRequest) (*File, error)`
- `ListFiles(ctx context.Context, params *ListFilesParams) (*FileList, error)`
- `GetFile(ctx context.Context, fileID string) (*File, error)`
- `DeleteFile(ctx context.Context, fileID string) (*DeleteFileResponse, error)`
- `DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)`

### Models

- `ListModels(ctx context.Context) (*ModelList, error)`
- `GetModel(ctx context.Context, modelID string) (*Model, error)`
- `DeleteModel(ctx context.Context, modelID string) (*DeleteModelResponse, error)`

## Requirements

- Go 1.24.2 or later

## License

See LICENSE file for details.
