package ocfg

import (
	"encoding/json"
	"path/filepath"

	"github.com/buypal/oapi-go/pkg/container"
	"github.com/buypal/oapi-go/pkg/oapi/spec"
)

type Config struct {
	// Path of this config
	FilePath string `json:"-"`

	// Extends file path
	ExtendsPath string `json:"extends"`

	// what version of config for oapi go should be used
	Version string `json:"version"`

	// what openapi version to produce
	OpenAPI string `json:"openapi"`

	// Format to produce
	Format string `json:"format"`

	// Directory to execute from
	Dir string `json:"dir"`

	// where to save, valid options should be stdout, stderr, file
	Output string `json:"output"`

	// where to save, valid options should be stdout, stderr, file
	// LogLevel string `json:"loglevel"`

	// go pkgs to exclude from scan
	Exclude []string `json:"exclude"`

	// Provides metadata about the API.
	// The metadata MAY be used by tooling as required.
	Info *spec.Info `json:"info"`

	// An array of Server Objects, which provide connectivity information to a target server.
	// If the servers property is not provided, or is an empty array, the default value would be a Server Object with a url value of /.
	Servers []*spec.Server `json:"servers"`

	// A declaration of which security mechanisms can be used across the API.
	// The list of values includes alternative security requirement objects that can be used.
	// Only one of the security requirement objects need to be satisfied to authorize a request.
	// Individual operations can override this definition.
	Security []map[string]spec.SecurityRequirement `json:"security"`

	// A list of tags used by the specification with additional metadata.
	// The order of the tags can be used to reflect on their order by the parsing tools.
	// Not all tags that are used by the Operation Object must be declared.
	// The tags that are not declared MAY be organized randomly or based on the tools' logic.
	// Each tag name in the list MUST be unique.
	Tags []*spec.Tag `json:"tags,omitempty"`

	// Additional external docs
	ExternalDocs *spec.ExternalDocumentation `json:"externalDocs"`

	// An element to hold various schemas for the specification.
	Components *spec.Components `json:"components,omitempty"`

	// types to override, key is pointer
	Overrides map[string]spec.Schema `json:"overrides"`

	// Operations are defaults for operations
	Operations map[string]spec.Operation `json:"operations"`
}

// ReadYAML ...
func ReadYAML(data []byte) (c Config, err error) {
	cx, err := container.ReadYAML(data)
	if err != nil {
		return Config{}, err
	}
	err = extends(cx, "")
	if err != nil {
		return
	}
	err = json.Unmarshal(cx.Bytes(), &c)
	return
}

// ReadJSON ...
func ReadJSON(data []byte) (c Config, err error) {
	cx, err := container.ReadJSON(data)
	if err != nil {
		return
	}
	err = extends(cx, "")
	if err != nil {
		return
	}
	err = json.Unmarshal(cx.Bytes(), &c)
	if err != nil {
		return
	}
	return
}

// ReadFile ...
func ReadFile(file string) (c Config, err error) {
	cx, err := container.ReadFile(file)
	if err != nil {
		return
	}
	err = extends(cx, filepath.Dir(file))
	if err != nil {
		return
	}
	err = json.Unmarshal(cx.Bytes(), &c)
	if err != nil {
		return
	}
	c.FilePath = file
	return
}

func extends(c container.Container, fp string) (err error) {
	if !c.ExistsP("extends") {
		return
	}
	p, ok := c.Path("extends").Data().(string)
	if !ok {
		return
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(fp, p)
	}
	ny, err := container.ReadFile(p)
	if err != nil {
		return
	}
	return c.Merge(ny, container.MergeDefault)
}
