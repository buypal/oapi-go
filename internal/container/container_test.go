package container

import (
	"bytes"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c := New()
	c.Data().(map[string]interface{})["a"] = 1

	require.Equal(t, c.Path("a").Data(), 1)
	require.Equal(t, c.ExistsP("a"), true)
	require.Equal(t, c.IsNil(), false)
	require.Equal(t, c.FilePath(), "")
}

func TestMake(t *testing.T) {
	c, err := Make(map[string]interface{}{
		"a": 1,
	})
	require.NoError(t, err)
	require.Equal(t, c.Path("a").Data(), 1.)

	c, err = Make(c)
	require.NoError(t, err)
	require.Equal(t, c.Path("a").Data(), 1.)

	x := 1
	_, err2 := Make(map[*int]interface{}{
		&x: 1,
	})
	require.Error(t, err2)
}

func TestReadJSON(t *testing.T) {
	c, err := Make(map[string]interface{}{
		"a": 1,
	})
	require.NoError(t, err)
	require.Equal(t, c.Path("a").Data(), 1.)

	x, err := ReadJSON(c.Bytes())
	require.NoError(t, err)
	require.Equal(t, len(x.c.Data().(map[string]interface{})), 1)

	bf := bytes.NewBuffer(c.Bytes())
	bf.Write([]byte(`bad json`))

	_, err = ReadJSON(bf.Bytes())
	require.Error(t, err)
}

func TestReadYAML(t *testing.T) {
	c, err := Make(map[string]interface{}{
		"b": 1,
	})
	require.NoError(t, err)
	require.Equal(t, c.Path("b").Data(), 1.)
	bb, _ := c.MarshalYAML()

	x, err := ReadYAML(bb)
	require.NoError(t, err)
	require.Equal(t, len(x.c.Data().(map[string]interface{})), 1)

	bf := bytes.NewBuffer(bb)
	bf.Write([]byte(`			bad yaml@#$~`))

	_, err = ReadYAML(bf.Bytes())
	require.Error(t, err)
}

func TestReadFile(t *testing.T) {
	x, err := ReadFile("./testdata/a.json")
	require.NoError(t, err)
	require.Equal(t, len(x.c.Data().(map[string]interface{})), 1)
	require.Equal(t, x.path, "./testdata/a.json")

	x, err = ReadFile("./testdata/b.yaml")
	require.NoError(t, err)
	require.Equal(t, len(x.c.Data().(map[string]interface{})), 1)
	require.Equal(t, x.path, "./testdata/b.yaml")

	x, err = ReadFile("./testdata/b.test")
	require.Error(t, err)
}

func TestReadFiles(t *testing.T) {
	x, err := ReadDir("./testdata")
	require.NoError(t, err)
	var paths []string
	for _, p := range x.Sort() {
		paths = append(paths, p.path)
	}
	require.Equal(t, paths, []string{
		"testdata/a.json",
		"testdata/b.yaml",
	})

	c, err := x.Merge(MergeStrict)
	require.NoError(t, err)

	require.Equal(t, string(c.Bytes()), `{"a":1,"b":1}`)
}

func TestChildrenMap(t *testing.T) {
	x, err := ReadDir("./testdata")
	require.NoError(t, err)

	c, err := x.Sort().Merge(MergeStrict)
	require.NoError(t, err)

	m, err := c.ChildrenMap()
	require.NoError(t, err)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	require.Equal(t, keys, []string{"a", "b"})
}

func TestFlatten(t *testing.T) {
	x, err := ReadDir("./testdata")
	require.NoError(t, err)

	c, err := x.Sort().Merge(MergeStrict)
	require.NoError(t, err)

	m, err := c.Flatten()
	require.NoError(t, err)

	require.Equal(t, m, map[string]interface{}{"a": 1., "b": 1.})
}

func TestSetP(t *testing.T) {
	c := New()
	c.Data().(map[string]interface{})["a"] = 1

	require.Equal(t, c.Path("a").Data(), 1)

	err := c.SetP("a", 2)
	require.NoError(t, err)

	require.Equal(t, c.Path("a").Data(), 2.)
}
