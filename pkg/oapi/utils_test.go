package oapi

import (
	"testing"

	"github.com/buypal/oapi-go/pkg/container"
	"github.com/buypal/oapi-go/pkg/oapi/spec"
	"github.com/stretchr/testify/require"
)

var demoCnt = map[string]interface{}{
	"paths": map[string]interface{}{
		"/v1/demo": map[string]interface{}{
			"get": map[string]interface{}{
				"description": "Description",
				"summary":     "Summary",
			},
		},
	},
}

func TestDefaults(t *testing.T) {
	defops := map[string]spec.Operation{
		"/v1/demo": {
			Summary:     "override1",
			OperationID: "override2",
		},
	}

	cnt, _ := container.Make(demoCnt)

	err := SetPathsDefaults(cnt, defops)
	require.NoError(t, err)

	data, _ := cnt.MarshalYAML()
	require.Equal(t, string(data), `paths:
  /v1/demo:
    get:
      description: Description
      operationId: override2
      summary: Summary
`)
}
