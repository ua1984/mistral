package mistral

import "time"

// APIError represents an error response from the Mistral API
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Type       string `json:"type,omitempty"`
	Code       string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return e.Type + ": " + e.Message
	}
	return e.Message
}

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role       Role        `json:"role"`
	Content    interface{} `json:"content"` // Can be string or array of content parts
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call made by the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool that can be used by the model
type Tool struct {
	Type     string              `json:"type"`
	Function ToolFunctionDetails `json:"function"`
}

// ToolFunctionDetails represents the details of a tool function
type ToolFunctionDetails struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolChoice represents how the model should use tools
type ToolChoice string

const (
	ToolChoiceAuto ToolChoice = "auto"
	ToolChoiceAny  ToolChoice = "any"
	ToolChoiceNone ToolChoice = "none"
)

// ResponseFormat represents the format of the response
type ResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Model represents a model card
type Model struct {
	ID           string   `json:"id"`
	Object       string   `json:"object"`
	Created      int64    `json:"created"`
	OwnedBy      string   `json:"owned_by"`
	Type         string   `json:"type,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Description  string   `json:"description,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
}

// ModelList represents a list of models
type ModelList struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// DeleteModelResponse represents the response from deleting a model
type DeleteModelResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// EmbeddingObject represents a single embedding
type EmbeddingObject struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// File represents a file object
type File struct {
	ID        string    `json:"id"`
	Object    string    `json:"object"`
	Bytes     int       `json:"bytes"`
	CreatedAt time.Time `json:"created_at"`
	Filename  string    `json:"filename"`
	Purpose   string    `json:"purpose"`
	Source    string    `json:"source,omitempty"`
}

// FileList represents a list of files
type FileList struct {
	Object string `json:"object"`
	Data   []File `json:"data"`
}

// DeleteFileResponse represents the response from deleting a file
type DeleteFileResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// Agent represents an agent entity
type Agent struct {
	ID          string                 `json:"id"`
	Object      string                 `json:"object"`
	CreatedAt   time.Time              `json:"created_at"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Model       string                 `json:"model"`
	Instructions string                `json:"instructions,omitempty"`
	Tools       []Tool                 `json:"tools,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Conversation represents a conversation entity
type Conversation struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	CreatedAt time.Time              `json:"created_at"`
	Model     string                 `json:"model,omitempty"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
