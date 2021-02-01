// Package pointer implements IETF rfc6901
// JSON Pointers are a string syntax for
// identifying a specific value within a JavaScript Object Notation
// (JSON) document [RFC4627].  JSON Pointer is intended to be easily
// expressed in JSON string values as well as Uniform Resource
// Identifier (URI) [RFC3986] fragment identifiers.
//
// this package is intended to work like net/url from the go
// standard library.
// Original author https://github.com/qri-io/jsonpointer.
package pointer

import (
	"fmt"
	"net/url"
	"strconv"
)

const defaultPointerAllocationSize = 32

// Pointer represents a parsed JSON pointer
type Pointer struct {
	url.URL
	Fragment Fragment
}

// NewPointer creates a Pointer with a pre-allocated block of memory
// to avoid repeated slice expansions
func NewPointer() Pointer {
	return Pointer{
		Fragment: make([]string, 0, defaultPointerAllocationSize),
	}
}

// NewGoPointer ...
func NewGoPointer(pkg string, path string) (Pointer, error) {
	return Parse(fmt.Sprintf("go://%s#/%s", pkg, path))
}

// Parse parses str into a Pointer structure.
// str may be a pointer or a url string.
// If a url string, Parse will use the URL's fragment component
// (the bit after the '#' symbol)
func Parse(str string) (p Pointer, err error) {
	switch {
	case len(str) == 0:
		return Pointer{}, nil
	case str == "#":
		return Pointer{}, nil
	}

	var u *url.URL
	u, err = url.Parse(str)
	if err != nil {
		return
	}
	p.Fragment, err = NewFragment(u.Fragment)
	p.URL = *u
	return
}

func MustParse(str string) Pointer {
	p, err := Parse(str)
	if err != nil {
		panic(err)
	}
	return p
}

func (p *Pointer) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}
	t, err := strconv.Unquote(string(data))
	if err != nil {
		t = string(data)
	}
	x, err := Parse(t)
	if err != nil {
		return err
	}
	*p = x
	return nil
}

func (p *Pointer) MarshalJSON() ([]byte, error) {
	x := fmt.Sprintf("%q", p.String())
	return []byte(x), nil
}

func (p Pointer) String() string {
	p.URL.Fragment = ""
	return fmt.Sprintf("%s#%s", p.URL.String(), p.Fragment.String())
}

func (p Pointer) PkgPath() string {
	return fmt.Sprintf("%s%s", p.Hostname(), p.Path)
}

func (p Pointer) Clone() Pointer {
	f := p.Fragment.Clone()
	u, _ := url.Parse(p.URL.String())
	return Pointer{URL: *u, Fragment: f}
}

func (p Pointer) IsExternal() bool {
	return p.Scheme != ""
}

type Pointers map[string]Pointer

func NewPointers(pp []Pointer) Pointers {
	m := make(Pointers)
	for _, p := range pp {
		m[p.String()] = p
	}
	return m
}

// Merge will merge pointers
func (pp Pointers) Merge(px Pointers) Pointers {
	x := make(Pointers)
	for k, v := range pp {
		x[k] = v
	}
	for k, v := range px {
		x[k] = v
	}
	return x
}
