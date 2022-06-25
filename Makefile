install:
	go mod tidy

run:
	go run cmd/main.go

dev:
	air

build:
	go build -o ./bin/main cmd/main.go

run-build:
	./bin/main

test:
	go test -cover -v ./...

swagger:
	swag init -g cmd/main.go 

up:
	docker compose up -d
	docker compose logs -f

upb:
	docker compose up --build -d
	docker compose logs -f

down:
	docker compose down

updb:
	docker compose up -d redpanda
	docker compose logs -f

logs:
	docker compose logs -f

tool:
	sh ./tools/install.sh

migration:
	migrate create -seq -ext sql -dir migrations $(filter-out $@,$(MAKECMDGOALS))

migrate:
	cd ./migrations/script/ && \
	go run migrate.go

migrate-down:
	cd ./migrations/script/ && \
	go run migrate.go -action down

.PHONY: build logs
