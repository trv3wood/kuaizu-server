.PHONY: run build generate tidy run-admin build-admin

run:
	go run cmd/server/main.go

build:
	go build -o bin/kuaizu-server cmd/server/main.go

run-admin:
	go run cmd/admin/main.go

build-admin:
	go build -o bin/kuaizu-admin cmd/admin/main.go

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=api/config.yaml api/service.yaml

tidy:
	go mod tidy
