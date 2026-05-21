// This is a simple tool that can convert Swagger JSON definitions file into
// OpenAPI-compatible one. This tool is used to generate HTTP client.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 3 {
		log.Fatal("usage: openapi-convert <swagger.json> <openapi.yaml>")
	}

	input := os.Args[1]
	output := os.Args[2]

	inputData, err := os.ReadFile(input)
	if err != nil {
		log.Fatalf("failed to read swagger file: %v", err)
	}

	var docV2 openapi2.T
	if err := json.Unmarshal(inputData, &docV2); err != nil {
		log.Fatalf("failed to parse swagger file: %v", err)
	}

	docV3, err := openapi2conv.ToV3(&docV2)
	if err != nil {
		log.Fatalf("failed to convert Swagger to OpenAPI: %v", err)
	}

	annotateProtocolTypes(docV3)

	if err := docV3.Validate(context.Background()); err != nil {
		log.Fatalf("failed to validate OpenAPI: %v", err)
	}

	data, err := yaml.Marshal(docV3)
	if err != nil {
		log.Fatalf("failed to marshal OpenAPI schema into yaml file: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	if err := os.WriteFile(output, data, 0o644); err != nil {
		log.Fatalf("failed to write OpenAPI schema into yaml file: %v", err)
	}

	fmt.Printf("Successfully wrote Swagger schema into OpenAPI: %s\n", output)
}

// annotateProtocolTypes adds x-go-type and x-go-type-import extensions into
// OpenAPI schema, so client generator won't write types from scratch, rather
// it would import them from a shared protocol folder.
func annotateProtocolTypes(doc *openapi3.T) {
	for name, schemaRef := range doc.Components.Schemas {
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		typeName, ok := strings.CutPrefix(name, "protocol.")
		if !ok {
			continue
		}

		if schemaRef.Value.Extensions == nil {
			schemaRef.Value.Extensions = make(map[string]any)
		}

		schemaRef.Value.Extensions["x-go-type"] = "protocol." + typeName
		schemaRef.Value.Extensions["x-go-type-import"] = map[string]any{
			"path": "github.com/Pelfox/gophkeeper/shared/protocol",
			"name": "protocol",
		}
	}
}
