package notebooks

import (
	"time"
)

type NotebookBlockType string

const (
	NotebookQueryBlockType    NotebookBlockType = "query"
	NotebookMarkdownBlockType NotebookBlockType = "md"
	NotebookFileBlockType     NotebookBlockType = "file"
)

type NotebookQueryBlockInput struct {
	Text string `json:"text"`
}

type NotebookMarkdownBlockInput struct {
	Text string `json:"text"`
}

type LineRange struct {
	// StartLine is the 1-based inclusive start line of the range.
	StartLine int32 `json:"startLine"`

	// EndLine is the 1-based inclusive end line of the range.
	EndLine int32 `json:"endLine"`
}

type NotebookFileBlockInput struct {
	RepositoryName string     `json:"repositoryName"`
	FilePath       string     `json:"filePath"`
	Revision       *string    `json:"revision,omitempty"`
	LineRange      *LineRange `json:"lineRange,omitempty"`
}

type NotebookBlock struct {
	ID            string                      `json:"id"`
	Type          NotebookBlockType           `json:"type"`
	QueryInput    *NotebookQueryBlockInput    `json:"queryInput,omitempty"`
	MarkdownInput *NotebookMarkdownBlockInput `json:"markdownInput,omitempty"`
	FileInput     *NotebookFileBlockInput     `json:"fileInput,omitempty"`
}

type Notebook struct {
	ID            int64
	Title         string
	Blocks        []NotebookBlock
	Public        bool
	CreatorUserID int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
