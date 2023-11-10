package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"

	"github.com/stateful/runme/internal/document/constants"
	"github.com/stateful/runme/internal/document/identity"
	"github.com/stateful/runme/internal/renderer/cmark"
)

var DefaultRenderer = cmark.Render

type Document struct {
	source           []byte
	identityResolver *identity.IdentityResolver
	nameResolver     *nameResolver
	parser           parser.Parser
	renderer         Renderer

	onceParse      sync.Once
	parseErr       error
	rootASTNode    ast.Node
	rootNode       *Node
	frontmatterRaw []byte
	content        []byte // raw data behind frontmatter
	contentOffset  int

	onceFrontmatter sync.Once
	frontmatterErr  error
	frontmatter     *Frontmatter
}

func New(source []byte, identityResolver *identity.IdentityResolver) *Document {
	return &Document{
		source:           source,
		identityResolver: identityResolver,
		nameResolver: &nameResolver{
			namesCounter: map[string]int{},
			cache:        map[interface{}]string{},
		},
		parser:    goldmark.DefaultParser(),
		renderer:  DefaultRenderer,
		onceParse: sync.Once{},
	}
}

func (d *Document) Content() []byte {
	return d.content
}

func (d *Document) ContentOffset() int {
	return d.contentOffset
}

func (d *Document) RootASTNode() (ast.Node, error) {
	d.parse()

	if d.parseErr != nil {
		return nil, d.parseErr
	}

	return d.rootASTNode, nil
}

func (d *Document) Root() (*Node, error) {
	d.parse()

	if d.parseErr != nil {
		return nil, d.parseErr
	}

	return d.rootNode, nil
}

func (d *Document) parse() {
	d.onceParse.Do(func() {
		if err := d.splitSource(); err != nil {
			d.parseErr = err
			return
		}

		d.rootASTNode = d.parser.Parse(text.NewReader(d.content))

		node := &Node{}
		if err := d.buildBlocksTree(d.rootASTNode, node); err != nil {
			d.parseErr = err
			return
		}

		d.rootNode = node

		// Retain trailing new lines. This information must be stored in
		// ast.Node's attributes because it's later used by internal/renderer/cmark.Render,
		// which does not use anything else than ast.Node.
		finalNewLines := CountFinalLineBreaks(d.source, DetectLineBreak(d.source))
		d.rootASTNode.SetAttributeString(constants.FinalLineBreaksKey, finalNewLines)
	})
}

func (d *Document) splitSource() error {
	l := &itemParser{input: d.source}

	runItemParser(l, parseInit)

	for _, item := range l.items {
		switch item.Type() {
		case parsedItemFrontMatter:
			d.frontmatterRaw = item.Value(d.source)
		case parsedItemContent:
			d.content = item.Value(d.source)
			d.contentOffset = item.start
		case parsedItemError:
			// TODO(adamb): handle this error somehow
			if !errors.Is(item.err, errParseFrontmatter) {
				return item.err
			}
		}
	}

	return nil
}

// TODO: use errors from stdlib
var (
	ErrFrontmatterInvalid  = errors.New("invalid frontmatter")
	ErrFrontmatterNotFound = errors.New("not found frontmatter")
)

type Frontmatter struct {
	Shell       string
	Cwd         string
	SkipPrompts bool `yaml:"skipPrompts,omitempty"`
}

func (f *Frontmatter) IsEmpty() bool {
	return *f != Frontmatter{}
}

func (d *Document) RawFrontmatter() ([]byte, error) {
	d.parse()

	if d.parseErr != nil {
		return nil, d.parseErr
	}

	return d.frontmatterRaw, nil
}

func (d *Document) Frontmatter() (f *Frontmatter, _ error) {
	d.parse()

	if d.parseErr != nil {
		return f, d.parseErr
	}

	d.parseFrontmatterOnce()

	return d.frontmatter, d.frontmatterErr
}

func (d *Document) parseFrontmatterOnce() {
	d.onceFrontmatter.Do(func() {
		if len(d.frontmatterRaw) == 0 {
			return
		}

		var f Frontmatter

		raw := d.frontmatterRaw
		lines := bytes.Split(raw, []byte{'\n'})

		if len(lines) < 2 || !bytes.Equal(bytes.TrimSpace(lines[0]), bytes.TrimSpace(lines[len(lines)-1])) {
			d.frontmatterErr = errors.WithStack(ErrFrontmatterInvalid)
			return
		}

		raw = bytes.Join(lines[1:len(lines)-1], []byte{'\n'})

		empty := Frontmatter{}

		{
			yamlerr := yaml.Unmarshal(raw, &f)
			if f != empty {
				d.frontmatter = &f
				d.frontmatterErr = errors.WithStack(yamlerr)
				return
			}
		}

		{
			jsonerr := json.Unmarshal(raw, &f)
			if f != empty {
				d.frontmatter = &f
				d.frontmatterErr = errors.WithStack(jsonerr)
				return
			}
		}

		{
			tomlerr := toml.Unmarshal(raw, &f)
			if f != empty {
				d.frontmatter = &f
				d.frontmatterErr = errors.WithStack(tomlerr)
				return
			}
		}
	})
}

func (d *Document) buildBlocksTree(parent ast.Node, node *Node) error {
	for astNode := parent.FirstChild(); astNode != nil; astNode = astNode.NextSibling() {
		switch astNode.Kind() {
		case ast.KindFencedCodeBlock:
			block, err := newCodeBlock(
				d,
				astNode.(*ast.FencedCodeBlock),
				d.identityResolver,
				d.nameResolver,
				d.content,
				d.renderer,
			)
			if err != nil {
				return errors.WithStack(err)
			}
			node.add(block)
		case ast.KindBlockquote, ast.KindList, ast.KindListItem:
			block, err := newInnerBlock(astNode, d.content, d.renderer)
			if err != nil {
				return errors.WithStack(err)
			}
			nNode := node.add(block)
			if err := d.buildBlocksTree(astNode, nNode); err != nil {
				return err
			}
		default:
			block, err := newMarkdownBlock(astNode, d.content, d.renderer)
			if err != nil {
				return errors.WithStack(err)
			}
			node.add(block)
		}
	}
	return nil
}

type nameResolver struct {
	namesCounter map[string]int
	cache        map[interface{}]string
}

func (r *nameResolver) Get(obj interface{}, name string) string {
	if v, ok := r.cache[obj]; ok {
		return v
	}
	var result string
	r.namesCounter[name]++
	if r.namesCounter[name] == 1 {
		result = name
	} else {
		result = fmt.Sprintf("%s-%d", name, r.namesCounter[name])
	}
	r.cache[obj] = result
	return result
}

func CountFinalLineBreaks(source []byte, lineBreak []byte) int {
	i := len(source) - len(lineBreak)
	numBreaks := 0

	for i >= 0 && bytes.Equal(source[i:i+len(lineBreak)], lineBreak) {
		i -= len(lineBreak)
		numBreaks++
	}

	return numBreaks
}

func DetectLineBreak(source []byte) []byte {
	crlfCount := bytes.Count(source, []byte{'\r', '\n'})
	lfCount := bytes.Count(source, []byte{'\n'})
	if crlfCount == lfCount {
		return []byte{'\r', '\n'}
	}
	return []byte{'\n'}
}
