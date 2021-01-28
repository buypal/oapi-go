package container

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"
)

type SorterFn func(string, interface{}) (interface{}, error)

type Sorter struct {
	Container
	sorter SorterFn
}

func NewSortMarshaller(c Container, s SorterFn) Sorter {
	return Sorter{Container: c, sorter: s}
}

func (c Sorter) MarshalJSON() ([]byte, error) {
	data, err := c.sorter("json", c.c.Data())
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (c Sorter) MarshalIndentJSON(prefix string, indent string) ([]byte, error) {
	data, err := c.sorter("json", c.c.Data())
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, prefix, indent)
}

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
				MapItem: yaml.MapItem{
					Key:   k,
					Value: v,
				},
				Index: indexOf(k, order),
			})

		}
		sort.Sort(ms)
		return ms, nil
	}
}

// MapItem representation of one map item.
type MapItem struct {
	yaml.MapItem `json:",inline" yaml:",inline"`
	Index        int
}

// UnmarshalJSON for map item.
func (mi *MapItem) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	mi.Value = v
	return nil
}

// // UnmarshalYAML ...
// func (mi *MapItem) UnmarshalYAML(unmarshal func(v interface{}) error) error {
// 	var v interface{}
// 	if err := unmarshal(&v); err != nil {
// 		return err
// 	}
// 	mi.Value = v
// 	return nil
// }

// MapSlice of map items.
type MapSlice []MapItem

func (ms MapSlice) Len() int           { return len(ms) }
func (ms MapSlice) Less(i, j int) bool { return ms[i].Index < ms[j].Index }
func (ms MapSlice) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }

// MarshalJSON for map slice.
func (ms MapSlice) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{'{'})
	for i, mi := range ms {
		b, err := json.Marshal(&mi.Value)
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

// MarshalYAML for map slice.
func (ms MapSlice) MarshalYAML() (interface{}, error) {
	var m yaml.MapSlice
	for _, x := range ms {
		m = append(m, x.MapItem)
	}
	return m, nil
}

// UnmarshalJSON for map slice.
func (ms *MapSlice) UnmarshalJSON(b []byte) error {
	m := map[string]MapItem{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		*ms = append(*ms, MapItem{
			MapItem: yaml.MapItem{
				Key:   k,
				Value: v.Value,
			},
			Index: v.Index,
		})
	}
	return nil
}

// UnmarshalYAML ...
func (ms *MapSlice) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var v yaml.MapSlice
	if err := unmarshal(&v); err != nil {
		return err
	}
	for i, x := range v {
		*ms = append(*ms, MapItem{
			MapItem: yaml.MapItem{
				Key:   x.Key,
				Value: x.Value,
			},
			Index: i,
		})
	}
	return nil
}
