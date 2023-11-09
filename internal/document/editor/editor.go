package editor

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"

	"github.com/stateful/runme/internal/document/constants"
)

const FrontmatterKey = "frontmatter"

func Deserialize(data []byte) (*Notebook, error) {
	// Deserialize content to cells.
	doc := document.New(data)
	node, err := doc.Root()
	if err != nil {
		return nil, err
	}

	frontmatter, err := doc.Frontmatter()
	if err != nil {
		return nil, err
	}

	notebook := &Notebook{
		Cells:       toCells(doc, node, data),
		Frontmatter: frontmatter,
	}

	// TODO(adamb): this should be available from `doc`
	finalLinesBreaks := document.CountFinalLineBreaks(data, document.DetectLineBreak(data))
	notebook.Metadata = map[string]string{
		PrefixAttributeName(InternalAttributePrefix, constants.FinalLineBreaksKey): fmt.Sprint(finalLinesBreaks),
	}

	if raw, err := doc.RawFrontmatter(); err == nil && len(raw) > 0 {
		notebook.Metadata[PrefixAttributeName(InternalAttributePrefix, FrontmatterKey)] = string(raw)
	}

	return notebook, nil
}

func Serialize(notebook *Notebook) ([]byte, error) {
	var result []byte

	if intro, ok := notebook.Metadata[PrefixAttributeName(InternalAttributePrefix, FrontmatterKey)]; ok {
		intro := []byte(intro)
		lb := document.DetectLineBreak(intro)
		result = append(
			intro,
			bytes.Repeat(lb, 2)...,
		)
	}

	result = append(result, serializeCells(notebook.Cells)...)

	if lineBreaks, ok := notebook.Metadata[PrefixAttributeName(InternalAttributePrefix, constants.FinalLineBreaksKey)]; ok {
		desired, err := strconv.ParseInt(lineBreaks, 10, 32)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		lb := document.DetectLineBreak(result)
		actual := document.CountFinalLineBreaks(result, lb)
		delta := int(desired) - actual

		if delta < 0 {
			end := len(result) + delta*len(lb)
			result = result[0:max(0, end)]
		} else {
			result = append(result, bytes.Repeat(lb, delta)...)
		}
	}

	return result, nil
}
