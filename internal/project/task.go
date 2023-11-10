package project

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
)

type Task struct {
	Filename    string
	Frontmatter *document.Frontmatter
	CodeBlock   *document.CodeBlock
}

func (t Task) ID() string {
	return t.Filename + ":" + t.CodeBlock.Name()
}

func (t Task) IsEmpty() bool {
	return t == Task{}
}

type Tasks []Task

func (tasks Tasks) LookupByID(query string) (Tasks, error) {
	if query == "" {
		return tasks, nil
	}

	matcher, err := compileQuery(query)
	if err != nil {
		return nil, err
	}

	var result Tasks

	for _, task := range tasks {
		if matcher.MatchString(task.ID()) {
			continue
		}
		result = append(result, task)
	}

	return result, nil
}

func (tasks Tasks) LookupByName(name string) Tasks {
	var result Tasks

	for _, task := range tasks {
		if task.CodeBlock.Name() == name {
			result = append(result, task)
		}
	}

	return result
}

type ErrCodeBlockFileNotFound struct {
	queryFile string
}

func (e ErrCodeBlockFileNotFound) Error() string {
	return fmt.Sprintf("unable to find file in project matching regex %q", e.FailedFileQuery())
}

func (e ErrCodeBlockFileNotFound) FailedFileQuery() string {
	return e.queryFile
}

type ErrCodeBlockNameNotFound struct {
	queryName string
}

func (e ErrCodeBlockNameNotFound) Error() string {
	return fmt.Sprintf("unable to find any script named %q", e.queryName)
}

func (e ErrCodeBlockNameNotFound) FailedNameQuery() string {
	return e.queryName
}

func IsCodeBlockNotFoundError(err error) bool {
	return errors.As(err, &ErrCodeBlockNameNotFound{}) || errors.As(err, &ErrCodeBlockFileNotFound{})
}

func (tasks Tasks) LookupWithFileAndName(queryFile, name string) (Tasks, error) {
	if queryFile == "" {
		return tasks.LookupByName(name), nil
	}

	matcher, err := compileQuery(queryFile)
	if err != nil {
		return nil, err
	}

	var result Tasks

	foundFile := false

	for _, task := range tasks {
		if !matcher.MatchString(task.Filename) {
			continue
		}

		foundFile = true

		if name != task.CodeBlock.Name() {
			continue
		}

		result = append(result, task)
	}

	if len(result) == 0 {
		if !foundFile {
			return nil, ErrCodeBlockFileNotFound{queryFile: queryFile}
		}

		return nil, ErrCodeBlockNameNotFound{queryName: name}
	}

	return result, nil
}

func CodeBlocksFromTasks(tasks Tasks) []*document.CodeBlock {
	result := make([]*document.CodeBlock, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, task.CodeBlock)
	}
	return result
}

func compileQuery(query string) (*regexp.Regexp, error) {
	reg, err := regexp.Compile(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed compiling query %q to regexp: %v", query, err)
	}
	return reg, nil
}

func FilterTasks(tasks Tasks, fn func(Task) (bool, bool)) Tasks {
	var result Tasks

	for _, t := range tasks {
		include, finish := fn(t)

		if include {
			result = append(result, t)
		}

		if finish {
			break
		}
	}

	return result
}
