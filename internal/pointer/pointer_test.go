package pointer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPointer(t *testing.T) {
	doc, err := Parse("go://pointer.com/somthing#/Object")
	require.NoError(t, err)

	// require.Equal(t, doc.Fragment.Head(), "Object")
	require.Equal(t, doc.Scheme, "go")
	require.Equal(t, doc.Path, "/somthing")
}
