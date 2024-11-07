package parsers

import (
	"fmt"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A CommentAst struct represents a Comment.
type CommentAst struct {
	gast.BaseInline
}

// Dump implements Node.Dump.
func (n *CommentAst) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindComment is a NodeKind of the Comment node.
var KindComment = gast.NewNodeKind("Comment")

// Kind implements Node.Kind.
func (n *CommentAst) Kind() gast.NodeKind {
	return KindComment
}

// NewCommentAst returns a new Comment node.
func NewCommentAst() *CommentAst {
	return &CommentAst{}
}

type commentDelimiterProcessor struct {
}

func (p *commentDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '%'
}

func (p *commentDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *commentDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return NewCommentAst()
}

var defaultcommentDelimiterProcessor = &commentDelimiterProcessor{}

type commentParser struct {
}

var defaultcommentParser = &commentParser{}

// NewcommentParser return a new InlineParser that parses
// comment expressions.
func NewcommentParser() parser.InlineParser {
	fmt.Println("New comment parser")
	return defaultcommentParser
}

func (s *commentParser) Trigger() []byte {
	return []byte{'%'}
}

func (s *commentParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultcommentDelimiterProcessor)
	fmt.Println("commentParser Parse got line:", string(line))
	if node == nil || node.OriginalLength > 2 || before == '%' {
		fmt.Println("commentParser Parse skipping line:", string(line))
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *commentParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
	fmt.Println("Comment Parser close block")
}

// CommentHTMLRenderer is a renderer.NodeRenderer implementation that
// renders comment nodes.
type CommentHTMLRenderer struct {
	html.Config
}

// NewCommentHTMLRenderer returns a new CommentHTMLRenderer.
func NewCommentHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &CommentHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *CommentHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindComment, r.renderComment)
}

// CommentAttributeFilter defines attribute names which dd elements can have.
var CommentAttributeFilter = html.GlobalAttributeFilter

func (r *CommentHTMLRenderer) renderComment(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<span class=\"obsidian-comment\"")
			html.RenderAttributes(w, n, CommentAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<span class = \"obsidian-comment\">")
		}
	} else {
		_, _ = w.WriteString("</span>")
	}
	return gast.WalkContinue, nil
}

type comment struct {
}

// Comment is an extension that allow you to use comment expressions like '%%text%%' .
var Comment = &comment{}

func (e *comment) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewcommentParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewCommentHTMLRenderer(), 500),
	))
}
