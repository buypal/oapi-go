package oapi

import (
	"errors"
	"fmt"
	"strings"

	"github.com/buypal/oapi-go/pkg/container"
	"github.com/buypal/oapi-go/pkg/oapi/spec"
	"github.com/buypal/oapi-go/pkg/pointer"
	"github.com/buypal/oapi-go/pkg/route"
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

// FlattenPath represents single reuqest path in oapi spec.
type FlattenPath struct {
	Method    string
	Path      string
	Key       string
	Operation container.Container
}

// FlattenPaths list of FlattenPath
type FlattenPaths []FlattenPath

// Match will match given path to given route
func (fp FlattenPaths) Match(pattern string) (xx []FlattenPath, err error) {
	for _, x := range fp {
		var matched bool
		matched, err = route.Match(pattern, x.Method, x.Path)
		if err != nil {
			return
		}
		if !matched {
			continue
		}
		xx = append(xx, x)
	}
	return
}

// Paths will parse paths in oapi container and return array of paths.
func Paths(cnt container.Container) (ff FlattenPaths, err error) {
	paths, err := cnt.Path("paths").ChildrenMap()
	if err != nil {
		return
	}

	for path, methods := range paths {
		var mx map[string]container.Container
		mx, err = methods.ChildrenMap()
		if err != nil {
			return
		}
		for m, obj := range mx {
			ff = append(ff, FlattenPath{
				Method:    m,
				Path:      path,
				Operation: obj,
				Key: container.SliceToDotPath([]string{
					"paths", path, m,
				}),
			})
		}
	}
	return
}

// SetPathsDefaults will iterate through paths and
// apply default on each path. This might be useful for
// supplying default headers or default responses.
func SetPathsDefaults(cnt container.Container, defops map[string]spec.Operation) (err error) {
	paths, err := Paths(cnt)
	if err != nil {
		return
	}

	for pattern, override := range defops {
		var ff FlattenPaths
		ff, err = paths.Match(pattern)
		if err != nil {
			return
		}
		for _, f := range ff {
			var nc, ov container.Container
			nc = f.Operation

			ov, err = container.Make(override)
			if err != nil {
				return
			}

			err = nc.Merge(ov, container.MergeDefault)
			if err != nil {
				return
			}

			err = cnt.SetP(f.Key, nc)
			if err != nil {
				return
			}
		}
	}

	return
}

// MergeWithRoot will merge container with root document.
func MergeWithRoot(root spec.OpenAPI, c container.Container) (err error) {
	r, err := container.Make(root)
	if err != nil {
		return
	}

	type mx struct {
		merge container.Merger
		key   string
	}

	for _, x := range []mx{
		{merge: container.MergeOverride, key: "info"},
		{merge: container.MergeStrict, key: "components"},
		{merge: container.MergeDefault, key: "paths"},
		{merge: container.MergeDefault, key: "externalDocs"},
		{merge: container.MergeDefault, key: "security"},
		{merge: container.MergeDefault, key: "servers"},
		{merge: container.MergeDefault, key: "tags"},
	} {
		z := c.Path(x.key)
		y := r.Path(x.key)

		if z.IsNil() && y.IsNil() {
			continue
		}

		s := container.New()

		err := s.Merge(z, x.merge)
		if err != nil {
			return err
		}

		err = s.Merge(y, x.merge)
		if err != nil {
			return err
		}

		p, _ := container.Make(map[string]interface{}{
			x.key: s.Data(),
		})

		err = c.Merge(p, container.MergeOverride)
		if err != nil {
			return err
		}
	}
	return nil
}
