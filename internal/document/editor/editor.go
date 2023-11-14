package editor

import (
	"bytes"
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
		Cells:       toCells(doc, node, doc.Content()),
		Frontmatter: frontmatter,
		Metadata: map[string]string{
			PrefixAttributeName(InternalAttributePrefix, constants.FinalLineBreaksKey): strconv.Itoa(doc.TrailingLineBreaksCount()),
		},
	}

	// Additionally, put raw frontmatter in notebook's metadata.
	// TODO(adamb): handle the error.
	if raw, err := doc.RawFrontmatter(); err == nil && len(raw) > 0 {
		notebook.Metadata[PrefixAttributeName(InternalAttributePrefix, FrontmatterKey)] = string(raw)
	}

	return notebook, nil
}

func Serialize(notebook *Notebook) ([]byte, error) {
	var result []byte

	// Serialize frontmatter.
	if intro, ok := notebook.Metadata[PrefixAttributeName(InternalAttributePrefix, FrontmatterKey)]; ok {
		intro := []byte(intro)
		lb := document.DetectLineBreak(intro)
		result = append(
			intro,
			bytes.Repeat(lb, 2)...,
		)
	}

	// Serialize cells.
	result = append(result, serializeCells(notebook.Cells)...)

	// Add trailing line breaks.
	if lineBreaks, ok := notebook.Metadata[PrefixAttributeName(InternalAttributePrefix, constants.FinalLineBreaksKey)]; ok {
		desired, err := strconv.Atoi(lineBreaks)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		lb := document.DetectLineBreak(result)
		actual := document.CountTrailingLineBreaks(result, lb)
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
