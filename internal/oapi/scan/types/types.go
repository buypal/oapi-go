package types

import (
	"go/types"
	"strings"

	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/buypal/oapi-go/internal/pointer"
	"github.com/buypal/oapi-go/tag"

	"github.com/pkg/errors"
)

// collect types will wrap walk providing inital arguments
func collectTypes(t types.Object, m *pointmap) error {
	return walk(newPoint(t), t.Type(), path{}, m)
}

// elementer represents types such as slice, array, map or pointer
type elementer interface {
	types.Type
	Elem() types.Type
}

// if you are digging into this function, I tried to write docs
// but lets be honest nobody would understand that.
// But I'll give you some overview at least what it does.
//
// I tries to collect pointers (not go pointers but json pointers) ie:
//   - go://github.com/example/project/pkg/path#/Struct => struct {Filed1 string, Field2 string, Field3 github.com/some/pkg#/Struct}
//   - go://github.com/example/project/pkg/path#/Struct/Field1 => string
//   - go://github.com/example/project/pkg/path#/Struct/Field2 => string
//   - go://github.com/example/project/pkg/path#/Struct/Field3 => github.com/some/pkg#/Struct
//   - go://github.com/some/pkg#/Struct => struct {Filed1, Field2, ...etc}
//
// Later it is save to PtrType (last arg).
// First argument is pointer to be saved, type and map.
func walk(r point, t types.Type, p path, m *pointmap) (err error) {
	t = t.Underlying()
	m.append(t, r)

	p, visited := p.visit(t)
	if visited {
		return
	}

	type tptr struct {
		p point
		t types.Type
	}

	var next []tptr

	switch u := t.(type) {
	case *types.Struct:
		for i := 0; i < u.NumFields(); i++ {
			field := u.Field(i)

			if !field.Exported() {
				continue
			}

			next = append(next, tptr{
				p: withDescendant(r, field.Name()),
				t: field.Type(),
			})
		}

	case *types.Pointer:
		next = append(next, tptr{
			p: withPtr(r),
			t: u.Elem(),
		})

	case elementer:
		next = append(next, tptr{
			p: withArray(r),
			t: u.Elem(),
		})
	}

	for _, e := range next {
		err = walk(e.p, e.t, p, m)
		if err != nil {
			return
		}
		named, ok := e.t.(*types.Named)
		if ok {
			np := newPoint(named.Obj())
			err = walk(np, e.t, p, m)
		}
		if err != nil {
			return
		}
	}

	return
}

// path represents list of types going in direction.
// this is very much handy for cycle detection
type path []types.Type

func (tp path) String() string {
	var ss []string
	for _, x := range tp {
		ss = append(ss, x.String())
	}
	return strings.Join(ss, ", ")
}

func (tp path) has(t types.Type) bool {
	return tp.index(t) >= 0
}

func (tp path) index(t types.Type) int {
	for i, tx := range tp {
		if types.Identical(tx, t) {
			return i
		}
	}
	return -1
}

func (tp path) visit(t types.Type) (p path, visited bool) {
	if tp.has(t) {
		return tp, true
	}
	return append(tp, t), false
}

// type2schema will conver type to spec.Scheme
func type2schema(t types.Type, m pointmap, tp path, tg tag.Tag) (*spec.Schema, error) {
	t = t.Underlying()

	if tp.has(t) {
		return reference2schema(t, m, tp, tg)
	}
	tp = append(tp, t)

	switch u := t.(type) {

	// So struct will be converted into object
	case *types.Struct:
		if len(tp) == 1 {
			return struct2schema(u, m, tp)
		}
		return reference2schema(u, m, tp, tg)

	case *types.Array:
		return array2schema(u, m, tp, tg)

	// Basic type will just popoulate type and format
	case *types.Basic:
		return basic2schema(u.Kind(), tg)

	case *types.Map:
		return map2schema(u, m, tp, tg)

	case *types.Slice:
		return slice2schema(u, m, tp, tg)

	case *types.Pointer:
		return pointer2schema(u, m, tp, tg)

	// We cant marshal to schema
	default:
		err := errors.Errorf("type %q is not supported as a element for openapi", u.String())
		return nil, err
	}
}

func typeElement2schema(t elementer, m pointmap, tp path, tg tag.Tag) (s *spec.Schema, err error) {
	// Saying if element is slice, map, array, pointer and its inner element
	// is same as its outer element it is invalid type for scheme.
	// Exmple would be `type T *T` or `type T map[string]T`
	if types.Identical(t, t.Elem()) {
		return nil, errors.Errorf("type %q is self referencing identical type", t.String())
	}
	return type2schema(t.Elem(), m, tp, tg)
}

