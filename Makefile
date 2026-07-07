.PHONY: run test lint docker-up docker-down build migrate-up seed

DATABASE_URL ?= postgres://blog:blog@localhost:5439/blog?sslmode=disable
JWT_SECRET ?= dev-only-change-in-production-32chars

run:
	DATABASE_URL=$(DATABASE_URL) JWT_SECRET=$(JWT_SECRET) CORS_ORIGINS=http://localhost:3001 go run ./cmd/server

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

web-dev:
	cd web && npm install && npm run dev

seed:
	curl -s -X POST localhost:8090/api/v1/auth/register \
	  -H 'Content-Type: application/json' \
	  -d '{"email":"admin@example.com","password":"secretpass"}' | jq -r .access_token > /tmp/cms_token
	curl -s -X POST localhost:8090/api/v1/admin/categories \
	  -H "Authorization: Bearer $$(cat /tmp/cms_token)" \
	  -H 'Content-Type: application/json' -d '{"name":"Engineering"}' | jq
	curl -s -X POST localhost:8090/api/v1/admin/posts \
	  -H "Authorization: Bearer $$(cat /tmp/cms_token)" \
	  -H 'Content-Type: application/json' \
	  -d '{"title":"Hello Portfolio","body":"## First post\n\nBuilt with Go + Next.js.","status":"published","excerpt":"Kickoff post"}' | jq
