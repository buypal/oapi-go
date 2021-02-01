package specs

import (
	"github.com/buypal/oapi-go/internal/container"
	"github.com/buypal/oapi-go/internal/pkgutil"
	"github.com/buypal/oapi-go/internal/pointer"
	"golang.org/x/tools/go/packages"
)

// Scanner allows to scan go packages and search for
// yaml definitions. It will read them and store its pointers.
type Scanner struct {
	Containers container.Containers
	patterns   []string
	Pointers   pointer.Pointers
}

func NewScanner() *Scanner {
	return &Scanner{
		patterns: []string{"oapi.yaml", "oapi.yml", "oapi.json"},
		Pointers: make(pointer.Pointers),
	}
}

func (r *Scanner) Scan(pkg *packages.Package) (err error) {
	dir, err := pkgutil.GetPkgPath(pkg)
	if err != nil {
		return err
	}

	ff, err := container.ReadDir(dir, r.patterns...)
	if err != nil {
		return err
	}

	var cc container.Containers

	for _, c := range ff {
		if !c.ExistsP("openapi") {
			continue
		}

		refs, err := container.ExtractKey(c, "$ref")
		if err != nil {
			return err
		}

		for _, v := range refs {
			x, ok := v.Val.(string)
			if !ok {
				continue
			}
			p, err := pointer.Parse(x)
			if err != nil {
				return err
			}

			// this allow local references of go://#/Struct
			if p.Scheme == "go" && p.PkgPath() == "" {
				pf, _ := pointer.NewGoPointer(pkg.PkgPath, "")
				pf.Fragment = p.Fragment
				p = pf
			}

			r.Pointers[p.String()] = p

			err = c.SetP(v.Key, p.String())
			if err != nil {
				return err
			}
		}

		cc = append(cc, c)
	}

	r.Containers = append(r.Containers, cc...)

	return
}

// Merge will provide single container as result of merging.
func (r *Scanner) Merge() (c container.Container, err error) {
	return r.Containers.Sort().Merge(container.MergeStrict)
}
