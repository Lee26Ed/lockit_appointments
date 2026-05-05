include .envrc
# export the environment variables from .envrc file so scripts can use them
export

## run: run the cmd/api application
.PHONY: run
run:
	@echo  'Running application…'
	@go run ./cmd/api -port=3000 -env=development -dsn=${DB_DSN} \
		--smtp-host=${SMTP_HOST} --smtp-port=${SMTP_PORT} --smtp-username=${SMTP_USERNAME} --smtp-password=${SMTP_PASSWORD} --smtp-sender=${SMTP_SENDER} \
		-limiter-burst=5 \
		-limiter-rps=2 \
		-limiter-enabled=true \
		-cors-trusted-origins="http://localhost:9000"

## run/no-limiter: run the cmd/api application with rate limiter disabled
.PHONY: run/no-limiter
run/no-limiter:
	@echo  'Running application (rate limiter disabled)…'
	@go run ./cmd/api -port=3000 -env=development -dsn=${DB_DSN} \
		--smtp-host=${SMTP_HOST} --smtp-port=${SMTP_PORT} --smtp-username=${SMTP_USERNAME} --smtp-password=${SMTP_PASSWORD} --smtp-sender=${SMTP_SENDER} \
		-limiter-burst=5 \
		-limiter-rps=2 \
		-limiter-enabled=false \
		-cors-trusted-origins="http://localhost:9000"

## db/psql: connect to the database using psql (terminal)
.PHONY: db/psql
db/psql:
	@psql ${DB_DSN}
	
## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB_DSN} up

## db/migrations/down: rollback last migration
# use version=N to rollback to a specific version
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Rolling back last successful migration...'
	migrate -path ./migrations -database ${DB_DSN} down ${version}

.PHONY: db/migrations/version
db/migrations/version:
	@echo 'Current migration version...'
	migrate -path ./migrations -database ${DB_DSN} version

# force the migration version (use with caution)
.PHONY: db/migrations/force
db/migrations/force:
	@echo 'Forcing migration to ${version} version...'
	migrate -path ./migrations -database ${DB_DSN} force ${version}

## db/setup: setup the database (requires DB_NAME, DB_USER, DB_PASSWORD environment variables)
.PHONY: db/setup
db/setup:
	@echo 'Setting up database...'
	@bash ./scripts/setup_database.sh "${DB_NAME}" "${DB_USER}" "${DB_PASSWORD}" "${DB_HOST}" "${DB_PORT}"


SHELL := /bin/bash

.PHONY: demo/rate-limit
demo/rate-limit:
	@echo 'Demonstrating rate limiting...'
	@for i in {1..10}; do curl -i http://localhost:3000/healthcheck; done

.PHONY: demo/logging
demo/logging:
	@echo 'Demonstrating logging middleware...'
	@curl -i http://localhost:3000/services?page_size=3

.PHONY: demo/cors
demo/cors:
	@echo 'Demonstrating CORS middleware...'
	@go run ./cmd/web/cors/main.go

.PHONY: demo/compression
demo/compression:
	@echo 'Demonstrating gzip compression middleware...'
	@echo ""
	@echo "=> Request WITH gzip encoding (Content-Encoding will be set):"
	@curl -s -i -H "Accept-Encoding: gzip" http://localhost:3000/services?page_size=13 | head -n 20
	@echo ""
	@echo ""
	@echo "=> Request WITHOUT gzip encoding (no Content-Encoding header):"
	@curl -s -i http://localhost:3000/services?page_size=13 | head -n 20

.PHONY: demo/metrics
demo/metrics:
	@echo 'Displaying application metrics endpoint...'
	@echo ""
	@curl -s http://localhost:3000/v1/metrics | jq . 2>/dev/null || curl -s http://localhost:3000/v1/metrics

.PHONY: demo/load-test
demo/load-test:
	@echo 'Running load test against /services endpoint...'
	@echo "Note: Start the server with 'make run/no-limiter' first"
	@echo ""
	@$(HOME)/go/bin/hey -n 100 -c 10 http://localhost:3000/services 