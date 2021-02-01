package container

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"
)

// SorterFn is sorting function
type SorterFn func(string, interface{}) (interface{}, error)

// Sorter wraps container and allows and provides new Masrshaling interface.
type Sorter struct {
	Container
	sorter SorterFn
}

// NewSortMarshaller makes new marshaller with sorting provided
// in second argument.
func NewSortMarshaller(c Container, s SorterFn) Sorter {
	return Sorter{Container: c, sorter: s}
}

// MarshalJSON returns the JSON encoding of container.
func (c Sorter) MarshalJSON() ([]byte, error) {
	data, err := c.sorter("json", c.c.Data())
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

// MarshalIndentJSON is like Marshal but applies Indent to format the output.
// Each JSON element in the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to the indentation nesting.
func (c Sorter) MarshalIndentJSON(prefix string, indent string) ([]byte, error) {
	data, err := c.sorter("json", c.c.Data())
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, prefix, indent)
}

// MarshalYAML serializes the value provided into a YAML document. The structure
// of the generated document will reflect the structure of the value itself.
func (c Sorter) MarshalYAML() ([]byte, error) {
	data, err := c.sorter("yaml", c.c.Data())
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(data)
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

// SortMapMarhsaler is one of sorting functions allowing to
// sorty by "order" which is supplied as a first argument.
func SortMapMarhsaler(order []string) SorterFn {
	return func(_ string, data interface{}) (interface{}, error) {
		if order == nil {
			return data, nil
		}
		dd, ok := data.(map[string]interface{})
		if !ok {
			return data, nil
		}
		var ms MapSlice
		for k, v := range dd {
			ms = append(ms, MapItem{
				Key:   k,
				Val:   v,
				Index: indexOf(k, order),
			})

		}
		sort.Sort(ms)
		return ms, nil
	}
}

// MapItem representation of one map item.
// MapItem is used internally by yaml pkg to provide
// key/value structure allowing object to be serialozed as array.
// By setting Index you are setting position of key in object.
type MapItem struct {
	Key   string
	Val   interface{}
	Index int
}

// MapSlice of map items.
type MapSlice []MapItem

// Values returns values of MapSlice
func (ms MapSlice) Values() (vx []interface{}) {
	for _, v := range ms.Sort() {
		vx = append(vx, v.Val)
	}
	return
}

// sort interface support

func (ms MapSlice) Len() int           { return len(ms) }
func (ms MapSlice) Less(i, j int) bool { return ms[i].Index < ms[j].Index }
func (ms MapSlice) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }

// Sort will sort items by index and return copy
func (ms MapSlice) Sort() MapSlice {
	a := make(MapSlice, len(ms))
	copy(a, ms)
	sort.Sort(a)
	return a
}

// MarshalJSON for map slice.
func (ms MapSlice) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{'{'})
	for i, mi := range ms.Sort() {
		b, err := json.Marshal(&mi.Val)
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("%q:", fmt.Sprintf("%v", mi.Key)))
		buf.Write(b)
		if i < len(ms)-1 {
			buf.Write([]byte{','})
		}
	}
	buf.Write([]byte{'}'})
	return buf.Bytes(), nil
}

// MarshalYAML will marshal yaml to MapSlice.
func (ms MapSlice) MarshalYAML() (interface{}, error) {
	var m yaml.MapSlice
	for _, x := range ms.Sort() {
		m = append(m, yaml.MapItem{
			Key:   x.Key,
			Value: x.Val,
		})
	}
	return m, nil
}
