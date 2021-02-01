package types

import (
	"github.com/buypal/oapi-go/internal/logging"
	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/buypal/oapi-go/internal/pointer"
	"github.com/buypal/oapi-go/tag"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// Scanner will scan types and allow to resolve tem into full structures
type Scanner struct {
	Pointers pointer.Pointers
	points   pointmap
}

func NewScanner(ptrs pointer.Pointers) *Scanner {
	return &Scanner{
		Pointers: ptrs,
	}
}

// Resolve will return new pointer and scheme, new pointer might be returned in cases
// where original pointer is not fully resolved.
func (r *Scanner) Resolve(ptr pointer.Pointer) (*spec.Schema, error) {
	tp, _ := r.points.findType(ptr)
	// pp, ok := r.points.pick(tp)
	// if !ok {
	// 	return nil, errors.Errorf("failed to resolve %q", ptr.String())
	// }
	sch, err := type2schema(tp, r.points, path{}, tag.Tag{})
	return sch, err
}

func (r *Scanner) log(log logging.Printer) {
	r.points.log(log)
}

// Scan will scan types in pkgs
func (r *Scanner) Scan(pkg *packages.Package) (errs error) {
	scope := pkg.Types.Scope()

	for _, ptr := range r.Pointers {
		url := ptr.URL
		if url.Scheme != "go" {
			continue
		}
		if pkg.Types.Path() != ptr.PkgPath() {
			continue
		}
		head, ok := ptr.Fragment.Head()
		if !ok {
			continue
		}
		obj := scope.Lookup(head)
		if obj == nil {
			continue
		}
		err := collectTypes(obj, &r.points)
		if err != nil {
			return errors.Wrapf(err, "failed to register type %q", obj.Type().String())
		}
	}

	return
}
