package project

import "github.com/stateful/runme/internal/document"

type Task struct {
	Filename  string
	CodeBlock *document.CodeBlock
}

type Tasks []Task
