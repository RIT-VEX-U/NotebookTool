package main

import (
	"NotebookTool/parsers"
	"fmt"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	callout "gitlab.com/staticnoise/goldmark-callout"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/mermaid"
	"go.abhg.dev/goldmark/wikilink"
)

type wikilinkResolver struct{}

// ResolveWikilink returns the address of the page that the provided
// wikilink points to. The destination will be URL-escaped before
// being placed into a link.
//
// If ResolveWikilink returns a non-nil error, rendering will be
// halted.
//
// If ResolveWikilink returns a nil destination and error, the
// Renderer will omit the link and render its contents as a regular
// string.
func (wr wikilinkResolver) ResolveWikilink(n *wikilink.Node) (destination []byte, err error) {
	var _html = []byte(".html")
	var _hash = []byte("#")

	var s = []byte{}
	s = append(s, []byte("Assets/")...)
	if len(n.Target) > 0 {
		if n.Target[len(n.Target)-1] == '\\' {
			n.Target = n.Target[:len(n.Target)-1]
		}
		s = append(s, n.Target...)
		if filepath.Ext(string(n.Target)) == "" {
			s = append(s, _html...)
		}
	}
	if len(n.Fragment) > 0 {
		s = append(s, _hash...)
		s = append(s, n.Fragment...)
	}
	return s, nil

}

type PreWrapper struct {
}

// End implements html.PreWrapper.
func (p *PreWrapper) End(code bool) string {
	return "</div>"
}

// Start implements html.PreWrapper.
func (p *PreWrapper) Start(code bool, styleAttr string) string {
	fmt.Println("code", code, "styleAttr", styleAttr)
	return "<div class='dontdisplaycheck'>"
}

var _ chromahtml.PreWrapper = &PreWrapper{}

func Md() goldmark.Markdown {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.New(meta.WithStoresInDocument()),
			&wikilink.Extender{
				Resolver: wikilinkResolver{},
			},
			&hashtag.Extender{
				Resolver: nil,
				Variant:  hashtag.ObsidianVariant,
			},
			parsers.Highlight,
			parsers.Comment,
			emoji.Emoji,
			&mermaid.Extender{},
			callout.CalloutExtention,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
				highlighting.WithCodeBlockOptions(func(c highlighting.CodeBlockContext) []chromahtml.Option {
					if language, ok := c.Language(); ok {
						// Add custom wrapper for js (dv.views(check))
						if strings.Contains(string(language), "js") {
							return []chromahtml.Option{
								chromahtml.WithPreWrapper(&PreWrapper{}),
							}
						}
					}
					return nil
				}),
				// highlighting.WithWrapperRenderer(func(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
				// 	lang, ok := c.Language()

				// 	wasJS := strings.Contains(string(lang), "js")
				// 	if entering {
				// 		if !ok {
				// 			w.WriteString("<pre><code")
				// 			if wasJS {
				// 				w.WriteString(" class='dontdisplaycheck'")
				// 			}
				// 			w.WriteString(">")
				// 			return
				// 		}
				// 		w.WriteString(`<div class="highlight">`)
				// 	} else {
				// 		if !ok {
				// 			w.WriteString("</pre></code>")
				// 			return
				// 		}
				// 		w.WriteString(`</div>`)
				// 	}
				// }),
			),
		),
	)
	return md
}

func NotebookParser() parser.Parser {
	return Md().Parser()
}

func NotebookRender() renderer.Renderer {
	r := Md().Renderer()
	r.AddOptions(html.WithUnsafe())
	return r
}
