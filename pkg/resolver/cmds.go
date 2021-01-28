package resolver

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/buypal/oapi-go/pkg/pointer"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

const prefix = "openapi"

// Scanner
// -----------------

type cmdsScanner struct {
	cmds cmdsMap
}

func newCmdsScanner() *cmdsScanner {
	return &cmdsScanner{
		cmds: make(cmdsMap),
	}
}

func (r *cmdsScanner) scan(pkg *packages.Package) (errs []error) {
	groups := []*ast.CommentGroup{}
	comments := []string{}

	for _, s := range pkg.Syntax {
		groups = append(groups, s.Comments...)
	}

	for _, g := range groups {
		comments = append(comments, parseCommentGroup(g)...)
	}

	var cc cmds
	for _, c := range comments {
		cmd, err := getCmd(pkg, c)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to parse openapi sytnax"))
			return
		}
		cc = append(cc, cmd)
	}

	r.cmds[pkg.PkgPath] = cc

	return
}

// parseCommentGroup will take comment group and returns
// parsed comments
func parseCommentGroup(gg *ast.CommentGroup) []string {
	cc := []string{}

	for _, cmt := range gg.List {
		c := cmt.Text[2:]
		hasPrefix := strings.HasPrefix(c, prefix)
		if !hasPrefix {
			continue
		}
		c = c[len(prefix):]
		c = strings.TrimRight(c, " \n")
		cc = append(cc, c)
	}

	return cc
}

// args are aguements to commands
type args []string

func newArguments(a []string) args {
	x := args{}
	for _, z := range a {
		x = append(x, strings.Trim(z, " "))
	}
	return x
}

func (a args) len() int {
	return len(a)
}

func (a args) get(index uint) (string, bool) {
	if int(index) > len(a)-1 {
		return "", false
	}
	return a[index], true
}

// Commands
// -----------------

// cmdKind is cmd kind
type cmdKind string

const (
	rootKind   cmdKind = ":root"
	schemaKind cmdKind = ":schema"
)

type commander interface {
	getCmd() cmdKind
	getArgs() args
}

func getCmd(pkg *packages.Package, comment string) (s commander, err error) {
	cmd := strings.Split(comment, " ")
	if len(cmd) == 0 {
		err = errors.New("failed to parse cmd")
		return
	}
	x := strings.Trim(cmd[0], " ")
	k := cmdKind(x)
	if x == "" {
		k = rootKind
	}
	r := cmdBase{
		cmd:    k,
		args:   newArguments(cmd[1:]),
		pkg:    pkg,
		origin: comment,
	}

	switch k {
	case schemaKind:
		s, err = newCmdSchema(r)
		return
	case rootKind:
		s, err = newCmdRoot(r)
		return
	default:
		err = errors.New(fmt.Sprintf("invalid open api cmd: %q", cmd))
		return
	}
}

type cmds []commander

func (cc cmds) pointers() pointer.Pointers {
	x := make(pointer.Pointers)
	for _, c := range cc {
		switch n := c.(type) {
		case cmdSchema:
			x[n.Ptr.String()] = n.Ptr
		}
	}
	return x
}

type cmdsMap map[string]cmds

func (cc cmdsMap) pointers() (pp pointer.Pointers) {
	pp = make(pointer.Pointers)
	for _, v := range cc {
		pp = pp.Merge(v.pointers())
	}
	return
}

// base command
type cmdBase struct {
	origin string
	cmd    cmdKind
	args   args
	pkg    *packages.Package
}

func (c cmdBase) getCmd() cmdKind {
	return c.cmd
}

func (c cmdBase) getArgs() args {
	return c.args
}

// root command parses the syntax '//openapi ....'
type cmdRoot struct {
	cmdBase
}

func newCmdRoot(cmd cmdBase) (cmdRoot, error) {
	s := cmdRoot{cmdBase: cmd}
	return s, nil
}

// schema command for schems exports
type cmdSchema struct {
	cmdBase
	Name string
	Ptr  pointer.Pointer
}

func newCmdSchema(cmd cmdBase) (s commander, err error) {
	name, nok := cmd.args.get(0)
	ptr, pok := cmd.args.get(1)
	sx := cmdSchema{cmdBase: cmd}
	rerr := errors.Errorf("failed to parse openapi:schema comment: %q", cmd.origin)

	invalid := func(a bool, s string) bool {
		return !a || len(s) == 0
	}

	makePtr := func(a string) (pointer.Pointer, error) {
		if strings.Contains(a, "://") || strings.Contains(a, "#") {
			return pointer.Parse(a)
		}
		return pointer.NewGoPointer(cmd.pkg.Types.Path(), a)
	}

	switch cmd.args.len() {
	case 1:
		if invalid(nok, name) {
			return nil, rerr
		}
		sx.Name = name
		sx.Ptr, err = makePtr(name)
		return sx, nil
	case 2:
		if invalid(nok, name) {
			return nil, rerr
		}
		if invalid(pok, ptr) {
			return nil, rerr
		}
		sx.Name = name
		sx.Ptr, err = makePtr(ptr)
		return sx, nil
	default:
		return nil, errors.New("invalid number of arguments")
	}
}