func struct2schema(t *types.Struct, m pointmap, tp path) (s *spec.Schema, err error) {
	s = &spec.Schema{}
	s.Type = spec.TypeObject
	s.Properties = make(map[string]*spec.Schema)

	fields, err := collectStructFields(t, path{})
	if err != nil {
		return
	}

	for _, x := range fields {
		var pschema *spec.Schema

		if len(x.tag.Type) != 0 {
			pschema, err = basicString2schema(x.tag.Type, x.tag)
		} else {
			switch z := x.field.Type().Underlying().(type) {
			// struct
			case *types.Struct:
				pschema, err = reference2schema(z, m, tp, x.tag)

			default:
				pschema, err = type2schema(z, m, tp, x.tag)
			}
		}

		if err != nil {
			return
		}

		name := x.field.Name()
		if len(x.tag.Name) > 0 {
			name = x.tag.Name
		}

		s.Properties[name] = pschema

		if pschema.Ref != nil {
			continue
		}

		pschema.Deprecated = x.tag.Deprecated
		pschema.ReadOnly = x.tag.ReadOnly
		pschema.WriteOnly = x.tag.WriteOnly
		pschema.Format = x.tag.Format

		if x.tag.Nullable != nil {
			pschema.Nullable = *x.tag.Nullable
		}

		if x.tag.Required {
			s.Required = append(s.Required, name)
		}
	}

	return
}

func map2schema(t *types.Map, m pointmap, tp path, tg tag.Tag) (s *spec.Schema, err error) {
	sch, err := typeElement2schema(t, m, tp, tg)
	if err != nil {
		return
	}

	s = &spec.Schema{}
	s.Type = spec.TypeObject
	s.AdditionalProperties = sch
	s.Nullable = true

	s.MinProperties = tg.MinProps
	s.MaxProperties = tg.MaxProps
	if tg.Nullable != nil {
		s.Nullable = *tg.Nullable
	}

	return
}

func slice2schema(t *types.Slice, m pointmap, tp path, tg tag.Tag) (s *spec.Schema, err error) {
	sch, err := typeElement2schema(t, m, tp, tg)
	if err != nil {
		return
	}

	s = &spec.Schema{}
	s.Type = spec.TypeArray
	s.Items = sch
	s.Nullable = true

	s.MinItems = tg.MinItems
	s.MaxItems = tg.MaxItems
	s.UniqueItems = tg.UniqItems
	if tg.Nullable != nil {
		s.Nullable = *tg.Nullable
	}

	return
}

func array2schema(t *types.Array, m pointmap, tp path, tg tag.Tag) (s *spec.Schema, err error) {
	sch, err := typeElement2schema(t, m, tp, tg)
	if err != nil {
		return
	}

	s = &spec.Schema{}
	s.Type = spec.TypeArray
	s.Items = sch

	s.MinItems = tg.MinItems
	s.MaxItems = tg.MaxItems

	if t.Len() > 0 {
		x := t.Len()
		s.MaxItems = &x
	}

	s.UniqueItems = tg.UniqItems
	if tg.Nullable != nil {
		s.Nullable = *tg.Nullable
	}

	return
}

func pointer2schema(t *types.Pointer, m pointmap, tp path, tg tag.Tag) (s *spec.Schema, err error) {
	s, err = typeElement2schema(t, m, tp, tg)
	if err != nil {
		return
	}
	s.Nullable = true

	if tg.Nullable != nil {
		s.Nullable = *tg.Nullable
	}

	if s.Ref == nil {
		return s, nil
	}

	if !s.Nullable {
		return s, nil
	}

	s.Nullable = false

	s = spec.OneSchema(s, &spec.Schema{
		Type:     spec.TypeObject,
		Nullable: true,
	})

	return s, nil
}

var refStdMapping = map[string]spec.Schema{
	"go://time#/Time": {
		Type: spec.TypeString,
	},
}

// AddRefOverride allows to add default override
func AddRefOverride(p pointer.Pointer, s spec.Schema) {
	refStdMapping[p.String()] = s
}

func reference2schema(t types.Type, m pointmap, tp path, tg tag.Tag) (s *spec.Schema, err error) {
	ptr, ok := m.pick(t)
	if !ok {
		return nil, errors.New("failed to resolve sturct")
	}
	if sch, ok := refStdMapping[ptr.String()]; ok {
		return &sch, nil
	}
	s = &spec.Schema{}
	s.Ref = &ptr.Pointer
	return s, nil
}

// uint8  : 0 to 255
// uint16 : 0 to 65535
// uint32 : 0 to 4294967295
// uint64 : 0 to 18446744073709551615
// int8   : -128 to 127
// int16  : -32768 to 32767
// int32  : -2147483648 to 2147483647
// int64  : -9223372036854775808 to 9223372036854775807

