package mistral

// EmbeddingRequest represents a request to the Mistral AI embeddings API.
// Embeddings convert text into dense numerical vectors that capture semantic meaning,
// enabling tasks like similarity search, clustering, recommendation systems, and classification.
type EmbeddingRequest struct {
	// Model is the ID of the embeddings model to use (e.g., "mistral-embed").
	// This is required. Different models may produce embeddings with different dimensions.
	Model string `json:"model"`

	// Input is an array of strings to generate embeddings for. Each string will be converted
	// into a separate embedding vector. Maximum items and total length depend on the model's limits.
	Input []string `json:"input"`

	// EncodingFormat is the format in which to return the embeddings. Options:
	//   - "float" (default) - Returns embeddings as arrays of floating-point numbers
	//   - "base64" - Returns embeddings as base64-encoded strings, which is more compact
	//     for transmission but requires decoding before use
	EncodingFormat string `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents a response from the embeddings API.
// This contains the generated embedding vectors along with metadata about the request.
type EmbeddingResponse struct {
	// ID is a unique identifier for this embeddings request.
	ID string `json:"id"`

	// Object is the object type, typically "list".
	Object string `json:"object"`

	// Data is an array of EmbeddingObject instances, one for each input string.
	// The order matches the order of strings in the Input field of the request.
	// Use the Index field of each embedding to correlate results with inputs.
	Data []EmbeddingObject `json:"data"`

	// Model is the model that was used to generate the embeddings.
	Model string `json:"model"`

	// Usage contains token usage statistics for this request. Embeddings consume prompt tokens
	// based on the length of input text.
	Usage Usage `json:"usage"`
}
