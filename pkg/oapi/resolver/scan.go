package resolver

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/buypal/oapi-go/pkg/container"
	"github.com/buypal/oapi-go/pkg/logging"
	"github.com/buypal/oapi-go/pkg/oapi"
	"github.com/buypal/oapi-go/pkg/oapi/spec"
	"github.com/buypal/oapi-go/pkg/pointer"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

const pkgMode = packages.NeedSyntax | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedName | packages.NeedModule

// Options represents options of scan.
type Options struct {
	dir      *string
	log      logging.Printer
	override map[string]spec.Schema
	defops   map[string]spec.Operation
	root     spec.OpenAPI
}

// var _ resolver.Resolver = &Resolver{}

// Option is option for Scan method
type Option func(*Options) error

// WithDir sets the directory for scan
func WithDir(dir string) Option {
	return func(r *Options) error {
		r.dir = &dir
		return nil
	}
}

// WithLog will set new logger
func WithLog(l logging.Printer) Option {
	return func(r *Options) error {
		if l == nil {
			r.log = logging.Void()
		}
		r.log = l
		return nil
	}
}

// WithOverride will add sets of overrides
func WithOverride(or map[string]spec.Schema) Option {
	return func(r *Options) error {
		r.override = or
		return nil
	}
}

// WithDefOps ...
func WithDefOps(defops map[string]spec.Operation) Option {
	return func(r *Options) error {
		r.defops = defops
		return nil
	}
}

// WithRootSchema ...
func WithRootSchema(oapi spec.OpenAPI) Option {
	return func(r *Options) error {
		r.root = oapi
		return nil
	}
}

// Scan will scan types in directory for which match given pointers, it will resolve
// pointers and types to both its relative and absolute path.
// After scan you are free to call .Resolve() allowing to get oapi scheme.
func Scan(ctx context.Context, options ...Option) (s oapi.OAPI, err error) {

	opts := &Options{
		override: make(map[string]spec.Schema),
	}

	for _, opt := range options {
		err = opt(opts)
		if err != nil {
			return
		}
	}

	var dir string
	if opts.dir != nil {
		dir = *opts.dir
	} else {
		dir, err = os.Getwd()
	}
	if err != nil {
		err = errors.Wrap(err, "os")
		return
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode:    pkgMode,
		Dir:     dir,
		Context: ctx,
	})
	if err != nil {
		err = errors.Wrap(err, "packages")
		return
	}

	pp := make(pointer.Pointers)

	// Here wi will start scanning
	// commands, comments in go code
	// starting with //oapi:
	cmdsScan := newCmdsScanner()
	err = scan(pkgs, cmdsScan)
	if err != nil {
		return
	}
	pp = pp.Merge(cmdsScan.cmds.pointers())

	// now we scan for yaml files specifications
	specsScan := newSpecScanner()
	err = scan(pkgs, specsScan)
	if err != nil {
		return
	}
	pp = pp.Merge(specsScan.pointers)

	// collect and handle types
	typesScan := newTypeScanner(pp)
	err = scan(pkgs, typesScan)
	if err != nil {
		return
	}

	var exports oapi.Exports

	var cc cmds
	for _, cmd := range cmdsScan.cmds {
		cc = append(cc, cmd...)
	}

	for _, cmd := range cc {
		switch x := cmd.(type) {
		case cmdSchema:
			if _, ok := exports.Get(x.Ptr); ok {
				err = errors.Errorf("schema of type %q already registered", x.Name)
				return
			}
			entity := oapi.Entity{
				Entity: spec.SchemaKind,
				Name:   x.Name,
			}
			exp := oapi.Pointer{
				Pointer: x.Ptr,
				Entity:  entity,
			}
			exports = append(exports, exp)
		}
	}

	c, err := specsScan.
		containers.
		Sort().
		Merge(container.MergeStrict)
	if err != nil {
		return
	}

	err = oapi.MergeWithRoot(opts.root, c)
	if err != nil {
		return
	}

	cnt, err := oapi.Resolve(c, exports, func(ptr pointer.Pointer) (e spec.Entiter, err error) {
		if ovrd, ok := opts.override[ptr.String()]; ok {
			return ovrd, nil
		}
		switch ptr.Scheme {
		case "go":
			_, e, err = typesScan.resolve(ptr)
		default:
			err = errors.New("unknown protocol to resolve")
		}
		return
	})
	if err != nil {
		return
	}

	err = oapi.SetPathsDefaults(cnt, opts.defops)
	if err != nil {
		return
	}

	return oapi.New(cnt)
}

type scanner interface {
	scan(*packages.Package) []error
}

// Visit visits all the packages in the import graph
func scan(pkgs []*packages.Package, fn scanner) (err error) {
	var errs []error

	seen := make(map[*packages.Package]bool)
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if seen[pkg] || isStdLibPkg(pkg) || pkg == nil {
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
		errs = append(errs, fn.scan(pkg)...)
	}

	for _, pkg := range pkgs {
		visit(pkg)
	}

	if len(errs) > 0 {
		return newSingleError(errs)
	}

	return nil
}

func newSingleError(errs []error) error {
	var msgs []string
	for _, m := range errs {
		msgs = append(msgs, "\t- "+m.Error())
	}
	return errors.Errorf("failed to scan: \n%s", strings.Join(msgs, "\n"))
}
