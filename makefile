## run: run the cmd/api application
.PHONY: run
run:
	@echo  'Running application…'
	@go run ./cmd/api -port=3000 -env=production -dsn=postgres://lockit_appointments:one2enter@localhost/lockit_appointments?sslmode=disable
