package container

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	c := New()
	c.Data().(map[string]interface{})["a"] = 1
	c.Data().(map[string]interface{})["b"] = 1

	marhsaller := NewSortMarshaller(c, SortMapMarhsaler([]string{"a", "b"}))
	x, err := marhsaller.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(x), `{"a":1,"b":1}`)

	x, err = marhsaller.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, string(x), "a: 1\nb: 1\n")

	x, err = marhsaller.MarshalIndentJSON("", "")
	require.NoError(t, err)
	require.Equal(t, string(x), "{\n\"a\": 1,\n\"b\": 1\n}")

	marhsaller = NewSortMarshaller(c, SortMapMarhsaler([]string{"b", "a"}))
	x, err = marhsaller.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(x), `{"b":1,"a":1}`)

	x, err = marhsaller.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, string(x), "b: 1\na: 1\n")

}
