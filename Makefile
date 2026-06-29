.PHONY: dev swagger test

dev:
	go run ./cmd/server

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/main.go -o docs --parseInternal
	cd web && npm run generate:api

test:
	go test ./...
	cd web && npm run build
