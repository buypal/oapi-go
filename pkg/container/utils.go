package container

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/pkg/errors"
)

// ExtractKey will extract keys from container
func ExtractKey(c Container, key string) (pp MapSlice, err error) {
	ff, err := c.Flatten()
	if err != nil {
		return nil, err
	}
	var ii int
	for k, v := range ff {
		x := strings.Split(k, ".")
		if len(x) == 0 {
			continue
		}
		if x[len(x)-1] != key {
			continue
		}
		pp = append(pp, MapItem{
			Key:   k,
			Val:   v,
			Index: ii,
		})
		ii++
	}
	return pp, nil
}

// SliceToDotPath converts slice into dot path
func SliceToDotPath(path []string) string {
	hierarchy := make([]string, len(path))
	for i, v := range path {
		v = strings.Replace(v, ".", "~1", -1)
		v = strings.Replace(v, "~", "~0", -1)
		hierarchy[i] = v
	}
	return strings.Join(hierarchy, ".")
}

// clone will clone given interface by marshaling and
// unmarshaling via json
func clone(c interface{}) (interface{}, error) {
	if x, ok := cast(c); ok {
		c = x.Data()
	}
	bb, err := json.Marshal(mmap(c))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clone interface{}")
	}
	var mm interface{}
	json.Unmarshal(bb, &mm) // no need for error check
	return mm, nil
}

// cast will cast interface to container
func cast(c interface{}) (Container, bool) {
	if x, ok := c.(Container); ok {
		return x, true
	}
	if x, ok := c.(*Container); ok && x != nil {
		return *x, true
	}
	return Container{}, false
}

// wrap will wrap given value into container only
// if it is not already one
func wrap(c interface{}) Container {
	if x, ok := cast(c); ok {
		return x
	}
	return Container{
		c: gabs.Wrap(c),
	}
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
