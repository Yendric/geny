package headings

import (
	"reflect"
	"testing"
)

func TestBetween(t *testing.T) {
	tree := nest([]Heading{
		{Level: 1, ID: "a"},
		{Level: 2, ID: "b"},
		{Level: 4, ID: "c"},
		{Level: 3, ID: "d"},
		{Level: 2, ID: "e"},
	})

	want := Headings{
		{Level: 2, ID: "b", Children: Headings{{Level: 3, ID: "d"}}},
		{Level: 2, ID: "e"},
	}
	if got := tree.Between(2, 3); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}
