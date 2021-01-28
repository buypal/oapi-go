package resolver

import (
	"github.com/buypal/oapi-go/pkg/container"
	"github.com/buypal/oapi-go/pkg/pointer"
	"golang.org/x/tools/go/packages"
)

// Scanner
// -----------------

type specScanner struct {
	containers container.Containers
	patterns   []string
	pointers   pointer.Pointers
}

func newSpecScanner() *specScanner {
	return &specScanner{
		patterns: []string{"oapi.yaml", "oapi.yml", "oapi.json"},
		pointers: make(pointer.Pointers),
	}
}

func (r *specScanner) scan(pkg *packages.Package) (errs []error) {
	info, err := newPkgInfo(pkg)
	if err != nil {
		return []error{err}
	}

	ff, err := container.ReadDir(info.dir, r.patterns...)
	if err != nil {
		return []error{err}
	}

	var cc container.Containers

	for _, c := range ff {
		if !c.ExistsP("openapi") {
			continue
		}

		refs, err := container.ExtractKey(c, "$ref")
		if err != nil {
			return
		}

		for _, v := range refs {
			x, ok := v.Val.(string)
			if !ok {
				continue
			}
			p, err := pointer.Parse(x)
			if err != nil {
				return []error{err}
			}

			// this allow local references of go://#/Struct
			if p.Scheme == "go" && p.PkgPath() == "" {
				pf, _ := pointer.NewGoPointer(pkg.PkgPath, "")
				pf.Fragment = p.Fragment
				p = pf
			}

			r.pointers[p.String()] = p

			_, err = c.SetP(v.Key, p.String())
			if err != nil {
				return []error{err}
			}
		}

		cc = append(cc, c)
	}

	r.containers = append(r.containers, cc...)

	return
}
