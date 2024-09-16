FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tender-service cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /tender-service /tender-service
COPY .env .env

EXPOSE 8080

CMD ["/tender-service"]
