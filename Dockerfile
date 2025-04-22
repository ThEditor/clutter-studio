FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

COPY . ./

RUN sqlc generate

RUN go build -o clutter-studio ./cmd/app.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/clutter-studio ./

EXPOSE 8081

ENV PORT=8081

CMD ["./clutter-studio"]
