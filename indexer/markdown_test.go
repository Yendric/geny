package indexer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/headings"
	"github.com/Yendric/geny/islands"
)

func parse(t *testing.T, src string) parsedMarkdown {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Counter.tsx"), []byte("export default function X() { return null }"), 0o644); err != nil {
		t.Fatal(err)
	}
	reg, err := islands.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := New(common.Config{}).parseMdFile(reg, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestHeadingsTree(t *testing.T) {
	parsed := parse(t, "# Guide\n\n## Install `geny`\n\n### Linux\n\n## Usage {#use}\n\n## Usage\n\n<Counter>\n## Hidden\n</Counter>\n")

	want := headings.Headings{{
		Level: 1, ID: "guide", Text: "Guide", HTML: "Guide",
		Children: headings.Headings{
			{
				Level: 2, ID: "install-geny", Text: "Install geny", HTML: "Install <code>geny</code>",
				Children: headings.Headings{{Level: 3, ID: "linux", Text: "Linux", HTML: "Linux"}},
			},
			{Level: 2, ID: "use", Text: "Usage", HTML: "Usage"},
			{Level: 2, ID: "usage", Text: "Usage", HTML: "Usage"},
		},
	}}
	if !reflect.DeepEqual(parsed.headings, want) {
		t.Errorf("got %+v\nwant %+v", parsed.headings, want)
	}
	if !strings.Contains(string(parsed.html), `<h2 id="use">Usage</h2>`) {
		t.Errorf("custom id not rendered:\n%s", parsed.html)
	}
}

func TestHeadingsDuplicateIDs(t *testing.T) {
	parsed := parse(t, "## Notes\n\n## Notes\n")
	if parsed.headings[0].ID != "notes" || parsed.headings[1].ID != "notes-1" {
		t.Errorf("ids: %+v", parsed.headings)
	}
}
