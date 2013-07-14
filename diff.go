// Package diff implements a diff algorithm for finding the longest common
// subsequence of two sequences.
package diff

// Constants used for SideBySide diffs.
const (
	NoChange = iota
	Added
	Deleted
	Changed
)

// A type that implements diff.Interface can be passed to the Diff function to
// find the largest common subsequence in two sequences.
type Interface interface {
	// Lengths returns the respective lengths of the left and right sequences.
	Lengths() (int, int)
	// Equal returns whether the elements at index i in the left and index
	// j in the right sequence are equal.
	Equal(i, j int) bool
	// Common is called to report a part of the longest common subsequence:
	// left[i:i+n] == right[j:j+n]. It is called for every part of the LCS
	// from top to bottom. To allow for finalizing actions the last call
	// always reports the tailing elements of left and right that are
	// equal, even if there are none, in which case n==0.
	Common(i, j, n int)
}

// Diff computes the longest common subsequence of two sequences. It returns
// the length of the edit script (number of inserts and deletes) needed to go
// from one sequence to the other. The algorithm is described here:
// http://neil.fraser.name/software/diff_match_patch/myers.pdf.
func Diff(data Interface) int {
	var vs [][]int
	n, m := data.Lengths()
	for d := 0; d <= m+n; d++ {
		v := make([]int, 2*(n+m)+3) // at least 3 diagonals for the initial step
		if d == 0 {
			v[1] = 0
		} else {
			copy(v, vs[d-1])
		}
		for k := -d; k <= d; k += 2 {
			var x int
			K := len(v)/2 + k
			if k == -d || (k != d && v[K-1] < v[K+1]) {
				x = v[K+1]
			} else {
				x = v[K-1] + 1
			}
			y := x - k
			for x < n && y < m && data.Equal(x, y) {
				x++
				y++
			}
			v[K] = x
			if x >= n && y >= m {
				vs = append(vs, v)
				common(data, vs, n, m, len(vs)-1)
				return len(vs) - 1
			}
		}
		vs = append(vs, v)
	}
	panic("diff: no path found")
}

func common(data Interface, vs [][]int, x1, y1, d int) {
	v := vs[d]
	k := x1 - y1
	K := len(v)/2 + k

	var x, y, xm int
	if insert := k == -d || (k != d && v[K-1] < v[K+1]); insert {
		x = v[K+1]
		y = x - (k + 1)
		xm = x
	} else {
		x = v[K-1]
		y = x - (k - 1)
		xm = x + 1
	}
	if d > 0 {
		common(data, vs, x, y, d-1)
	}
	if n := x1 - xm; n > 0 || d == len(vs)-1 {
		data.Common(xm, y1-n, n)
	}
}

// Side-by-side diff

// SideBySideLine represents a line in a side-by-side diff.
type SideBySideLine struct {
	Left  string // Left line, empty string if Type==Added.
	Right string // Right line, empty string if Type==Deleted.
	Type  int    // NoChange, Added, Deleted, Changed
}

// SideBySide computes a side-by-side diff of two sets of lines.
func SideBySide(a, b []string) []SideBySideLine {
	d := &sideBySide{a: a, b: b}
	Diff(d)
	return d.lines
}

type sideBySide struct {
	a     []string
	b     []string
	i     int
	j     int
	lines []SideBySideLine
}

func (d *sideBySide) Lengths() (int, int) { return len(d.a), len(d.b) }
func (d *sideBySide) Equal(i, j int) bool { return d.a[i] == d.b[j] }
func (d *sideBySide) Common(i, j, n int) {
	for d.i < i || d.j < j {
		var line SideBySideLine
		switch {
		case d.i >= i:
			line.Type = Added
		case d.j >= j:
			line.Type = Deleted
		default:
			line.Type = Changed
		}
		if d.i < i {
			line.Left = d.a[d.i]
			d.i++
		}
		if d.j < j {
			line.Right = d.b[d.j]
			d.j++
		}
		d.lines = append(d.lines, line)
	}
	for ; n > 0; n-- {
		d.lines = append(d.lines, SideBySideLine{
			Left:  d.a[d.i],
			Right: d.b[d.j],
			Type:  NoChange,
		})
		d.i++
		d.j++
	}
}

// Annotated diff

// AnnotatedLine represents a line in an annotated diff.
type AnnotatedLine struct {
	Text    string
	Version int
}

// Annotate computes an annotated diff from a to b, that is, it maintains for
// each line the version in which it was introduced. version is an int
// representing b's version.
func Annotate(a []AnnotatedLine, b []string, version int) []AnnotatedLine {
	d := &annotate{a: a, b: b, version: version}
	Diff(d)
	return d.lines
}

type annotate struct {
	a       []AnnotatedLine
	b       []string
	j       int
	version int
	lines   []AnnotatedLine
}

func (d *annotate) Lengths() (int, int) { return len(d.a), len(d.b) }
func (d *annotate) Equal(i, j int) bool { return d.a[i].Text == d.b[j] }
func (d *annotate) Common(i, j, n int) {
	for d.j < j {
		d.lines = append(d.lines, AnnotatedLine{d.b[d.j], d.version})
		d.j++
	}
	d.lines = append(d.lines, d.a[i:i+n]...)
	d.j += n
}
