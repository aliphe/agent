migrate-up:
	migrate -database "sqlite3://skipery.db" -path db/migrations up

migrate-down:
	migrate -database "sqlite3://skipery.db" -path db/migrations down
