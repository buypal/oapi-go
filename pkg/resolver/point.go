package resolver

import (
	"go/types"
	"sort"
	"strings"

	"github.com/buypal/oapi-go/pkg/logging"
	"github.com/buypal/oapi-go/pkg/pointer"
	"golang.org/x/tools/go/types/typeutil"
)

type point struct {
	pointer.Pointer
}

func newPoint(root types.Object) point {
	path := root.Name()
	ptr, _ := pointer.NewGoPointer(root.Pkg().Path(), path)
	return point{Pointer: ptr}
}

func withDescendant(ptr point, name string) point {
	x := ptr.Clone()
	fx, _ := x.Fragment.Descendant(name)
	x.Fragment = fx
	return x
}

func withArray(ptr point) point {
	x := ptr.Clone()
	fx, _ := x.Fragment.Descendant("[]")
	x.Fragment = fx
	return x
}

func withPtr(ptr point) point {
	x := ptr.Clone()
	fx, _ := x.Fragment.Descendant("*")
	x.Fragment = fx
	return x
}

func (p point) Clone() point {
	return point{Pointer: p.Pointer.Clone()}
}

func (p point) String() string {
	return p.Pointer.String()
}

func (p point) equal(p2 point) bool {
	return p.Pointer.String() == p2.Pointer.String()
}

type points []point

// Len is part of sort.Interface.
func (s points) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s points) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s points) Less(i, j int) bool {
	return s[i].Fragment.Len() < s[j].Fragment.Len()
}

func (s points) filter(compare func(p1 point) bool) (pp points) {
	for _, p := range s {
		if compare(p) {
			pp = append(pp, p)
		}
	}
	return pp
}

func (s points) has(p2 point) bool {
	return s.filter(p2.equal).Len() > 0
}

func (s points) String() string {
	var ss []string
	for _, sx := range s {
		ss = append(ss, sx.String())
	}
	return strings.Join(ss, ", ")
}

type pointmap struct {
	m typeutil.Map
}

func (tp pointmap) len(t types.Type, r point) bool {
	return tp.at(t).Len() > 0
}

func (tp *pointmap) append(t types.Type, p1 point) points {
	pp, _ := tp.m.At(t).(points)
	if pp.has(p1) {
		return pp
	}
	pp = append(pp, p1)
	tp.m.Set(t, pp)
	return pp
}

func (tp pointmap) at(t types.Type) points {
	if t == nil {
		return points{}
	}
	pp, ok := tp.m.At(t).(points)
	if !ok {
		return points{}
	}
	return pp
}

func (tp pointmap) iterate(fn func(types.Type, points)) {
	tp.m.Iterate(func(t types.Type, i interface{}) {
		pp, _ := i.(points)
		fn(t, pp)
	})
}

func (tp pointmap) each(fn func(types.Type, point)) {
	tp.m.Iterate(func(t types.Type, i interface{}) {
		pp, _ := i.(points)
		for _, p := range pp {
			fn(t, p)
		}
	})
}

func (tp pointmap) find(fn func(types.Type, point) bool) (point, types.Type, bool) {
	var pp *point
	var tt *types.Type
	tp.each(func(key types.Type, p2 point) {
		if pp != nil {
			return
		}
		if !fn(key, p2) {
			return
		}
		pp = &p2
		tt = &key
	})
	if pp != nil && tt != nil {
		return *pp, *tt, true
	}
	return point{}, nil, false
}

func (tp pointmap) pick(t types.Type) (point, bool) {
	pp := tp.at(t)
	if len(pp) == 0 {
		return point{}, false
	}
	sort.Sort(pp)
	return pp[0], true
}

func (tp pointmap) findType(p1 pointer.Pointer) (types.Type, bool) {
	_, t, ok := tp.find(func(key types.Type, p2 point) bool {
		return p1.String() == p2.Pointer.String()
	})
	return t, ok
}

func (tp pointmap) log(log logging.Printer) {
	m := make(map[string]string)
	var k []string
	tp.each(func(t types.Type, p point) {
		m[p.String()] = t.String()
		k = append(k, p.String())
	})
	sort.Strings(k)
	for _, v := range k {
		logging.LogFunc(log)("%s => %s", v, m[v])
	}
}