// Go2schemaType indicates type will transfer golang basic type to open api supported type.
func basic2schema(t types.BasicKind, tg tag.Tag) (s *spec.Schema, err error) {
	zero := 0.
	switch t {
	case types.Float32:
		s = spec.Float32Property()
	case types.Float64:
		s = spec.Float64Property()
	case types.Uint, types.Uint32:
		s = spec.IntFmtProperty("int32")
		s.Minimum = &zero
	case types.Uint8:
		s = spec.IntFmtProperty("int32")
		s.Minimum = &zero
	case types.Uint16:
		s = spec.IntFmtProperty("int32")
		s.Minimum = &zero
	case types.Uint64:
		s = spec.IntFmtProperty("int64")
		s.Minimum = &zero
	case types.Int, types.Int32:
		s = spec.IntFmtProperty("int32")
	case types.Int8:
		s = spec.IntFmtProperty("int32")
	case types.Int16:
		s = spec.IntFmtProperty("int32")
	case types.Int64:
		s = spec.IntFmtProperty("int64")
	case types.Bool:
		s = spec.BooleanProperty()
	case types.String:
		s = spec.StringProperty()
	default:
		return nil, errors.New("invalid basic kind")
	}

	switch s.Type {
	case spec.TypeNumber, spec.TypeInteger:
		if tg.Min != nil {
			s.Minimum = tg.Min
		}
		if tg.Max != nil {
			s.Maximum = tg.Max
		}
		if tg.EMin != nil {
			s.Minimum = tg.EMin
			s.ExclusiveMinimum = true
		}
		if tg.EMax != nil {
			s.Maximum = tg.EMax
			s.ExclusiveMaximum = true
		}
		if tg.MulOf != nil {
			s.MultipleOf = tg.MulOf
		}
	case spec.TypeString:
		if tg.MaxLen != nil {
			s.MaxLength = tg.MaxLen
		}
		if tg.MinLen != nil {
			s.MinLength = tg.MinLen
		}
		s.Pattern = tg.Pattern
	}

	return s, nil
}

func basicString2schema(t string, tg tag.Tag) (s *spec.Schema, err error) {
	y := strings.Trim(t, " ")
	switch y {
	case "string":
		return basic2schema(types.String, tg)
	case "float32", "float":
		return basic2schema(types.Float32, tg)
	case "float64", "double":
		return basic2schema(types.Float64, tg)
	case "uint":
		return basic2schema(types.Uint, tg)
	case "uint8":
		return basic2schema(types.Uint8, tg)
	case "uint16":
		return basic2schema(types.Uint16, tg)
	case "uint32":
		return basic2schema(types.Uint32, tg)
	case "uint64":
		return basic2schema(types.Uint64, tg)
	case "int":
		return basic2schema(types.Int, tg)
	case "int8":
		return basic2schema(types.Int8, tg)
	case "int16":
		return basic2schema(types.Int16, tg)
	case "int32", "integer":
		return basic2schema(types.Int32, tg)
	case "int64":
		return basic2schema(types.Int64, tg)
	case "bool":
		return basic2schema(types.Bool, tg)
	case "base64":
		s, err = basic2schema(types.String, tg)
		s.Format = "binary"
		return s, err
	case "uuid":
		s, err = basic2schema(types.String, tg)
		s.Format = "uuid"
		return s, err
	case "password":
		s, err = basic2schema(types.String, tg)
		s.Format = "password"
		return s, err
	case "number":
		s = &spec.Schema{Type: spec.TypeNumber}
		return s, nil
	case "object":
		s = &spec.Schema{Type: spec.TypeObject}
		return s, nil
	default:
		return nil, errors.Errorf("invalid type %q", t)
	}
}

func castInlineStruct(t *types.Var, tg tag.Tag) (*types.Struct, bool) {
	tx := t.Type().Underlying()
	// check for pointer
	if p, ok := tx.(*types.Pointer); ok {
		tx = p.Elem().Underlying()
	}
	st, ok := tx.(*types.Struct)
	if ok {
		if tg.Inline != nil {
			ok = *tg.Inline
		} else {
			ok = t.Embedded()
		}
	}
	return st, ok
}

type structField struct {
	field *types.Var
	tag   tag.Tag
}

func collectStructFields(t *types.Struct, p path) (arr []structField, err error) {
	// prevent cycles
	if p.has(t) {
		return []structField{}, nil
	}
	p = append(p, t)
	for i := 0; i < t.NumFields(); i++ {
		x := t.Field(i)
		var tx tag.Tag
		tx, err = tag.Parse(t.Tag(i))
		if err != nil {
			return
		}
		if tx.Ignore || !x.Exported() {
			continue
		}
		st, ok := castInlineStruct(x, tx)
		if ok {
			var z []structField
			z, err = collectStructFields(st, p)
			if err != nil {
				return
			}
			arr = append(arr, z...)
		} else {
			arr = append(arr, structField{
				field: x,
				tag:   tx,
			})
		}
	}
	return
}
