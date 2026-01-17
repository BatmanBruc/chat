FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

COPY --from=builder /go/bin/goose /usr/local/bin/goose

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["sh", "-c", "goose -dir ./migrations postgres \"host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSLMODE\" up && ./main"]