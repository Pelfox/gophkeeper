swagger:
	cd apps/server && swag init \
		-g cmd/main.go \
		-d .,../../shared/protocol \
		-o docs \
		--ot json,yaml

openapi: swagger
	cd tools/openapi-convert && go run . ../../apps/server/docs/swagger.json ../../api/openapi.yaml

client: openapi
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0 -config api/oapi-codegen.client.yaml api/openapi.yaml
