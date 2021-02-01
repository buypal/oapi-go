package container

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestExtractKey(t *testing.T) {
	c, _ := Make(map[string]interface{}{
		"a": map[string]interface{}{
			"xx": 1,
			"b":  1,
		},
		"xx": 2,
	})

	ss, err := ExtractKey(c, "xx")
	require.NoError(t, err)

	sort.Slice(ss, func(i, j int) bool {
		return len(ss[i].Key) > len(ss[j].Key)
	})

	require.Equal(t, ss[0].Key, "a.xx")
	require.Equal(t, ss[0].Val, 1.)

	require.Equal(t, ss[1].Key, "xx")
	require.Equal(t, ss[1].Val, 2.)

	require.Len(t, ss, 2)
}

func TestMmap(t *testing.T) {
	y := "a:\n  1: 1\n  2: 2\n"

	var v interface{}
	err := yaml.Unmarshal([]byte(y), &v)
	require.NoError(t, err)

	_, ok := v.(map[interface{}]interface{})
	require.True(t, ok)

	v = mmap(v)
	_, ok = v.(map[string]interface{})
	require.True(t, ok)
}
