package resolver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/buypal/oapi-go/internal/container"
	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/buypal/oapi-go/internal/pointer"
)

// Entity is named entity.
type Entity struct {
	Entity spec.Entity
	Name   string
}

// Path returns dot path.
func (pe Entity) Path() string {
	return fmt.Sprintf("%s.%s", pe.Entity.Path(), pe.Name)
}

// Fragment return fragment path
func (pe Entity) Fragment() pointer.Fragment {
	f, _ := pe.Entity.Fragment().Descendant(pe.Name)
	return f
}

// IsScheme reports if entitiy is scheme
func (pe Pointer) IsScheme() bool {
	return pe.Entity.Entity == spec.SchemaKind
}

// EntityValue represent entity with value.
type EntityValue struct {
	Entity
	Value interface{}
}

// SetEntity sets at path the entity.
func SetEntity(cx container.Container, key string, entity EntityValue) (err error) {
	err = cx.SetP(entity.Path(), entity.Value)
	return
}

// ReplacePtr will check if key is path to pointer (aka ends .$ref),
// if so it will replace given pointer object with value.
func ReplacePtr(cx container.Container, key string, value interface{}) (err error) {
	path, ok := isref(key)
	if !ok {
		return errors.New("not a pointer")
	}
	err = cx.SetP(path, value)
	return
}

// ReplacePtrWithEmptyObject similar to ReplacePtr except it will replace
// with empty object. This might be handy for circular references.
func ReplacePtrWithEmptyObject(cx container.Container, key string) (err error) {
	return ReplacePtr(cx, key, container.New())
}

// UpdatePtr will set new pointer in given location.
func UpdatePtr(cx container.Container, key string, pp pointer.Pointer) (err error) {
	_, ok := isref(key)
	if !ok {
		return errors.New("not a pointer")
	}
	err = cx.SetP(key, pp.String())
	return
}

// UpdatePtrToLocal will set new pointer but unline UpdatePtr it will
// use local path. It set only fragment pointing to local document.
func UpdatePtrToLocal(cx container.Container, key string, f pointer.Fragment) (err error) {
	nptr := pointer.NewPointer()
	nptr.Fragment = f
	err = UpdatePtr(cx, key, nptr)
	return
}

func isref(s string) (string, bool) {
	ss := strings.Split(s, ".")
	if len(ss) == 0 {
		return s, false
	}
	last := ss[len(ss)-1]
	if last != "$ref" {
		return s, false
	}
	return strings.Join(ss[:len(ss)-1], "."), true
}
