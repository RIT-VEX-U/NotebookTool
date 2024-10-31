package parsers

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type commentBlockParser struct {
}

var defaultCommentBlockParser = &commentBlockParser{}

type commentData struct {
	indent int
}

var commentBlockInfoKey = parser.NewContextKey()

func NewCommentBlockParser() parser.BlockParser {
	return defaultCommentBlockParser
}

func (b *commentBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos == -1 {
		return nil, parser.NoChildren
	}
	if line[pos] != '%' {
		return nil, parser.NoChildren
	}
	i := pos
	for ; i < len(line) && line[i] == '%'; i++ {
	}
	if i-pos < 2 {
		return nil, parser.NoChildren
	}
	pc.Set(commentBlockInfoKey, &commentData{indent: pos})
	node := NewCommentBlock()
	return node, parser.NoChildren
}

func (b *commentBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	data := pc.Get(commentBlockInfoKey).(*commentData)
	w, pos := util.IndentWidth(line, 0)
	if w < 4 {
		i := pos
		for ; i < len(line) && line[i] == '%'; i++ {
		}
		length := i - pos
		if length >= 2 && util.IsBlank(line[i:]) {
			reader.Advance(segment.Stop - segment.Start - segment.Padding)
			return parser.Close
		}
	}

	pos, padding := util.DedentPosition(line, 0, data.indent)
	seg := text.NewSegmentPadding(segment.Start+pos, segment.Stop, padding)
	node.Lines().Append(seg)
	reader.AdvanceAndSetPadding(segment.Stop-segment.Start-pos-1, padding)
	return parser.Continue | parser.NoChildren
}

func (b *commentBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	pc.Set(commentBlockInfoKey, nil)
}

func (b *commentBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *commentBlockParser) CanAcceptIndentedLine() bool {
	return false
}

func (b *commentBlockParser) Trigger() []byte {
	return nil
}

type CommentBlock struct {
	ast.BaseBlock
}

var KindCommentBlock = ast.NewNodeKind("CommentBlock")

func NewCommentBlock() *CommentBlock {
	return &CommentBlock{}
}

func (n *CommentBlock) Dump(source []byte, level int) {
	m := map[string]string{}
	ast.DumpHelper(n, source, level, m, nil)
}

func (n *CommentBlock) Kind() ast.NodeKind {
	return KindCommentBlock
}

func (n *CommentBlock) IsRaw() bool {
	return true
}

type comment struct{}

var CommentInst = &comment{}

func (e *comment) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewCommentBlockParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewCommentBlockRenderer(), 500),
	))
}

type commentBlockRenderer struct {
	html.Config
}

// NewHighlightHTMLRenderer returns a new HighlightHTMLRenderer.
func NewCommentBlockRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &commentBlockRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return &commentBlockRenderer{}
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *commentBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHighlight, r.renderHighlight)
}

func (r *commentBlockRenderer) renderHighlight(
	w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
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
	return ast.WalkContinue, nil
}

type inlineCommentParser struct {
}

var defaultInlineCommentParser = &inlineCommentParser{}

func NewInlineCommentParser() parser.InlineParser {
	return defaultInlineCommentParser
}

func (s *inlineCommentParser) Trigger() []byte {
	return []byte{'%'}
}

func (s *inlineCommentParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, startSegment := block.PeekLine()
	opener := 0
	for ; opener < len(line) && line[opener] == '%'; opener++ {
	}
	block.Advance(opener)
	l, pos := block.Position()
	node := NewInlineComment()
	for {
		line, segment := block.PeekLine()
		if line == nil {
			block.SetPosition(l, pos)
			return ast.NewTextSegment(startSegment.WithStop(startSegment.Start + opener))
		}
		for i := 0; i < len(line); i++ {
			c := line[i]
			if c == '%' {
				oldi := i
				for ; i < len(line) && line[i] == '%'; i++ {
				}
				closure := i - oldi
				if closure == opener && (i+1 >= len(line) || line[i+1] != '%') {
					segment := segment.WithStop(segment.Start + i - closure)
					if !segment.IsEmpty() {
						node.AppendChild(node, ast.NewRawTextSegment(segment))
					}
					block.Advance(i)
					goto end
				}
			}
		}
		if !util.IsBlank(line) {
			node.AppendChild(node, ast.NewRawTextSegment(segment))
		}
		block.AdvanceLine()
	}
end:

	if !node.IsBlank(block.Source()) {
		// trim first halfspace and last halfspace
		segment := node.FirstChild().(*ast.Text).Segment
		shouldTrimmed := true
		if !(!segment.IsEmpty() && block.Source()[segment.Start] == ' ') {
			shouldTrimmed = false
		}
		segment = node.LastChild().(*ast.Text).Segment
		if !(!segment.IsEmpty() && block.Source()[segment.Stop-1] == ' ') {
			shouldTrimmed = false
		}
		if shouldTrimmed {
			t := node.FirstChild().(*ast.Text)
			segment := t.Segment
			t.Segment = segment.WithStart(segment.Start + 1)
			t = node.LastChild().(*ast.Text)
			segment = node.LastChild().(*ast.Text).Segment
			t.Segment = segment.WithStop(segment.Stop - 1)
		}

	}
	return node
}

func NewInlineCommentRenderer(start, end string) renderer.NodeRenderer {
	return &InlineCommentRenderer{start, end}
}

type InlineCommentRenderer struct {
	startDelim string
	endDelim   string
}

func (r *InlineCommentRenderer) renderInlineComment(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<span class="math inline">` + r.startDelim)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				w.Write(value[:len(value)-1])
				if c != n.LastChild() {
					w.Write([]byte(" "))
				}
			} else {
				w.Write(value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString(r.endDelim + `</span>`)
	return ast.WalkContinue, nil
}

func (r *InlineCommentRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInlineComment, r.renderInlineComment)
}

type InlineComment struct {
	ast.BaseInline
}

func (n *InlineComment) Inline() {}

func (n *InlineComment) IsBlank(source []byte) bool {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		text := c.(*ast.Text).Segment
		if !util.IsBlank(text.Value(source)) {
			return false
		}
	}
	return true
}

func (n *InlineComment) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var KindInlineComment = ast.NewNodeKind("InlineComment")

func (n *InlineComment) Kind() ast.NodeKind {
	return KindInlineComment
}

func NewInlineComment() *InlineComment {
	return &InlineComment{
		BaseInline: ast.BaseInline{},
	}
}
