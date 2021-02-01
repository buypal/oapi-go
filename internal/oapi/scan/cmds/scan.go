package cmds

import (
	"go/ast"

	"github.com/buypal/oapi-go/internal/oapi/resolver"
	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// Scanner allow to scan go code for commands.
// Commands are specific comments allowing to alter
// behaviour of scanner.
type Scanner struct {
	Commands Map
}

// NewScanner creates new scanner.
func NewScanner() *Scanner {
	return &Scanner{
		Commands: make(Map),
	}
}

// Scan will scan package and store info.
func (r *Scanner) Scan(pkg *packages.Package) (err error) {
	groups := []*ast.CommentGroup{}
	comments := []string{}
	for _, s := range pkg.Syntax {
		groups = append(groups, s.Comments...)
	}
	for _, g := range groups {
		comments = append(comments, ParseCommentGroup(g)...)
	}
	var cc List
	for _, c := range comments {
		x, err := Parse(pkg, c)
		if err != nil {
			return errors.Wrapf(err, "failed to parse openapi sytnax")
		}
		cc = append(cc, x)
	}
	r.Commands[pkg.PkgPath] = cc
	return
}

// ExportedComponents will provide exported components in a form
// of resolver.Exports.
func (r *Scanner) ExportedComponents() (exports resolver.Exports, err error) {
	var cc List
	for _, cmd := range r.Commands {
		cc = append(cc, cmd...)
	}

	for _, cmd := range cc {
		switch x := cmd.(type) {
		case CmdSchema:
			if _, ok := exports.Get(x.Ptr); ok {
				err = errors.Errorf("schema of type %q already registered", x.Name)
				return
			}
			entity := resolver.Entity{
				Entity: spec.SchemaKind,
				Name:   x.Name,
			}
			exp := resolver.Pointer{
				Pointer: x.Ptr,
				Entity:  entity,
			}
			exports = append(exports, exp)
		}
	}
	return
}
