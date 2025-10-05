package mistral

// ChatCompletionRequest represents a request to the Mistral AI chat completions API.
// This structure contains all parameters needed to generate text completions based on
// conversational context. The request supports both simple text generation and advanced
// features like function calling, structured outputs, and streaming responses.
type ChatCompletionRequest struct {
	// Model is the ID of the model to use (e.g., "mistral-large-latest", "mistral-small").
	// This is required. Choose based on your needs for speed vs. capability.
	Model string `json:"model"`

	// Messages is an array of ChatMessage objects that form the conversation history.
	// Must contain at least one message. The model will generate a response based on this context.
	Messages []ChatMessage `json:"messages"`

	// Temperature controls randomness in generation. Range: 0.0 to 1.0 (or higher for some models).
	// Lower values (e.g., 0.2) make output more focused and deterministic. Higher values (e.g., 0.8)
	// make output more creative and varied. Default is typically 0.7.
	Temperature *float64 `json:"temperature,omitempty"`

	// TopP is the nucleus sampling parameter. Range: 0.0 to 1.0. Alternative to temperature for
	// controlling randomness. The model considers tokens with top_p probability mass.
	// For example, 0.1 means only tokens comprising the top 10% probability mass are considered.
	TopP *float64 `json:"top_p,omitempty"`

	// MaxTokens is the maximum number of tokens to generate in the completion.
	// The total length of input tokens plus max_tokens cannot exceed the model's context length.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// MinTokens is the minimum number of tokens to generate. Useful when you need
	// a response of at least a certain length.
	MinTokens *int `json:"min_tokens,omitempty"`

	// Stream indicates whether to stream partial message deltas as server-sent events.
	// If true, use CreateChatCompletionStream instead of CreateChatCompletion.
	Stream bool `json:"stream,omitempty"`

	// Stop contains up to 4 sequences where the API will stop generating further tokens.
	// The returned text will not contain the stop sequence.
	Stop []string `json:"stop,omitempty"`

	// RandomSeed, if specified, makes the system attempt to sample deterministically
	// such that repeated requests with the same seed and parameters return the same result.
	// Determinism is not guaranteed.
	RandomSeed *int `json:"random_seed,omitempty"`

	// Tools is a list of tools/functions the model may call. The model can choose to call
	// one or more of these functions if it determines they would help fulfill the request.
	Tools []Tool `json:"tools,omitempty"`

	// ToolChoice controls how the model uses the provided tools. Options: "auto", "any", "none".
	// Default is "auto", which lets the model decide whether to use tools.
	ToolChoice ToolChoice `json:"tool_choice,omitempty"`

	// ResponseFormat specifies the format of the response. Set to {"type": "json_object"}
	// to enable JSON mode, which guarantees the message the model generates is valid JSON.
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// SafePrompt indicates whether to inject a safety prompt before all conversations. Default is false.
	SafePrompt bool `json:"safe_prompt,omitempty"`

	// N is how many chat completion choices to generate for each input message.
	// Note: N>1 may consume significantly more tokens.
	N *int `json:"n,omitempty"`

	// PresencePenalty is a number between -2.0 and 2.0. Positive values penalize new tokens
	// based on whether they appear in the text so far, increasing the model's likelihood
	// to talk about new topics.
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty is a number between -2.0 and 2.0. Positive values penalize new tokens
	// based on their existing frequency in the text so far, decreasing the model's likelihood
	// to repeat the same line verbatim.
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// Metadata is optional metadata to attach to the request for tracking and filtering purposes.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatCompletionResponse represents a complete response from the chat completions API.
// This is returned by non-streaming chat completion requests and contains the full
// generated response along with metadata about the request.
type ChatCompletionResponse struct {
	// ID is a unique identifier for this completion request.
	ID string `json:"id"`

	// Object is the object type, typically "chat.completion".
	Object string `json:"object"`

	// Created is a Unix timestamp (in seconds) of when the completion was created.
	Created int64 `json:"created"`

	// Model is the model used for generating the completion.
	Model string `json:"model"`

	// Choices is an array of completion choices. If N=1 in the request, this will contain
	// a single choice. Each choice contains the generated message and metadata.
	Choices []ChatCompletionChoice `json:"choices"`

	// Usage contains token usage statistics for this request, including prompt tokens,
	// completion tokens, and total tokens used.
	Usage Usage `json:"usage"`
}

// ChatCompletionChoice represents a single generated completion alternative.
// When N>1 in the request, multiple choices are generated and you can select
// the most appropriate one for your use case.
type ChatCompletionChoice struct {
	// Index is the index of this choice in the list of choices (0-based).
	Index int `json:"index"`

	// Message is the generated chat message from the assistant. This contains the actual
	// response content and any tool calls the model wants to make.
	Message ChatMessage `json:"message"`

	// FinishReason is the reason why the model stopped generating tokens. Possible values:
	//   - "stop" - Natural stopping point or provided stop sequence was reached
	//   - "length" - Maximum token limit was reached
	//   - "tool_calls" - The model called a function/tool
	//   - "content_filter" - Content was filtered due to safety settings
	FinishReason string `json:"finish_reason"`

	// Delta is used only in streaming responses. Contains the incremental changes
	// to the message as new tokens are generated. Null in non-streaming responses.
	Delta *ChatMessage `json:"delta,omitempty"`
}

// ChatCompletionStreamResponse represents a single chunk in a streaming response.
// When using streaming mode (Stream: true), the API returns multiple chunks as
// server-sent events. Each chunk contains incremental updates to the response.
// Combine all chunks to reconstruct the complete response.
type ChatCompletionStreamResponse struct {
	// ID is a unique identifier for this completion request (consistent across all chunks).
	ID string `json:"id"`

	// Object is the object type, typically "chat.completion.chunk".
	Object string `json:"object"`

	// Created is a Unix timestamp (in seconds) of when the completion was created.
	Created int64 `json:"created"`

	// Model is the model used for generating the completion.
	Model string `json:"model"`

	// Choices is an array of completion choice deltas. Each choice contains the Delta field
	// with incremental content. The last chunk will have a FinishReason set.
	Choices []ChatCompletionChoice `json:"choices"`
}
