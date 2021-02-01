package container

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	c1, _ := Make(map[string]interface{}{
		"a": 1,
	})
	c2, _ := Make(map[string]interface{}{
		"b": 1,
		"c": map[string]interface{}{
			"d": 1,
		},
	})

	err := c1.Merge(c2, MergeStrict)
	require.NoError(t, err)

	m := c2.c.Data().(map[string]interface{})
	m["b"] = 2
	m["c"].(map[string]interface{})["d"] = 2

	require.Equal(t, c1.Path("a").Data(), 1.)
	require.Equal(t, c1.Path("b").Data(), 1.)
	require.Equal(t, c1.Path("c.d").Data(), 1.)
}

func TestMergeStrict(t *testing.T) {
	c1, _ := Make(map[string]interface{}{
		"a": 1,
	})
	c2, _ := Make(map[string]interface{}{
		"a": 1,
	})

	err := c1.Merge(c2, MergeStrict)
	require.Error(t, err)
}

func TestMergeDefault(t *testing.T) {
	c1, _ := Make(map[string]interface{}{
		"a": 1,
	})
	c2, _ := Make(map[string]interface{}{
		"a": 2,
	})

	err := c1.Merge(c2, MergeDefault)
	require.NoError(t, err)

	require.Equal(t, c1.c.Data().(map[string]interface{})["a"], 1.)
	require.Equal(t, c2.c.Data().(map[string]interface{})["a"], 2.)
}

func TestMergeOverride(t *testing.T) {
	c1, _ := Make(map[string]interface{}{
		"a": 1,
	})
	c2, _ := Make(map[string]interface{}{
		"a": 2,
	})

	err := c1.Merge(c2, MergeOverride)
	require.NoError(t, err)

	require.Equal(t, c1.c.Data().(map[string]interface{})["a"], 2.)
	require.Equal(t, c2.c.Data().(map[string]interface{})["a"], 2.)
}
