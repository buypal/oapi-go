package scan

import (
	"github.com/buypal/oapi-go/internal/oapi/spec"
	"github.com/buypal/oapi-go/internal/pointer"
	"golang.org/x/tools/go/packages"
)

// Scanner is intended for scanning packages.
// It recieves package scan it and store relevant info.
type Scanner interface {
	Scan(*packages.Package) error
}

// Resolver is sort of oposite of scanner its supposed to take info
// already resolved and return schema.
type Resolver interface {
	Resolve(ptr pointer.Pointer) (*spec.Schema, error)
}

// ScannerResolver is both Scanner and Resolver
type ScannerResolver interface {
	Scanner
	Resolver
}
