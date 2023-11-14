package document

import (
	"bytes"
	"testing"

	"github.com/stateful/runme/internal/document/identity"
	"github.com/stateful/runme/internal/renderer/cmark"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var identityResolver = identity.NewResolver(identity.AllLifecycleIdentity)

func TestDocument_Parse(t *testing.T) {
	data := []byte(`# Examples

First paragraph.

> bq1
>
>     echo "inside bq"
>
> bq2
> bq3

1. Item 1

   ` + "```" + `sh {name=echo first= second=2}
   $ echo "Hello, runme!"
   ` + "```" + `

   Inner paragraph

2. Item 2
3. Item 3
`)
	doc := New(data)
	node, err := doc.Root()
	require.NoError(t, err)
	assert.Len(t, node.children, 4)
	assert.Len(t, node.children[0].children, 0)
	assert.Len(t, node.children[1].children, 0)
	assert.Len(t, node.children[2].children, 3)
	assert.Equal(t, "> bq1\n>\n>     echo \"inside bq\"\n>\n> bq2\n> bq3\n", string(node.children[2].Item().Value()))
	assert.Len(t, node.children[2].children[0].children, 0)
	assert.Equal(t, "bq1\n", string(node.children[2].children[0].Item().Value()))
	assert.Len(t, node.children[2].children[1].children, 0)
	assert.Equal(t, "    echo \"inside bq\"\n", string(node.children[2].children[1].Item().Value()))
	assert.Len(t, node.children[2].children[2].children, 0)
	assert.Equal(t, "bq2\nbq3\n", string(node.children[2].children[2].Item().Value()))
	assert.Len(t, node.children[3].children, 3)
	assert.Len(t, node.children[3].children[0].children, 3)
	assert.Equal(t, "1. Item 1\n\n   ```sh {name=echo first= second=2}\n   $ echo \"Hello, runme!\"\n   ```\n\n   Inner paragraph\n", string(node.children[3].children[0].Item().Value()))
	assert.Len(t, node.children[3].children[0].children[0].children, 0)
	assert.Equal(t, "Item 1\n", string(node.children[3].children[0].children[0].Item().Value()))
	assert.Len(t, node.children[3].children[0].children[1].children, 0)
	assert.Equal(t, "```sh {name=echo first= second=2}\n$ echo \"Hello, runme!\"\n```\n", string(node.children[3].children[0].children[1].Item().Value()))
	assert.Len(t, node.children[3].children[0].children[2].children, 0)
	assert.Equal(t, "Inner paragraph\n", string(node.children[3].children[0].children[2].Item().Value()))
	assert.Len(t, node.children[3].children[1].children, 1)
	assert.Equal(t, "2. Item 2\n", string(node.children[3].children[1].Item().Value()))
	assert.Len(t, node.children[3].children[1].children[0].children, 0)
	assert.Equal(t, "Item 2\n", string(node.children[3].children[1].children[0].Item().Value()))
	assert.Len(t, node.children[3].children[2].children, 1)
	assert.Equal(t, "3. Item 3\n", string(node.children[3].children[2].Item().Value()))
	assert.Len(t, node.children[3].children[2].children[0].children, 0)
	assert.Equal(t, "Item 3\n", string(node.children[3].children[2].children[0].Item().Value()))
}

func TestDocument_Frontmatter(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		data := bytes.TrimSpace([]byte(`
---
key: value
---

First paragraph
`,
		))

		doc := New(data)
		_, err := doc.Root()
		require.NoError(t, err)

		rawFrontmatter, err := doc.RawFrontmatter()
		require.NoError(t, err)
		assert.Equal(t, []byte("---\nkey: value\n---"), rawFrontmatter)

		frontmatter, err := doc.Frontmatter()
		require.NoError(t, err)
		assert.EqualValues(t, &Frontmatter{}, frontmatter)

		assert.Equal(t, []byte("First paragraph"), doc.Content())
		assert.Equal(t, 20, doc.ContentOffset())
	})

	t.Run("Format", func(t *testing.T) {
		testCases := []struct {
			Name string
			Data []byte
		}{
			{
				Name: "YAML",
				Data: bytes.TrimSpace([]byte(`
---
shell: fish
---
				`,
				)),
			},
			{
				Name: "YAML",
				Data: bytes.TrimSpace([]byte(`
---
{
  shell: fish
}
---
				`,
				)),
			},
			{
				Name: "TOML",
				Data: bytes.TrimSpace([]byte(`
---
shell = "fish"
---
				`,
				)),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				doc := New(tc.Data)
				_, err := doc.Root()
				require.NoError(t, err)

				frontmatter, err := doc.Frontmatter()
				assert.NoError(t, err)
				assert.EqualValues(t, &Frontmatter{Shell: "fish"}, frontmatter)
			})
		}
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		data := bytes.TrimSpace([]byte(`
---
{
  "shell": "bin/sh",
  "cwd": "/path/to/cwd
}
---
`,
		))
		doc := New(data)
		_, err := doc.Root()
		require.NoError(t, err)

		frontmatter, err := doc.Frontmatter()
		assert.ErrorContains(t, err, "failed to parse frontmatter content: yaml: line 3: found unexpected end of stream")
		assert.EqualValues(t, &Frontmatter{}, frontmatter)
	})
}

func TestDocument_TrailingLineBreaks(t *testing.T) {
	data := []byte(`This will test final line breaks`)

	t.Run("No breaks", func(t *testing.T) {
		doc := New(data)
		astNode, err := doc.RootASTNode()
		require.NoError(t, err)

		actual, err := cmark.Render(astNode, data)
		require.NoError(t, err)

		assert.Equal(
			t,
			string(data),
			string(actual),
		)
		assert.Equal(t, 0, doc.TrailingLineBreaksCount())
	})

	t.Run("1 LF", func(t *testing.T) {
		withLineBreaks := append(data, bytes.Repeat([]byte{'\n'}, 1)...)
		doc := New(withLineBreaks)
		astNode, err := doc.RootASTNode()
		require.NoError(t, err)

		actual, err := cmark.Render(astNode, withLineBreaks)
		require.NoError(t, err)

		assert.Equal(
			t,
			string(withLineBreaks),
			string(actual),
		)
		assert.Equal(t, 1, doc.TrailingLineBreaksCount())
	})

	t.Run("1 CRLF", func(t *testing.T) {
		withLineBreaks := append(data, bytes.Repeat([]byte{'\r', '\n'}, 1)...)
		doc := New(withLineBreaks)
		astNode, err := doc.RootASTNode()
		require.NoError(t, err)

		actual, err := cmark.Render(astNode, withLineBreaks)
		require.NoError(t, err)

		assert.Equal(
			t,
			string(withLineBreaks),
			string(actual),
		)
		assert.Equal(t, 1, doc.TrailingLineBreaksCount())
	})

	t.Run("7 LFs", func(t *testing.T) {
		withLineBreaks := append(data, bytes.Repeat([]byte{'\n'}, 7)...)
		doc := New(withLineBreaks)
		astNode, err := doc.RootASTNode()
		require.NoError(t, err)

		actual, err := cmark.Render(astNode, withLineBreaks)
		require.NoError(t, err)

		assert.Equal(
			t,
			string(actual),
			string(withLineBreaks),
		)
		assert.Equal(t, 7, doc.TrailingLineBreaksCount())
	})
}
