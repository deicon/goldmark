package extension

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
	"regexp"
	"sort"
)

type frontMatterParser struct{}

func (f frontMatterParser) Trigger() []byte {
	return []byte("---")
}

func (f frontMatterParser) Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	line, _ := reader.PeekLine()
	line = bytes.TrimRight(line, "\r\n")
	if matched, _ := regexp.Match("^-{3,}$", line); matched {
		reader.AdvanceLine()

		// read all lines and interpret as yaml
		var buf bytes.Buffer
		var y map[interface{}]interface{}
		for {
			line, _ = reader.PeekLine()
			if matched, _ := regexp.Match("^-{3,}$", line); !matched {
				buf.Write(line)
			} else {
				break
			}
			reader.AdvanceLine()
		}
		if err := yaml.Unmarshal(buf.Bytes(), &y); err != nil {
			fmt.Errorf("Unable to parse Frontmatter as YAML %s", err.Error())
			return &ast.FrontmatterStart{MetaData: nil, Status: err.Error()}, parser.NoChildren
		}
		// reset position

		return &ast.FrontmatterStart{MetaData: y}, parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (f frontMatterParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	// all parsing done in Open already
	return parser.Close
}

func (f frontMatterParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (f frontMatterParser) CanInterruptParagraph() bool {
	return false
}

func (f frontMatterParser) CanAcceptIndentedLine() bool {
	return false
}

var DefaultFronMatterParser = &frontMatterParser{}

func NewFrontMatterParser() parser.BlockParser {
	return DefaultFronMatterParser
}

type FrontmatterHTMLRenderer struct {
	Config html.Config
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *FrontmatterHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFrontmatterStart, r.renderFrontmatterStart)
}

func (r *FrontmatterHTMLRenderer) renderFrontmatterStart(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		w.WriteString("<!--\n")
		if node, ok := n.(*ast.FrontmatterStart); ok {
			if node.Status != "" || node.MetaData == nil {
				w.WriteString(node.Status)
			} else {
				keys := make([]string, 0, len(node.MetaData))
				for key := range node.MetaData {
					keys = append(keys, fmt.Sprintf("%v", key))
				}
				sort.Strings(keys)
				for _, key := range keys {
					w.WriteString(fmt.Sprintf("%s: %v\n", key, node.MetaData[key]))
				}
			}
		}
	} else {
		w.WriteString("-->\n")
	}
	return gast.WalkContinue, nil
}

func NewFrontmatterHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &FrontmatterHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (f frontMatterParser) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewFrontMatterParser(), 99)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(NewFrontmatterHTMLRenderer(), 99)))
}
