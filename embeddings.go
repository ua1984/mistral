package mistral

// EmbeddingRequest represents a request to the embeddings API
type EmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"` // "float" or "base64"
}

// EmbeddingResponse represents a response from the embeddings API
type EmbeddingResponse struct {
	ID     string            `json:"id"`
	Object string            `json:"object"`
	Data   []EmbeddingObject `json:"data"`
	Model  string            `json:"model"`
	Usage  Usage             `json:"usage"`
}
