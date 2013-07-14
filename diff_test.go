package diff

import (
	"fmt"
	"reflect"
	"testing"
)

type stringDiff struct {
	a    string
	b    string
	lcsa []string
	lcsb []string
}

func (d *stringDiff) Lengths() (int, int) { return len(d.a), len(d.b) }
func (d *stringDiff) Equal(i, j int) bool { return d.a[i] == d.b[j] }
func (d *stringDiff) Common(i, j, n int) {
	d.lcsa = append(d.lcsa, d.a[i:i+n])
	d.lcsb = append(d.lcsb, d.b[j:j+n])
}

func TestDiff(t *testing.T) {
	var tests = []struct {
		a     string
		b     string
		lcs   []string
		edits int
	}{
		{"", "", []string{""}, 0},
		{"", "a", []string{""}, 1},
		{"a", "", []string{""}, 1},
		{"a", "a", []string{"a"}, 0},
		{"ab", "a", []string{"a", ""}, 1},
		{"a", "ab", []string{"a", ""}, 1},
		{"abc", "abc", []string{"abc"}, 0},
		{"abc", "ac", []string{"a", "c"}, 1},
		{"bc", "abc", []string{"bc"}, 1},
		{"ab", "abc", []string{"ab", ""}, 1},
		{"abcdefghijk", "abxyzcdxyzfgxyzj", []string{"ab", "cd", "fg", "j", ""}, 13},
	}

	for i, test := range tests {
		d := &stringDiff{a: test.a, b: test.b}
		edits := Diff(d)
		if !reflect.DeepEqual(d.lcsa, test.lcs) {
			t.Errorf("test %d lcsa:\nwant %q\nhave %q\n", i, test.lcs, d.lcsa)
		}
		if !reflect.DeepEqual(d.lcsb, test.lcs) {
			t.Errorf("test %d lcsb:\nwant %q\nhave %q\n", i, test.lcs, d.lcsb)
		}
		if edits != test.edits {
			t.Errorf("test %d number of edits:\nwant %d\nhave %d\n", i, test.edits, edits)
		}
	}
}

func TestSideBySide(t *testing.T) {
	var tests = []struct {
		a     []string
		b     []string
		lines []SideBySideLine
	}{{
		[]string{},
		[]string{},
		nil,
	}, {
		[]string{"a", "b"},
		[]string{"a", "c"},
		[]SideBySideLine{{"a", "a", NoChange}, {"b", "c", Changed}},
	}, {
		[]string{"a", "b"},
		[]string{"b"},
		[]SideBySideLine{{"a", "", Deleted}, {"b", "b", NoChange}},
	}, {
		[]string{"a", "b"},
		[]string{"a", "c", "b"},
		[]SideBySideLine{{"a", "a", NoChange}, {"", "c", Added}, {"b", "b", NoChange}},
	}, {
		[]string{"a"},
		[]string{"b", "c"},
		[]SideBySideLine{{"a", "b", Changed}, {"", "c", Added}},
	}}
	for i, test := range tests {
		lines := SideBySide(test.a, test.b)
		if !reflect.DeepEqual(lines, test.lines) {
			t.Errorf("test %d:\nwant %q\nhave %q\n", i, test.lines, lines)
		}
	}
}

func ExampleAnnotate() {
	files := [][]string{
		{"0a", "0b", "0c"},
		{"1a", "0a", "1b", "0c", "1c"},
		{"0a", "1b", "0c", "2a", "2b", "1c"},
	}
	lines := Annotate(nil, files[0], 0)
	lines = Annotate(lines, files[1], 1)
	lines = Annotate(lines, files[2], 2)
	for _, l := range lines {
		fmt.Println(l.Version, l.Text)
	}
	// Output:
	// 0 0a
	// 1 1b
	// 0 0c
	// 2 2a
	// 2 2b
	// 1 1c
}
