package main

import (
	"path/filepath"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
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
			HighlightInst,
			CommentInst,
			emoji.Emoji,
			&mermaid.Extender{},
			callout.CalloutExtention,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
		),
	)
	return md
}

func NotebookParser() parser.Parser {
	return Md().Parser()
}

func NotebookRender() renderer.Renderer {
	return Md().Renderer()
}
