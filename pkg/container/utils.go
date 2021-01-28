package container

import (
	"encoding/json"
	"fmt"
	"strings"
)

type KeyVal struct {
	Key string
	Val interface{}
}

type KeyVals []KeyVal

func (vv KeyVals) Values() (vx []interface{}) {
	for _, v := range vv {
		vx = append(vx, v)
	}
	return
}

func (vv KeyVals) Strings() (vx []string) {
	for _, v := range vv {
		s, ok := v.Val.(string)
		if !ok {
			continue
		}
		vx = append(vx, s)
	}
	return
}

// ExtractKey will extract keys from container
func ExtractKey(c Container, key string) (pp KeyVals, err error) {
	ff, err := c.Flatten()
	if err != nil {
		return nil, err
	}
	for k, v := range ff {
		x := strings.Split(k, ".")
		if len(x) == 0 {
			continue
		}
		if x[len(x)-1] != key {
			continue
		}
		pp = append(pp, KeyVal{
			Key: k,
			Val: v,
		})
	}
	return pp, nil
}

// clone will clone given interface by marshaling and
// unmarshaling via json
func clone(c interface{}) (interface{}, error) {
	if x, ok := c.(Container); ok {
		c = x.c.Data()
	}
	if x, ok := c.(*Container); ok {
		c = x.c.Data()
	}
	bb, err := json.Marshal(mmap(c))
	if err != nil {
		return nil, err
	}
	var mm interface{}
	err = json.Unmarshal(bb, &mm)
	if err != nil {
		return nil, err
	}
	return mm, nil
}

// mmap walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func mmap(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = mmap(v2)
			default:
				m[fmt.Sprint(k)] = mmap(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = mmap(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = mmap(v2)
		}
	}

	return v
}
