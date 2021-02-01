package resolver

import (
	"github.com/buypal/oapi-go/internal/container"
	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/buypal/oapi-go/internal/pointer"
)

var zero = container.Zero()

// Pointer is specifiying what component we are dealing with
type Pointer struct {
	Entity
	pointer.Pointer
}

// Exports as list of components
type Exports []Pointer

// Get returns conponent by pointer
func (e Exports) Get(p pointer.Pointer) (Pointer, bool) {
	for _, x := range e {
		if x.String() == p.String() {
			return x, true
		}
	}
	return Pointer{}, false
}

// Fn allows to resolve entities, $refs into actual schemes
type Fn func(pointer.Pointer) (spec.Entiter, error)

func (r Fn) call(p pointer.Pointer) (container.Container, error) {
	x, err := r(p)
	if err != nil {
		return zero, err
	}
	return container.Make(x)
}

// Resolve will resolve all references (pointers) in given scheme
func Resolve(c container.Container, exp Exports, fn Fn) (container.Container, error) {
	r := &resolver{
		con: c.Clone(),
		exp: exp,
		res: fn,
	}
	return r.resolve()
}

type resolver struct {
	con container.Container
	exp Exports
	res Fn
}

func (r resolver) resolve() (container.Container, error) {
	return r.iterator(r.con, path{}, 0)
}

func (r resolver) iterator(cx container.Container, pp path, dept int) (container.Container, error) {
	refs, err := container.ExtractKey(cx, "$ref")
	if err != nil {
		return cx, err
	}

	for _, v := range refs {
		s, ok := v.Val.(string)
		if !ok {
			continue
		}
		p, err := pointer.Parse(s)
		if err != nil {
			return cx, err
		}
		if !p.IsExternal() {
			continue
		}

		if pp.has(p) {
			err = r.setRecursive(cx, p, v.Key)
			if err != nil {
				return zero, err
			}
			continue
		}

		nc, err := r.res.call(p)
		if err != nil {
			return zero, err
		}

		nc, err = r.iterator(nc, append(pp, p), dept+1)
		if err != nil {
			return zero, err
		}

		err = r.set(cx, p, v.Key, nc)
		if err != nil {
			return zero, err
		}
	}
	return cx, nil
}

// set will set in container key to be a value, if
// exported sets as global value
func (r *resolver) set(cx container.Container, ptr pointer.Pointer, key string, value container.Container) (err error) {
	ep, ok := r.exp.Get(ptr)
	if !ok {
		return ReplacePtr(cx, key, value)
	}
	e := EntityValue{
		Entity: ep.Entity,
		Value:  value,
	}
	err = SetEntity(r.con, key, e)
	if err != nil {
		return
	}
	return UpdatePtrToLocal(cx, key, e.Fragment())
}

func (r *resolver) setRecursive(cx container.Container, ptr pointer.Pointer, key string) (err error) {
	ep, ok := r.exp.Get(ptr)
	if ok && ep.IsScheme() {
		return UpdatePtrToLocal(cx, key, ep.Entity.Fragment())
	}
	return ReplacePtrWithEmptyObject(cx, key)
}

type path []pointer.Pointer

func (p path) has(px pointer.Pointer) bool {
	for i := len(p) - 1; i >= 0; i-- {
		el := p[i]
		if el.String() == px.String() {
			return true
		}
	}
	return false
}
