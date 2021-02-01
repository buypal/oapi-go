package cmds

import (
	"strings"

	"github.com/buypal/oapi-go/internal/pointer"
	"github.com/pkg/errors"
)

// CmdSchema is command responsible for allowing
// schema exports to be possible.
// It has simple syntax: //oapi:schema <uri>,
// causing schema to be exported at root of document
// usually components.*.
type CmdSchema struct {
	CmdBase
	Name string
	Ptr  pointer.Pointer
}

// NewCmdSchema creates new command schema
func NewCmdSchema(cmd CmdBase) (s Commander, err error) {
	name, nok := cmd.args.Get(0)
	ptr, pok := cmd.args.Get(1)
	sx := CmdSchema{CmdBase: cmd}
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

	switch cmd.args.Len() {
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
