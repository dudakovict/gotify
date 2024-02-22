VERSION := 1.0.1
DB_URL=postgres://root:secret@localhost:5432/gotify?sslmode=disable

# ==============================================================================
# Building containers

gotify:
	docker buildx build \
		-f Dockerfile \
		-t gotify-api \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running a local database 

network:
	docker network create gotify-network

postgres:
	docker run --name postgres12 --network gotify-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -ti postgres12 createdb --username=root --owner=root gotify

dropdb:
	docker exec -ti postgres12 dropdb gotify

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

# ==============================================================================
# Running migrations 

migrateup:
	migrate -path platform/migrations -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path platform/migrations -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path platform/migrations -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path platform/migrations -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir platform/migrations -seq $(name)

# ==============================================================================
# Running tests within the local computer

test-only:
	CGO_ENABLED=0 go test -v -cover -short -count=1 ./...

test-race:
	CGO_ENABLED=1 go test -race -count=1 ./...

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check:
	govulncheck ./...

cover-profile:
	CGO_ENABLED=0 go test -coverprofile=p.out -count=1 ./...

cover: cover-profile
	go tool cover -html=p.out

test: test-only lint vuln-check

# ==============================================================================
# Mocking interfaces for testing

mock:
	mockgen -package mockntf -destination internal/core/notification/mock/notification.go github.com/dudakovict/gotify/internal/core/notification Storer
	mockgen -package mocksessn -destination internal/core/session/mock/session.go github.com/dudakovict/gotify/internal/core/session Storer
	mockgen -package mocksub -destination internal/core/subscription/mock/subscription.go github.com/dudakovict/gotify/internal/core/subscription Storer
	mockgen -package mocktpc -destination internal/core/topic/mock/topic.go github.com/dudakovict/gotify/internal/core/topic Storer
	mockgen -package mockusr -destination internal/core/user/mock/user.go github.com/dudakovict/gotify/internal/core/user Storer
	mockgen -package mockvrf -destination internal/core/verification/mock/verification.go github.com/dudakovict/gotify/internal/core/verification Storer
	mockgen -package mocktd -destination internal/worker/mock/distributor.go github.com/dudakovict/gotify/internal/worker TaskDistributor

# ==============================================================================
# Generating API documentation 

docs:
	swag init -g cmd/main.go --output docs/swagger

db-docs:
	dbdocs build docs/db.dbml

db-schema:
	dbml2sql --postgres -o docs/schema.sql docs/db.dbml

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all

# ==============================================================================
# Docker support

docker-down:
	docker rm -f $(shell docker ps -aq)

docker-clean:
	docker system prune -f
	docker volume prune -af

.PHONY: gotify network postgres createdb dropdb redis migrateup migrateup1 migratedown migratedown1 new_migration test-only test-race lint vuln-check cover-profile cover test mock docs db-docs db-schema deps-reset tidy deps-list deps-upgrade deps-cleancache list docker-down docker-clean