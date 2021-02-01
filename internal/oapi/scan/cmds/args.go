package cmds

import "strings"

// Args are aguements to commands
type Args []string

func newArguments(a []string) Args {
	x := Args{}
	for _, z := range a {
		x = append(x, strings.Trim(z, " "))
	}
	return x
}

// Len return length of arguments
func (a Args) Len() int {
	return len(a)
}

// Get will return value and true if at given
// index argument exists.
func (a Args) Get(index uint) (string, bool) {
	if int(index) > len(a)-1 {
		return "", false
	}
	return a[index], true
}
