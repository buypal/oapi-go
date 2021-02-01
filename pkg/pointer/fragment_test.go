package pointer

import (
	"fmt"
	"strings"
	"testing"
)

func Example() {
	// parse a json pointer. Pointers can also be url fragments
	// the following are equivelent pointers:
	// "/foo/bar/baz/1"
	// "#/foo/bar/baz/1"
	// "http://example.com/document.json#/foo/bar/baz/1"
	ptr, _ := NewFragment("/foo/bar/baz/1")

	// evaluate the pointer against the document
	// evaluation always starts at the root of the document
	got, _ := ptr.Tail()

	fmt.Println(got)
	// Output: 1
}

// doc pulled from spec:
var docBytes = []byte(`{
  "foo": ["bar", "baz"],
  "": 0,
  "a/b": 1,
  "c%d": 2,
  "e^f": 3,
  "g|h": 4,
  "i\\j": 5,
  "k\"l": 6,
  " ": 7,
  "m~n": 8
}`)

func TestParse(t *testing.T) {
	cases := []struct {
		raw    string
		parsed string
		err    string
	}{
		{"#/", "/", ""},
		{"#/foo", "/foo", ""},
		{"#/foo/", "/foo/", ""},

		{"://", "", "missing protocol scheme"},
		{"#7", "", "non-empty references must begin with a '/' character"},
		{"", "", ""},
		{"https://example.com#", "", ""},
	}

	for i, c := range cases {
		got, err := Parse(c.raw)
		if !(err == nil && c.err == "" || err != nil && strings.Contains(err.Error(), c.err)) {
			t.Errorf("case %d error mismatch. expected: '%s', got: '%s'", i, c.err, err)
			continue
		}

		if c.err == "" && got.Fragment.String() != c.parsed {
			t.Errorf("case %d string output mismatch: expected: '%s', got: '%s'", i, c.parsed, got.Fragment.String())
			continue
		}
	}
}

func TestDescendent(t *testing.T) {
	cases := []struct {
		parent string
		path   string
		parsed string
		err    string
	}{
		{"/", "0", "/0", ""},
		{"/0", "0", "/0/0", ""},
		{"/foo", "0", "/foo/0", ""},
		{"/foo", "0", "/foo/0", ""},
		{"/foo/0", "0", "/foo/0/0", ""},
	}

	for i, c := range cases {
		p, err := NewFragment(c.parent)
		if err != nil {
			t.Errorf("case %d error parsing parent: %s", i, err.Error())
			continue
		}

		desc, err := p.Descendant(c.path)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch. expected: %s, got: %s", i, c.err, err)
			continue
		}

		if desc.String() != c.parsed {
			t.Errorf("case %d: expected: %s, got: %s", i, c.parsed, desc.String())
			continue
		}
	}
}

func TestEscapeToken(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"/abc~1/~/0/~0/", "/abc~1/~/0/~0/"},
	}
	for i, c := range cases {
		got := unescapeToken(escapeToken(c.input))
		if got != c.output {
			t.Errorf("case %d result mismatch.  expected: '%s', got: '%s'", i, c.output, got)
		}
	}
}
