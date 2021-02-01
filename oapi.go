package oapi

import (
	"context"
	"encoding/json"
	"os"

	"github.com/buypal/oapi-go/internal/container"
	"github.com/buypal/oapi-go/internal/logging"
	"github.com/buypal/oapi-go/internal/oapi"
	"github.com/buypal/oapi-go/internal/oapi/resolver"
	"github.com/buypal/oapi-go/internal/oapi/scan/cmds"
	"github.com/buypal/oapi-go/internal/oapi/scan/specs"
	"github.com/buypal/oapi-go/internal/oapi/scan/types"
	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/buypal/oapi-go/internal/pkgutil"
	"github.com/buypal/oapi-go/internal/pointer"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

const pkgMode = packages.NeedSyntax | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedName | packages.NeedModule

var order = []string{
	"openapi",
	"info",
	"components",
	"paths",
}

// OAPI specification wraps container and
// parsed specification together, parsed
// specification is used to validate & enforce
// proper specification format, whereas container
// is rather manipulation tool for building openapi spec.
type OAPI struct {
	o spec.OpenAPI
	c container.Container
}

// newOAPI creates new specs from container
func newOAPI(c container.Container) (x OAPI, err error) {
	x = OAPI{c: c}
	err = json.Unmarshal(c.Bytes(), &x.o.OpenAPI)
	return x, err
}

// Options represents options of scan.
type Options struct {
	dir      *string
	log      logging.Printer
	override map[string]spec.Schema
	defops   map[string]spec.Operation
	root     spec.OpenAPI
}

func (opts *Options) path() (dir string, err error) {
	if opts.dir != nil {
		dir = *opts.dir
	} else {
		dir, err = os.Getwd()
	}
	if err != nil {
		err = errors.Wrap(err, "os")
		return
	}
	return
}

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
func Scan(ctx context.Context, options ...Option) (s OAPI, err error) {
	opts := &Options{
		override: make(map[string]spec.Schema),
	}

	for _, opt := range options {
		err = opt(opts)
		if err != nil {
			return
		}
	}

	dir, err := opts.path()
	if err != nil {
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

	// Here wi will start scanning commands, comments in go code
	cmdsScanner := cmds.NewScanner()
	err = pkgutil.Scan(pkgs, cmdsScanner)
	if err != nil {
		return
	}

	exports, err := cmdsScanner.ExportedComponents()
	if err != nil {
		return
	}

	// Now we scan for yaml files specifications
	specsScanner := specs.NewScanner()
	err = pkgutil.Scan(pkgs, specsScanner)
	if err != nil {
		return
	}

	c, err := specsScanner.Merge()
	if err != nil {
		return
	}

	pp := make(pointer.Pointers)
	pp = pp.Merge(cmdsScanner.Commands.Pointers())
	pp = pp.Merge(specsScanner.Pointers)

	// collect and handle types
	tps := types.NewScanner(pp)
	err = pkgutil.Scan(pkgs, tps)
	if err != nil {
		return
	}

	err = oapi.MergeWithRoot(opts.root, c)
	if err != nil {
		return
	}

	cnt, err := resolver.Resolve(c, exports, func(ptr pointer.Pointer) (e spec.Entiter, err error) {
		if ovrd, ok := opts.override[ptr.String()]; ok {
			return ovrd, nil
		}
		switch ptr.Scheme {
		case "go":
			e, err = tps.Resolve(ptr)
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

	return newOAPI(cnt)
}

// Format will format given specs into given format
func Format(f string, oapi OAPI) (data []byte, err error) {
	sorter := container.SortMapMarhsaler(order)
	cont := container.NewSortMarshaller(oapi.c, sorter)
	switch f {
	case "yaml", "yml":
		return cont.MarshalYAML()
	case "json":
		return cont.MarshalJSON()
	case "json:pretty":
		return cont.MarshalIndentJSON("", "  ")
	case "go":
		data, err = cont.MarshalJSON()
		if err != nil {
			return nil, err
		}
		data, err = produceGoFile(tpldata{
			PkgName:     "main",
			OpenAPIJson: string(data),
		})
		return
	default:
		return nil, errors.New("unknown format")
	}
}
