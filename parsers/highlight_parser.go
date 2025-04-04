package parsers

import (
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A HighlightAst struct represents a Highlight.
type HighlightAst struct {
	gast.BaseInline
}

// Dump implements Node.Dump.
func (n *HighlightAst) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindHighlight is a NodeKind of the Highlight node.
var KindHighlight = gast.NewNodeKind("Highlight")

// Kind implements Node.Kind.
func (n *HighlightAst) Kind() gast.NodeKind {
	return KindHighlight
}

// NewHighlightAst returns a new Highlight node.
func NewHighlightAst() *HighlightAst {
	return &HighlightAst{}
}

type highlightDelimiterProcessor struct {
}

func (p *highlightDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '='
}

func (p *highlightDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *highlightDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return NewHighlightAst()
}

var defaultHighlightDelimiterProcessor = &highlightDelimiterProcessor{}

type highlightParser struct {
}

var defaultHighlightParser = &highlightParser{}

// NewHighlightParser return a new InlineParser that parses
// highlight expressions.
func NewHighlightParser() parser.InlineParser {
	return defaultHighlightParser
}

func (s *highlightParser) Trigger() []byte {
	return []byte{'='}
}

func (s *highlightParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultHighlightDelimiterProcessor)
	if node == nil || node.OriginalLength > 2 || before == '=' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *highlightParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// HighlightHTMLRenderer is a renderer.NodeRenderer implementation that
// renders highlight nodes.
type HighlightHTMLRenderer struct {
	html.Config
}

// NewHighlightHTMLRenderer returns a new HighlightHTMLRenderer.
func NewHighlightHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &HighlightHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *HighlightHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHighlight, r.renderHighlight)
}

// HighlightAttributeFilter defines attribute names which dd elements can have.
var HighlightAttributeFilter = html.GlobalAttributeFilter

func (r *HighlightHTMLRenderer) renderHighlight(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<mark")
			html.RenderAttributes(w, n, HighlightAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<mark>")
		}
	} else {
		_, _ = w.WriteString("</mark>")
	}
	return gast.WalkContinue, nil
}

type highlight struct {
}

// Highlight is an extension that allow you to use mark expression like '==text==' .
var Highlight = &highlight{}

func (e *highlight) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewHighlightParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHighlightHTMLRenderer(), 500),
	))
}
