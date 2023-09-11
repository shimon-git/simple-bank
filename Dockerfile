# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
RUN curl -L https://github.com/eficode/wait-for/releases/download/v2.2.4/wait-for > wait-for.sh
RUN go mod tidy
COPY . .
RUN go build -o main main.go


# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/wait-for.sh .
COPY start.sh .
COPY .app.env .
COPY db/migration ./migration
RUN sh -c 'chmod +x /app/*.sh'

EXPOSE 8080

CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
