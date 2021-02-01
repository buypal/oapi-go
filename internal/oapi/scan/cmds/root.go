package cmds

// CmdRoot root command parses the syntax '//openapi ....'.
// Dummy schema for now.
type CmdRoot struct {
	CmdBase
}

// NewCmdRoot creates new root schema
func NewCmdRoot(cmd CmdBase) (CmdRoot, error) {
	s := CmdRoot{CmdBase: cmd}
	return s, nil
}
