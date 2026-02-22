.PHONY: run migrate migrate-down seed test test-isolation

run:
	go run ./cmd/api

migrate:
	go run ./cmd/migrate

migrate-down:
	@echo "migrate-down is not supported because no *.down.sql migrations are defined"
	@exit 1

seed:
	go run ./scripts/seed_hadiths.go
	go run ./scripts/seed_prophets.go
	go run ./scripts/seed_quran.go

test:
	go test ./... -v

test-isolation:
	go test ./tests/isolation_test.go -v
