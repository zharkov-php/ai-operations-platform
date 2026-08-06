.PHONY: install dev-api dev-web dev-mobile generate-api-client check-api-client fmt lint test test-api test-integration test-web test-mobile test-e2e build build-api build-api-client build-web build-mobile compose-config demo demo-reset

install:
	npm ci

dev-api:
	cd apps/api && go run ./cmd/api

dev-web:
	npm run dev --workspace web

dev-mobile:
	npm run start --workspace mobile

generate-api-client:
	npm run generate --workspace @ai-operations/api-client

check-api-client: generate-api-client
	git diff --exit-code -- packages/api-client/src/generated.ts

fmt:
	cd apps/api && gofmt -w $$(find . -name '*.go')

lint:
	cd apps/api && go vet ./...
	npm run lint

test: test-api test-web test-mobile

test-api:
	cd apps/api && go test ./...

test-integration:
	cd apps/api && TEST_DATABASE_URL="$${DATABASE_URL}" go test ./internal/portfolio ./internal/pricing ./internal/ingestion ./internal/analytics ./internal/recommendation ./internal/evaluation ./internal/localmodel ./internal/experiment ./internal/budget ./internal/notification -run Postgres

test-web:
	npm run test --workspace web

test-mobile:
	npm run test --workspace mobile

test-e2e:
	npm run test:e2e --workspace web

build: build-api build-api-client build-web build-mobile

build-api:
	cd apps/api && go build ./...

build-api-client:
	npm run typecheck --workspace @ai-operations/api-client

build-web:
	npm run build --workspace web

build-mobile:
	npm run typecheck --workspace mobile

compose-config:
	docker compose config

demo:
	docker compose up -d postgres redis
	cd apps/api && DATABASE_URL='postgres://ai_operations:local-development-only@localhost:5432/ai_operations?sslmode=disable' go run ./cmd/migrate
	cd apps/api && DATABASE_URL='postgres://ai_operations:local-development-only@localhost:5432/ai_operations?sslmode=disable' SEED_OWNER_PASSWORD='demo-owner-password' go run ./cmd/seed-auth
	cd apps/api && DATABASE_URL='postgres://ai_operations:local-development-only@localhost:5432/ai_operations?sslmode=disable' go run ./cmd/seed-demo
	docker compose up -d --build api web
	@echo 'Login: owner@example.test / demo-owner-password'
	@echo 'Web: http://localhost:3000'
	@echo 'API documentation: http://localhost:3000/docs/ingestion-api'

demo-reset:
	docker compose down --volumes
