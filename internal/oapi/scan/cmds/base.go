package cmds

import "golang.org/x/tools/go/packages"

// CmdBase is base command
type CmdBase struct {
	origin string
	cmd    CmdKind
	args   Args
	pkg    *packages.Package
}

// GetCmd returns matched command.
func (c CmdBase) GetCmd() CmdKind {
	return c.cmd
}

// GetArgs returns arguments of command
func (c CmdBase) GetArgs() Args {
	return c.args
}
