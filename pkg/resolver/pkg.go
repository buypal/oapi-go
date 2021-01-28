package resolver

import (
	"go/build"
	"os"
	"path/filepath"
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

func isStdLibPkg(pkg *packages.Package) bool {
	_, ok := stdpkgs[pkg.PkgPath]
	return ok
}

func getGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
}

func getPkgPath(pkg *packages.Package) (string, bool) {
	if pkg.Module == nil {
		p := filepath.Join(getGoPath(), pkg.PkgPath)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return p, false
		}
		return p, true
	}
	if len(pkg.Module.Dir) == 0 {
		return "", false
	}
	return filepath.Join(pkg.Module.Dir, strings.Replace(pkg.PkgPath, pkg.Module.Path, "", 1)), true
}

type pkgInfo struct {
	*packages.Package
	dir string
}

func newPkgInfo(pkg *packages.Package) (pkgInfo, error) {
	p, ok := getPkgPath(pkg)
	if !ok {
		return pkgInfo{}, errors.Errorf("failed to determine pkg info for pkg %s", pkg.PkgPath)
	}
	pi := pkgInfo{
		Package: pkg,
		dir:     p,
	}
	return pi, nil
}
