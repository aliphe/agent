all: migrate-up run

run:
	go run ./cmd/term

migrate-up:
	migrate -database "sqlite3://agent.db" -path db/migrations up

migrate-down:
	migrate -database "sqlite3://agent.db" -path db/migrations down
