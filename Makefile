.PHONY: run migrate migrate-down seed test test-isolation

run:
	go run ./cmd/api

migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

seed:
	go run ./scripts/seed_hadiths.go
	go run ./scripts/seed_prophets.go
	go run ./scripts/seed_quran.go

test:
	go test ./... -v

test-isolation:
	go test ./tests/isolation_test.go -v
