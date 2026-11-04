package posts

import (
	"bytes"
	"html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"

	"github.com/SailmanSeeulater/automorph/internal/media"
)

// md renders GitHub-flavoured Markdown (tables, strikethrough, autolinks, task lists).
// Raw HTML in the source is dropped, which is goldmark's safe default.
var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(
		// Priority 100 runs before goldmark's built-in image renderer (1000).
		renderer.WithNodeRenderers(util.Prioritized(imageRenderer{}, 100)),
	),
)

// The post column is at most 800px wide; browsers use this to pick a srcset width.
const imageSizes = "(max-width: 800px) 100vw, 800px"

func RenderMarkdown(source string) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// imageRenderer turns ![alt](/media/…-1920.webp) into a responsive, lazy-loaded <img>.
type imageRenderer struct{}

func (imageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, renderImage)
}

func renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	src := string(n.Destination)
	if gmhtml.IsDangerousURL(n.Destination) {
		src = ""
	}

	w.WriteString(`<img src="` + html.EscapeString(src) + `"`)
	if key, width, ok := media.ParseURL(src); ok {
		w.WriteString(` srcset="` + html.EscapeString(media.Srcset(key, width)) + `"`)
		w.WriteString(` sizes="` + imageSizes + `"`)
	}
	w.WriteString(` alt="` + html.EscapeString(altText(n, source)) + `"`)
	if len(n.Title) > 0 {
		w.WriteString(` title="` + html.EscapeString(string(n.Title)) + `"`)
	}
	w.WriteString(` loading="lazy" decoding="async">`)
	return ast.WalkSkipChildren, nil
}

// altText collects the plain text inside ![…], dropping any formatting.
func altText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if t, ok := c.(*ast.Text); ok && entering {
			buf.Write(t.Segment.Value(source))
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}
