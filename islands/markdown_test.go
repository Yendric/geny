package islands

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func render(t *testing.T, reg *Registry, src string) (string, *Collector) {
	t.Helper()
	md := goldmark.New(goldmark.WithExtensions(Extension), goldmark.WithRendererOptions(html.WithUnsafe()))
	pc, collector := NewParserContext(reg)
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf, parser.WithContext(pc)); err != nil {
		t.Fatal(err)
	}
	return buf.String(), collector
}

func TestMarkdownBlockIslandWithFallback(t *testing.T) {
	reg := testRegistry(t, "Counter")
	out, c := render(t, reg, "# Title\n\n<Counter client:visible start={5}>\nLoading *counter*\n</Counter>\n\nAfter\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}

	want := `<geny-island component="Counter" client="visible" props="{&#34;start&#34;:5}">
<p>Loading <em>counter</em></p>
</geny-island>
<p>After</p>`
	if !strings.Contains(out, want) {
		t.Errorf("got:\n%s\nwant to contain:\n%s", out, want)
	}
	if u := c.Usages(); len(u) != 1 || u[0].Line != 3 {
		t.Errorf("usages: %+v", u)
	}
}

func TestMarkdownSelfClosingBlockIsland(t *testing.T) {
	reg := testRegistry(t, "Counter")
	out, c := render(t, reg, "<Counter\n  start={1}\n/>\nnext paragraph\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `props="{&#34;start&#34;:1}">`+"\n</geny-island>\n<p>next paragraph</p>") {
		t.Errorf("got:\n%s", out)
	}
}

func TestMarkdownInlineIsland(t *testing.T) {
	reg := testRegistry(t, "Counter")
	out, c := render(t, reg, "Count: <Counter start={0} /> clicks\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}
	want := `<p>Count: <geny-island component="Counter" client="load" props="{&#34;start&#34;:0}"></geny-island> clicks</p>`
	if !strings.Contains(out, want) {
		t.Errorf("got:\n%s", out)
	}
}

func TestMarkdownLeavesOtherHTMLAlone(t *testing.T) {
	reg := testRegistry(t, "Counter")
	out, c := render(t, reg, "<div class=\"x\">\nhi\n</div>\n\nText <span>inline</span>\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "geny-island") || !strings.Contains(out, `<div class="x">`) || !strings.Contains(out, "<span>inline</span>") {
		t.Errorf("got:\n%s", out)
	}
	if len(c.Usages()) != 0 {
		t.Errorf("unexpected usages %+v", c.Usages())
	}
}

func TestMarkdownIslandInsideCodeIsLiteral(t *testing.T) {
	reg := testRegistry(t, "Counter")
	out, c := render(t, reg, "```\n<Counter />\n```\n\nUse `<Counter />` here.\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "geny-island") {
		t.Errorf("islands in code must stay literal, got:\n%s", out)
	}
}

func TestMarkdownErrors(t *testing.T) {
	reg := testRegistry(t, "Counter")
	tests := map[string]string{
		"unclosed":      "<Counter>\nfallback\n",
		"inline open":   "Text <Counter> more\n",
		"bad directive": "<Counter client:never />\n",
	}
	for name, src := range tests {
		_, c := render(t, reg, src)
		if c.Err() == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func componentRegistry(t *testing.T) *Registry {
	t.Helper()
	reg := testRegistry(t, "Counter")
	box := template.Must(template.New("Box.html").Parse(`<div class="{{ .Props.kind }}">{{ .Children }}</div>`))
	if err := reg.SetComponents(map[string]*template.Template{"Box": box}); err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestMarkdownComponentWithChildren(t *testing.T) {
	out, c := render(t, componentRegistry(t), "<Box kind=\"note\">\nHello *world*\n</Box>\n\nAfter\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}
	if want := "<div class=\"note\"><p>Hello <em>world</em></p>\n</div><p>After</p>"; !strings.Contains(out, want) {
		t.Errorf("got:\n%s\nwant to contain:\n%s", out, want)
	}
	if len(c.Usages()) != 0 {
		t.Errorf("components must not be island usages: %+v", c.Usages())
	}
}

func TestMarkdownInlineComponent(t *testing.T) {
	out, c := render(t, componentRegistry(t), "See <Box kind=\"tip\" /> here\n")
	if err := c.Err(); err != nil {
		t.Fatal(err)
	}
	if want := `<p>See <div class="tip"></div> here</p>`; !strings.Contains(out, want) || len(c.Usages()) != 0 {
		t.Errorf("got:\n%s\nusages: %+v", out, c.Usages())
	}
}

func TestComponentIslandNameClash(t *testing.T) {
	err := testRegistry(t, "Box").SetComponents(map[string]*template.Template{"Box": template.New("Box.html")})
	if err == nil {
		t.Error("expected name clash error")
	}
}
