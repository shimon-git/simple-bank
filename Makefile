include ./env/.db.env

postgres:
	docker run --rm --name postgres -d \
	-v /Users/shimonyaniv/Desktop/golang/simple_bank/data:/var/lib/postgresql/data \
	-p $(POSTGRES_PORT):$(POSTGRES_PORT) -e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	postgres:alpine

rm_postgres:
	docker rm postgres -f
	docker rmi postgres:alpine

createdb:
	docker exec -it postgres createdb \
	--username=$(POSTGRES_USER) \
	--owner=$(POSTGRES_USER) \
	simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path ./db/migration \
	-database '$(CONNECTION_STRING)' \
	-verbose up

migratedown:
	migrate -path ./db/migration \
	-database '$(CONNECTION_STRING)' \
	-verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run .

.PHONY:
	server \
	postgres \
	rm_postgres \
	createdb \
	dropdb \
	migrateup \
	migratedown \
	sqlc