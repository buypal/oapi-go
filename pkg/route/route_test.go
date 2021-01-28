package route

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoute(t *testing.T) {
	require.False(t, MustMatch("/[]foo/{av}/1/*2", "", "/ok"))
	require.False(t, MustMatch("/[]foo/{av}/1/*2", "", "/[]"))
	require.False(t, MustMatch("/[]foo/{av}/1/*2", "", "/[]foo"))
	require.False(t, MustMatch("/[]foo/{av}/1/*2", "", "/[]foo/{xxxx}/1"))
	require.False(t, MustMatch("/[]foo/{av}/1/*2", "", "/[]foo/{xxxx}/1/3"))
	require.False(t, MustMatch("/[]foo/{av}/1/*2", "", "/[]foo/{xxxx}/2"))
	require.False(t, MustMatch("GET:/foo", "POST", "/foo"))

	require.True(t, MustMatch("GET:/foo", "", "/foo"))
	require.True(t, MustMatch("POST:/foo", "", "/foo"))
	require.True(t, MustMatch("GET:/foo", "get", "/foo"))
	require.True(t, MustMatch("GET:/foo", "GET", "/foo"))
	require.True(t, MustMatch("/[]foo/{av}/1/*2", "", "/[]foo/{123}/1/3/2"))
	require.True(t, MustMatch("/[]foo/{av}/1/valid/*2", "", "/[]foo/{x}/1/valid/2"))
	require.True(t, MustMatch("/[]foo/{av}/1/valid/*2", "", "/[]foo/{xxxx}/1/valid/2"))
	require.True(t, MustMatch("GET:/[]foo/{av}/1/valid/*2", "", "/[]foo/{xxxx}/1/valid/2"))
}
