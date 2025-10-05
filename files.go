package mistral

import (
	"io"
)

// FilePurpose represents the purpose of a file
type FilePurpose string

const (
	FilePurposeFineTune FilePurpose = "fine-tune"
	FilePurposeBatch    FilePurpose = "batch"
)

// UploadFileRequest represents a request to upload a file
type UploadFileRequest struct {
	File     io.Reader
	Filename string
	Purpose  FilePurpose
}

// ListFilesParams represents parameters for listing files
type ListFilesParams struct {
	Page     int
	PageSize int
	Purpose  FilePurpose
	Search   string
}
