package oapi

import (
	"encoding/json"
	"errors"

	"github.com/buypal/oapi-go/pkg/container"
	"github.com/buypal/oapi-go/pkg/oapi/spec"
)

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
	spec.OpenAPI
	c container.Container
}

// New creates new specs from container
func New(c container.Container) (x OAPI, err error) {
	x = OAPI{c: c}
	err = json.Unmarshal(c.Bytes(), &x.OpenAPI)
	return x, err
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
