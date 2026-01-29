.PHONY: run build generate tidy

run:
	go run cmd/server/main.go

build:
	go build -o bin/ cmd/server/main.go

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=api/config.yaml api/openapi.yaml

tidy:
	go mod tidy
