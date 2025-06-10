# Dockerfile
FROM golang:1.22.0-alpine AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=build /app/app .
COPY .env .env

EXPOSE 8080

CMD ["./app"]
