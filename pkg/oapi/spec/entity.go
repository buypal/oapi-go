package spec

import (
	"fmt"

	"github.com/buypal/oapi-go/pkg/pointer"
)

// Entity reflets kind of component
type Entity uint

const (
	// InvalidKind designating invalid scheme
	InvalidKind Entity = iota + 1

	// ReferenceKind designating *Refable
	ReferenceKind

	// SchemaKind designating *Schema
	SchemaKind
	// ResponseKind designating *Response
	ResponseKind
	// ParameterKind designating *Parameter
	ParameterKind
	// ExampleKind designating *Example
	ExampleKind
	// RequestBodyKind designating *RequestBody
	RequestBodyKind
	// HeaderKind designating *Header
	HeaderKind
	// SecuritySchemeKind designating *SecurityScheme
	SecuritySchemeKind
	// LinkKind designating *Link
	LinkKind
	// CallbackKind designating *Callback
	CallbackKind
	// PathItemKind designating *Path
	PathItemKind
)

// Key represents to level key
func (c Entity) Key() string {
	switch c {
	case SchemaKind:
		return "schemas"
	case ResponseKind:
		return "responses"
	case ParameterKind:
		return "parameters"
	case ExampleKind:
		return "examples"
	case RequestBodyKind:
		return "requestBodies"
	case SecuritySchemeKind:
		return "securitySchemes"
	case LinkKind:
		return "links"
	case CallbackKind:
		return "callbacks"
	case PathItemKind:
		return "paths"
	case ReferenceKind:
		return "*"
	}
	return "invalid"
}

// Path returns top level path
func (c Entity) Path() string {
	return fmt.Sprintf("components.%s", c.Key())
}

// Fragment returns top level fragment
func (c Entity) Fragment() pointer.Fragment {
	f, _ := pointer.NewFragment(fmt.Sprintf("/components/%s", c.Key()))
	return f
}

// Entiter is forcing given struct to admit its purpose
type Entiter interface {
	Entity() Entity
}
