package api_test

import (
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIProblemDetailsContract(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "shared", "backend_api", "openapi", "openapi.yaml"))
	require.NoError(t, err)
	require.NoError(t, document.Validate(t.Context()))

	problem := document.Components.Schemas["Problem"]
	require.NotNil(t, problem)
	require.NotNil(t, problem.Value)
	require.ElementsMatch(t, []string{"type", "title", "status"}, problem.Value.Required)
	require.Contains(t, problem.Value.Properties, "detail")

	tests := []struct {
		path   string
		method string
		status string
	}{
		{path: "/todos", method: "GET", status: "500"},
		{path: "/todos", method: "POST", status: "400"},
		{path: "/todos", method: "POST", status: "500"},
		{path: "/todos/{todoId}", method: "PATCH", status: "400"},
		{path: "/todos/{todoId}", method: "DELETE", status: "400"},
		{path: "/todos/{todoId}", method: "PATCH", status: "404"},
		{path: "/todos/{todoId}", method: "PATCH", status: "409"},
		{path: "/todos/{todoId}", method: "PATCH", status: "500"},
		{path: "/todos/{todoId}", method: "DELETE", status: "404"},
		{path: "/todos/{todoId}", method: "DELETE", status: "500"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path+" "+test.status, func(t *testing.T) {
			pathItem := document.Paths.Find(test.path)
			require.NotNil(t, pathItem)
			operation := pathItem.GetOperation(test.method)
			require.NotNil(t, operation)
			response := operation.Responses.Value(test.status)
			require.NotNil(t, response)
			require.Contains(t, response.Value.Content, "application/problem+json")
		})
	}
}
