package cmds

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/buypal/oapi-go/internal/pointer"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

const prefix = "openapi"

// CmdKind is command kind
type CmdKind string

const (
	// RootKind is root command
	RootKind CmdKind = ":root"
	// SchemaKind is schema command
	SchemaKind CmdKind = ":schema"
)

// Commander interface allows to unite and cast given comments
// at need.
type Commander interface {
	GetCmd() CmdKind
	GetArgs() Args
}

// ParseCommentGroup will take ast comment group and returns
// parsed comments as a strings.
func ParseCommentGroup(gg *ast.CommentGroup) []string {
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

// Parse will parse pacage with given comment, returning command.
func Parse(pkg *packages.Package, comment string) (s Commander, err error) {
	cmd := strings.Split(comment, " ")
	if len(cmd) == 0 {
		err = errors.New("failed to parse cmd")
		return
	}
	x := strings.Trim(cmd[0], " ")
	k := CmdKind(x)
	if x == "" {
		k = RootKind
	}
	r := CmdBase{
		cmd:    k,
		args:   newArguments(cmd[1:]),
		pkg:    pkg,
		origin: comment,
	}

	switch k {
	case SchemaKind:
		s, err = NewCmdSchema(r)
		return
	case RootKind:
		s, err = NewCmdRoot(r)
		return
	default:
		err = errors.New(fmt.Sprintf("invalid open api cmd: %q", cmd))
		return
	}
}

// List is list of commands
type List []Commander

// Pointers will list of pointers
func (cc List) Pointers() pointer.Pointers {
	x := make(pointer.Pointers)
	for _, c := range cc {
		switch n := c.(type) {
		case CmdSchema:
			x[n.Ptr.String()] = n.Ptr
		}
	}
	return x
}

// Map is a map of commands
type Map map[string]List

// Pointers  retuns list of pointers
func (cc Map) Pointers() (pp pointer.Pointers) {
	pp = make(pointer.Pointers)
	for _, v := range cc {
		pp = pp.Merge(v.Pointers())
	}
	return
}
