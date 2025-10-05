package mistral

import (
	"io"
)

// FilePurpose represents the intended purpose of a file uploaded to the Mistral API.
// The purpose determines how the file can be used and which API endpoints accept it.
type FilePurpose string

const (
	// FilePurposeFineTune indicates the file contains training data for fine-tuning models.
	// Fine-tuning files should typically be in JSONL format with training examples.
	FilePurposeFineTune FilePurpose = "fine-tune"

	// FilePurposeBatch indicates the file contains batch processing requests.
	// Batch files allow you to process multiple API requests asynchronously
	// at a reduced cost compared to real-time API calls.
	FilePurposeBatch FilePurpose = "batch"
)

// UploadFileRequest represents a request to upload a file to the Mistral API.
// Files can be used for various purposes such as fine-tuning models or batch processing.
type UploadFileRequest struct {
	// File is an io.Reader providing the file content to upload. This could be an open file handle,
	// bytes.Buffer, or any other type implementing io.Reader.
	File io.Reader

	// Filename is the name to assign to the uploaded file. This should include the file extension
	// (e.g., "training_data.jsonl") to help the API identify the file format.
	Filename string

	// Purpose is the intended use for this file. This determines which API operations can use
	// the file and may affect validation and processing requirements.
	Purpose FilePurpose
}

// ListFilesParams represents optional parameters for filtering and paginating file lists.
// All fields are optional; omit them or use zero values to use defaults.
type ListFilesParams struct {
	// Page is the page number to retrieve (1-indexed). Use this with PageSize for pagination.
	// If 0, defaults to the first page.
	Page int

	// PageSize is the number of files to return per page. If 0, uses the API's default page size.
	// Useful for controlling the amount of data returned in a single request.
	PageSize int

	// Purpose filters files by their purpose. If empty, returns files of all purposes.
	// Use this to retrieve only fine-tuning files or only batch files.
	Purpose FilePurpose

	// Search is a search query to filter files by name. If empty, returns all files
	// (subject to other filters). This performs a substring match on filenames.
	Search string
}
