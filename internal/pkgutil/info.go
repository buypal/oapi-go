package pkgutil

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

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

// GetPkgPath reports pkg path directory.
func GetPkgPath(pkg *packages.Package) (string, error) {
	p, ok := getPkgPath(pkg)
	if !ok {
		return "", errors.Errorf("failed to determine pkg info for pkg %s", pkg.PkgPath)
	}
	return p, nil
}
