postgres:
	docker run --rm --name postgres -d -v /Users/shimonyaniv/Desktop/golang/simple_bank/data:/var/lib/postgresql/data -p 5432:5432 -e POSTGRES_USER=shimon -e POSTGRES_PASSWORD=ShimonTest123! postgres:alpine

rm_postgres:
	docker rm postgres -f
	docker rmi postgres:alpine

createdb:
	docker exec -it postgres createdb --username=shimon --owner=shimon simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path ./db/migration -database 'postgresql://shimon:ShimonTest123!@localhost:5432/simple_bank?sslmode=disable' -verbose up

migratedown:
	migrate -path ./db/migration -database 'postgresql://shimon:ShimonTest123!@localhost:5432/simple_bank?sslmode=disable' -verbose down


.PHONY: postgres rm_postgres createdb dropdb migrateup migratedown