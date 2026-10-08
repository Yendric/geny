package islands

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Collector struct {
	registry *Registry
	usages   []Usage
	errs     []error
}

func (c *Collector) Usages() []Usage {
	return c.usages
}

func (c *Collector) Err() error {
	return errors.Join(c.errs...)
}

func (c *Collector) record(island Island, src []byte, offset int) {
	c.usages = append(c.usages, Usage{Island: island, Line: lineAt(src, offset)})
}

func (c *Collector) fail(src []byte, offset int, err error) {
	c.errs = append(c.errs, fmt.Errorf("line %d: %w", lineAt(src, offset), err))
}

func lineAt(src []byte, offset int) int {
	return bytes.Count(src[:offset], []byte{'\n'}) + 1
}

var collectorKey = parser.NewContextKey()

func NewParserContext(reg *Registry) (parser.Context, *Collector) {
	c := &Collector{registry: reg}
	pc := parser.NewContext()
	pc.Set(collectorKey, c)
	return pc, c
}

func collectorFrom(pc parser.Context) (*Collector, bool) {
	c, ok := pc.Get(collectorKey).(*Collector)
	return c, ok && c.registry != nil
}

var (
	KindBlock  = ast.NewNodeKind("IslandBlock")
	KindInline = ast.NewNodeKind("IslandInline")
)

type blockNode struct {
	ast.BaseBlock
	// underlying island component
	island Island
	// source offset past opening tag
	tagEnd int
	// source offset for opening tags '<'
	tagOffset int
	// whether the tag is self closing
	selfClosing bool
	// closed once corresponding closing tag has been consumed
	closed bool
}

func (n *blockNode) Kind() ast.NodeKind { return KindBlock }

func (n *blockNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Name": n.island.Name}, nil)
}

type inlineNode struct {
	ast.BaseInline
	island Island
}

func (n *inlineNode) Kind() ast.NodeKind { return KindInline }

func (n *inlineNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Name": n.island.Name}, nil)
}

type blockParser struct{}

func (blockParser) Trigger() []byte { return []byte{'<'} }

func (blockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	c, ok := collectorFrom(pc)
	if !ok {
		return nil, parser.NoChildren
	}
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || line[pos] != '<' {
		return nil, parser.NoChildren
	}

	src := reader.Source()
	start := segment.Start + pos
	t, err := parseTag(src, start, c.registry)
	if errors.Is(err, errNotIsland) {
		return nil, parser.NoChildren
	}
	if err != nil {
		c.fail(src, start, err)
		return nil, parser.NoChildren
	}
	if !util.IsBlank(restOfLine(src, t.end)) {
		return nil, parser.NoChildren
	}

	c.record(t.island, src, start)
	return &blockNode{island: t.island, tagEnd: t.end, tagOffset: start, selfClosing: t.selfClosing}, parser.NoChildren
}

func restOfLine(src []byte, pos int) []byte {
	end := bytes.IndexByte(src[pos:], '\n')
	if end < 0 {
		return src[pos:]
	}
	return src[pos : pos+end]
}

func (blockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	n, ok := node.(*blockNode)
	if !ok {
		return parser.Close
	}
	line, segment := reader.PeekLine()
	if segment.Start < n.tagEnd {
		return parser.Continue | parser.NoChildren
	}
	if n.selfClosing {
		return parser.Close
	}
	if string(bytes.TrimSpace(line)) == "</"+n.island.Name+">" {
		n.closed = true
		reader.Advance(len(bytes.TrimRight(line, "\r\n")))
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

func (blockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	n, ok := node.(*blockNode)
	if !ok || n.selfClosing || n.closed {
		return
	}
	if c, ok := collectorFrom(pc); ok {
		c.fail(reader.Source(), n.tagOffset, fmt.Errorf("island %s: missing </%s>", n.island.Name, n.island.Name))
	}
}

func (blockParser) CanInterruptParagraph() bool { return true }

func (blockParser) CanAcceptIndentedLine() bool { return false }

type inlineParser struct{}

func (inlineParser) Trigger() []byte { return []byte{'<'} }

func (inlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	c, ok := collectorFrom(pc)
	if !ok {
		return nil
	}
	_, segment := block.PeekLine()
	src := block.Source()
	t, err := parseTag(src, segment.Start, c.registry)
	if errors.Is(err, errNotIsland) {
		return nil
	}
	if err != nil {
		c.fail(src, segment.Start, err)
		return nil
	}
	if !t.selfClosing {
		c.fail(src, segment.Start, fmt.Errorf("island %s: inline islands must be self-closing, or put the tag on its own line", t.island.Name))
		return nil
	}

	c.record(t.island, src, segment.Start)
	block.Advance(t.end - segment.Start)
	return &inlineNode{island: t.island}
}

type nodeRenderer struct{}

func (nodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindBlock, renderBlock)
	reg.Register(KindInline, renderInline)
}

func renderBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n, ok := node.(*blockNode)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %s", node.Kind())
	}
	if entering {
		_, _ = w.WriteString(n.island.openTag())
		_ = w.WriteByte('\n')
	} else {
		_, _ = w.WriteString(closeTag)
		_ = w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

func renderInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n, ok := node.(*inlineNode)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %s", node.Kind())
	}
	if entering {
		_, _ = w.WriteString(string(n.island.HTML()))
	}
	return ast.WalkContinue, nil
}

type extension struct{}

var Extension goldmark.Extender = extension{}

func (extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(blockParser{}, 850)),
		parser.WithInlineParsers(util.Prioritized(inlineParser{}, 350)),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(nodeRenderer{}, 100)))
}
