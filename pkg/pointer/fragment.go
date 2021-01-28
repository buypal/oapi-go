package pointer

import (
	"errors"
	"fmt"
	"strings"
)

const (
	separator        = "/"
	escapedSeparator = "~1"
	tilde            = "~"
	escapedTilde     = "~0"
)

// Fragment ...
type Fragment []string

// NewFragment ...
func NewFragment(str string) (Fragment, error) {
	return parseFragment(str)
}

// IsEmpty is a utility function to check if the Pointer
// is empty / nil equivalent
func (p Fragment) IsEmpty() bool {
	return len(p) == 0
}

// Head returns the root of the Pointer
func (p Fragment) Head() (string, bool) {
	if len(p) == 0 {
		return "", false
	}
	return p[0], true
}

// Tail returns everything after the Pointer head
func (p Fragment) Tail() Fragment {
	return Fragment(p[1:])
}

// Last returns last element
func (p Fragment) Last() (string, bool) {
	if len(p) == 0 {
		return "", false
	}
	return p[len(p)-1], true
}

// Len returns count of fragments
func (p Fragment) Len() int {
	return len(p)
}

// Clone ...
func (p Fragment) Clone() Fragment {
	tmp := make(Fragment, len(p))
	copy(tmp, p)
	return tmp
}

// Replace will replace element in fragment
func (p Fragment) Replace(n int, element string) (Fragment, error) {
	if n > len(p) || n < 0 {
		return Fragment{}, errors.New("n out of range")
	}
	x := p.Clone()
	x[n] = element
	return x, nil
}

// String implements the stringer interface for Pointer,
// giving the escaped string
func (p Fragment) String() (str string) {
	for _, tok := range p {
		str += "/" + escapeToken(tok)
	}
	return
}

// Descendant returns a new pointer to a descendant of the current pointer
// parsing the input path into components
func (p Fragment) Descendant(path string) (Fragment, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	dpath, err := parseFragment(path)
	if err != nil {
		return p, err
	}

	if p.String() == "/" {
		return dpath, nil
	}

	return append(p, dpath...), nil
}

// RawDescendant extends the pointer with 1 or more path tokens
// The function itself is unsafe as it doesnt fully parse the input
// and assumes the user is directly managing the pointer
// This allows for much faster pointer management
func (p Fragment) RawDescendant(path ...string) Fragment {
	return append(p, path...)
}

func unescapeToken(tok string) string {
	tok = strings.Replace(tok, escapedSeparator, separator, -1)
	return strings.Replace(tok, escapedTilde, tilde, -1)
}

func escapeToken(tok string) string {
	tok = strings.Replace(tok, tilde, escapedTilde, -1)
	return strings.Replace(tok, separator, escapedSeparator, -1)
}

// The ABNF syntax of a JSON Pointer is:
// json-pointer    = *( "/" reference-token )
// reference-token = *( unescaped / escaped )
// unescaped       = %x00-2E / %x30-7D / %x7F-10FFFF
//    ; %x2F ('/') and %x7E ('~') are excluded from 'unescaped'
// escaped         = "~" ( "0" / "1" )
//   ; representing '~' and '/', respectively
func parseFragment(str string) ([]string, error) {
	if len(str) == 0 {
		return []string{}, nil
	}

	if str[0] != '/' {
		return nil, fmt.Errorf("non-empty references must begin with a '/' character")
	}
	str = str[1:]

	toks := strings.Split(str, separator)
	for i, t := range toks {
		toks[i] = unescapeToken(t)
	}
	return toks, nil
}
