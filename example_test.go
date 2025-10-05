package mistral_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ua1984/mistral"
)

func ExampleClient_CreateChatCompletion() {
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

func ExampleClient_CreateChatCompletionStream() {
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
}

func ExampleClient_CreateEmbedding() {
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
}

func ExampleClient_ListModels() {
	client := mistral.NewClient(os.Getenv("MISTRAL_API_KEY"))

	models, err := client.ListModels(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, model := range models.Data {
		fmt.Printf("Model: %s\n", model.ID)
	}
}

func ExampleClient_UploadFile() {
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
}
