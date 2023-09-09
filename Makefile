include .app.env

postgres:
	docker run --rm --name postgres -d \
	-v /Users/shimonyaniv/Desktop/golang/simple_bank/data:/var/lib/postgresql/data \
	-p $(DB_PORT):$(DB_PORT) \
	-e POSTGRES_USER=$(DB_USER) \
	--network $(BANK_NETWORK) \
	-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
	postgres:alpine

app:
	docker run --rm --name app -d \
	-p 80:8080 \
	-e  GIN_MODE=release \
	--network $(BANK_NETWORK) \
	simplebank:latest

rm_postgres:
	docker rm postgres -f
	docker rmi postgres:alpine

createdb:
	docker exec -it postgres createdb \
	--username=$(DB_USER) \
	--owner=$(DB_USER) \
	$(DB_NAME)

dropdb:
	docker exec -it postgres dropdb $(DB_NAME)

migrateup:
	migrate -path ./db/migration \
	-database '$(DB_SOURCE)' \
	-verbose up

migratedown:
	migrate -path ./db/migration \
	-database '$(DB_SOURCE)' \
	-verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run .

mock:
	mockgen \
	-package mockdb \
	-destination ./db/mock/store.go \
	github.com/shimon-git/simple-bank/db/sqlc \
	Store

.PHONY:
	server \
	postgres \
	rm_postgres \
	createdb \
	dropdb \
	migrateup \
	migratedown \
	sqlc \
	mock