package pkgutil

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

var stdpkgs = make(map[string]struct{})

func init() {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		panic(err)
	}

	for _, p := range pkgs {
		stdpkgs[p.PkgPath] = struct{}{}
	}
}

// Scanner will report
type Scanner interface {
	Scan(*packages.Package) error
}

// Scan visits all the packages in the import graph
func Scan(pkgs []*packages.Package, fn Scanner) (err error) {
	var errs []error

	seen := make(map[*packages.Package]bool)
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if seen[pkg] || IsStdLibPkg(pkg) || pkg == nil {
			return
		}
		seen[pkg] = true

		// First collect and check parsing errors
		for _, err := range pkg.Errors {
			errs = append(errs, errors.Wrapf(err, "failed to parse pkg"))
		}

		paths := make([]string, 0, len(pkg.Imports))
		for path := range pkg.Imports {
			paths = append(paths, path)
		}
		sort.Strings(paths) // Imports is a map, this makes visit stable
		for _, path := range paths {
			visit(pkg.Imports[path])
		}

		if err := fn.Scan(pkg); err != nil {
			errs = append(errs, err)
			return
		}
	}

	for _, pkg := range pkgs {
		visit(pkg)
	}

	if len(errs) > 0 {
		return newSingleError(errs)
	}

	return nil
}

// IsStdLibPkg reposrt if pkg is stdlib pkg
func IsStdLibPkg(pkg *packages.Package) bool {
	_, ok := stdpkgs[pkg.PkgPath]
	return ok
}

func newSingleError(errs []error) error {
	var msgs []string
	for _, m := range errs {
		msgs = append(msgs, "\t- "+m.Error())
	}
	return errors.Errorf("failed to scan: \n%s", strings.Join(msgs, "\n"))
}
