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

type Entity struct {
	Entity spec.Entity
	Name   string
}

func (pe Entity) Path() string {
	return fmt.Sprintf("%s.%s", pe.Entity.Path(), pe.Name)
}

func (pe Entity) Fragment() pointer.Fragment {
	f, _ := pe.Entity.Fragment().Descendant(pe.Name)
	return f
}

func (pe Pointer) IsScheme() bool {
	return pe.Entity.Entity == spec.SchemaKind
}

// EntityValue ...
type EntityValue struct {
	Entity
	Value interface{}
}

// SetEntity ...
func SetEntity(cx container.Container, key string, entity EntityValue) (err error) {
	_, err = cx.SetP(entity.Path(), entity.Value)
	return
}

// ReplacePtr ...
func ReplacePtr(cx container.Container, key string, value interface{}) (err error) {
	path, ok := isref(key)
	if !ok {
		return errors.New("not a pointer")
	}
	_, err = cx.SetP(path, value)
	return
}

// ReplacePtrWithEmptyObject ...
func ReplacePtrWithEmptyObject(cx container.Container, key string) (err error) {
	return ReplacePtr(cx, key, container.EmptyObject())
}

// UpdatePtr ...
func UpdatePtr(cx container.Container, key string, pp pointer.Pointer) (err error) {
	_, ok := isref(key)
	if !ok {
		return errors.New("not a pointer")
	}
	_, err = cx.SetP(key, pp.String())
	return
}

// UpdatePtrToLocal ...
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

// FlattenPath represents single reuqest path merged with method
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

func SliceToDotPath(path []string) string {
	hierarchy := make([]string, len(path))
	for i, v := range path {
		v = strings.Replace(v, ".", "~1", -1)
		v = strings.Replace(v, "~", "~0", -1)
		hierarchy[i] = v
	}
	return strings.Join(hierarchy, ".")
}

// Paths will return array of paths
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
				Key: SliceToDotPath([]string{
					"paths", path, m,
				}),
			})
		}
	}
	return
}

// SetPathsDefaults ...
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

			_, err = cnt.SetP(f.Key, nc)
			if err != nil {
				return
			}
		}
	}

	return
}

// MergeWithRoot will merge container with root
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

		p := container.Wrap(map[string]interface{}{
			x.key: s.Data(),
		})

		err = c.Merge(p, container.MergeOverride)
		if err != nil {
			return err
		}
	}
	return nil
}
